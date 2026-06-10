package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/Guo-Chenxu/pay-log/pkg/logger"
)

type AuthConfig struct {
	TokenName string `yaml:"token_name"`
	Timeout   int64  `yaml:"timeout"`    // in seconds
	RenewTime int64  `yaml:"renew_time"` // in seconds
}

type LoginInfo struct {
	Token      string     `json:"token"`
	TokenName  string     `json:"token_name"`
	DeviceType DeviceType `json:"device_type"`
	LoginID    int64      `json:"login_id,string"`
	ExpireTime int64      `json:"expire,string"`
}

type auth struct {
	storage IStorage
	config  AuthConfig
}

type loginIDContextKey struct{}

const (
	// #nosec G101
	ginLoginIDKey = "auth_login_id" // gin context key for login id
)

var (
	a          *auth
	authOnce   sync.Once
	loginIDKey = &loginIDContextKey{}
)

func Init(ctx context.Context, storage IStorage, cfg AuthConfig) {
	authOnce.Do(func() {
		if storage == nil {
			panic("auth storage is nil")
		}
		a = &auth{
			storage: storage,
			config:  cfg,
		}
	})
}

// InitWithRedis 为向后兼容保留的初始化方法.
func InitWithRedis(ctx context.Context, redisClient *redis.Client, cfg AuthConfig) {
	if redisClient == nil {
		panic("auth redis client is nil")
	}
	storage := NewRedisStorage(redisClient)
	Init(ctx, storage, cfg)
}

// Login 用户登录，生成 token 并存储到 storage.
func Login(ctx context.Context, id int64, deviceType DeviceType) (*LoginInfo, error) {
	var err error
	token := gengerateToken()

	// 对于 unknown 设备类型，添加时间戳以允许多个会话
	finalDeviceType := deviceType
	if deviceType == DeviceUnknown {
		finalDeviceType = DeviceType(fmt.Sprintf("%s-%d", deviceType, time.Now().Unix()))
	}

	// 检查设备类型是否已在线
	isOnline, err := a.storage.IsDeviceOnline(ctx, id, finalDeviceType)
	if err != nil {
		logger.CtxErrorf(ctx, "check device online status error: %+v", err)
		return nil, err
	}

	// 如果同一设备类型已在线，先踢出旧会话
	if isOnline {
		logger.CtxInfof(ctx, "device type [%s] already online for user [%d], kicking old session", finalDeviceType, id)
		err = LogoutByDeviceType(ctx, id, finalDeviceType)
		if err != nil {
			logger.CtxErrorf(ctx, "kick device type [%s] error: %+v", finalDeviceType, err)
			return nil, err
		}
	}

	// 存储token
	err = a.storage.SetToken(ctx, token, id, finalDeviceType, time.Duration(a.config.Timeout)*time.Second)
	if err != nil {
		logger.CtxErrorf(ctx, "storage set token error: %+v", err)
		return nil, err
	}

	// 创建登录信息
	info := &LoginInfo{
		Token:      token,
		TokenName:  a.config.TokenName,
		LoginID:    id,
		DeviceType: finalDeviceType,
		ExpireTime: time.Now().Add(time.Duration(a.config.Timeout) * time.Second).Unix(),
	}

	// 存储登录信息
	err = a.storage.SetLoginInfo(ctx, id, finalDeviceType, info, time.Duration(a.config.Timeout)*time.Second)
	if err != nil {
		logger.CtxErrorf(ctx, "storage set login info error: %+v", err)
		return nil, err
	}

	// 添加设备信息
	deviceInfo := DeviceInfo{
		DeviceType: finalDeviceType,
		LoginTime:  time.Now(),
	}
	err = a.storage.AddDevice(ctx, id, deviceInfo, token)
	if err != nil {
		logger.CtxErrorf(ctx, "storage add device error: %+v", err)
		return nil, err
	}

	return info, nil
}

// LoginWithoutDevice 用户登录，不指定设备类型，使用unknown设备类型.
func LoginWithoutDevice(ctx context.Context, id int64) (*LoginInfo, error) {
	return Login(ctx, id, DeviceUnknown)
}

// getInfoByToken 通过 token 获取登录信息.
func getInfoByToken(ctx context.Context, token string) (*LoginInfo, error) {
	info, err := a.storage.GetLoginInfoByToken(ctx, token)
	if err != nil {
		logger.CtxErrorf(ctx, "storage get login info by token error: %+v", err)
		return nil, err
	}
	return info, nil
}

// ValidateToken 验证 token 是否有效并续期.
func ValidateToken(ctx context.Context, token string) (*LoginInfo, error) {
	info, err := getInfoByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	// 过期
	if info.ExpireTime < time.Now().Unix() {
		logger.CtxInfof(ctx, "token [%s] expired", token)
		return nil, errors.New("token expired")
	}
	// nolint:errcheck
	go RenewToken(ctx, token)

	// 如果是gin.Context，则将loginID存入context中
	if ginCtx, ok := ctx.(*gin.Context); ok {
		ginCtx.Set(ginLoginIDKey, info.LoginID)
	}

	return info, nil
}

// RenewToken 续期 token.
func RenewToken(ctx context.Context, token string) error {
	info, err := getInfoByToken(ctx, token)
	if err != nil {
		return err
	}

	// 过期
	if info.ExpireTime < time.Now().Unix() {
		logger.CtxInfof(ctx, "token [%s] expired", token)
		return errors.New("token expired")
	}

	if info.ExpireTime-time.Now().Unix() > a.config.RenewTime {
		return nil
	}

	// 续期token和登录信息
	err = a.storage.RenewToken(ctx, token, time.Duration(a.config.Timeout)*time.Second)
	if err != nil {
		logger.CtxErrorf(ctx, "storage renew token error: %+v", err)
		return err
	}

	// 更新登录信息的过期时间
	info.ExpireTime = time.Now().Add(time.Duration(a.config.Timeout) * time.Second).Unix()
	err = a.storage.SetLoginInfo(ctx, info.LoginID, info.DeviceType, info, time.Duration(a.config.Timeout)*time.Second)
	if err != nil {
		logger.CtxErrorf(ctx, "storage update login info error: %+v", err)
		return err
	}

	return nil
}

// AddLoginIDToContext 将 login_id 添加到上下文中.
func AddLoginIDToContext(ctx context.Context, loginID int64) context.Context {
	return context.WithValue(ctx, loginIDKey, loginID)
}

// GetLoginID 根据上下文获取 login_id.
func GetLoginID(ctx context.Context) (int64, error) {
	// 优先从 gin.Context 获取
	if ginCtx, ok := ctx.(*gin.Context); ok {
		val, exists := ginCtx.Get(ginLoginIDKey)
		if !exists {
			return 0, errors.New("not login")
		}
		loginID, ok := val.(int64)
		if !ok {
			return 0, errors.New("not login")
		}
		return loginID, nil
	}

	// 从普通 context 获取
	id, ok := ctx.Value(loginIDKey).(int64)
	if !ok {
		return 0, errors.New("not login")
	}
	return id, nil
}

// Logout 用户退出登录，通过token.
func Logout(ctx context.Context, token string) error {
	userID, err := a.storage.GetUserIDByToken(ctx, token)
	if err != nil {
		logger.CtxErrorf(ctx, "get user id by token error: %+v", err)
		return err
	}

	deviceType, err := a.storage.GetDeviceTypeByToken(ctx, token)
	if err != nil {
		logger.CtxErrorf(ctx, "get device type by token error: %+v", err)
		return err
	}

	// 删除token
	err = a.storage.DeleteToken(ctx, token)
	if err != nil {
		logger.CtxErrorf(ctx, "storage delete token error: %+v", err)
		return err
	}

	// 移除设备
	err = a.storage.RemoveDevice(ctx, userID, token)
	if err != nil {
		logger.CtxErrorf(ctx, "storage remove device error: %+v", err)
		return err
	}

	// 删除登录信息
	err = a.storage.DeleteLoginInfo(ctx, userID, deviceType)
	if err != nil {
		logger.CtxErrorf(ctx, "storage delete login info error: %+v", err)
		return err
	}

	logger.CtxInfof(ctx, "user [%d] logged out from device type [%s]", userID, deviceType)
	return nil
}

// LogoutByDeviceType 踢出指定设备类型的所有会话.
func LogoutByDeviceType(ctx context.Context, userID int64, deviceType DeviceType) error {
	userDevices, err := a.storage.GetUserDevices(ctx, userID)
	if err != nil {
		logger.CtxErrorf(ctx, "storage get user devices error: %+v", err)
		return err
	}

	if userDevices == nil {
		return errors.New("user has no devices")
	}

	// 查找指定设备类型的token并删除
	var tokensToRemove []string
	for i, device := range userDevices.Devices {
		if device.DeviceType == deviceType && i < len(userDevices.Tokens) {
			tokensToRemove = append(tokensToRemove, userDevices.Tokens[i])
		}
	}

	if len(tokensToRemove) == 0 {
		return errors.New("device type not found")
	}

	// 删除所有相关token和设备
	for _, token := range tokensToRemove {
		err = a.storage.DeleteToken(ctx, token)
		if err != nil {
			logger.CtxErrorf(ctx, "storage delete token error: %+v", err)
			continue
		}

		err = a.storage.RemoveDevice(ctx, userID, token)
		if err != nil {
			logger.CtxErrorf(ctx, "storage remove device error: %+v", err)
			continue
		}
	}

	// 删除登录信息
	err = a.storage.DeleteLoginInfo(ctx, userID, deviceType)
	if err != nil {
		logger.CtxErrorf(ctx, "storage delete login info error: %+v", err)
		return err
	}

	logger.CtxInfof(ctx, "user [%d] kicked from device type [%s]", userID, deviceType)
	return nil
}

// LogoutAll 踢出用户所有设备.
func LogoutAll(ctx context.Context, userID int64) error {
	userDevices, err := a.storage.GetUserDevices(ctx, userID)
	if err != nil {
		logger.CtxErrorf(ctx, "storage get user devices error: %+v", err)
		return err
	}

	if userDevices == nil {
		return nil
	}

	// 删除所有token
	for _, token := range userDevices.Tokens {
		err = a.storage.DeleteToken(ctx, token)
		if err != nil {
			logger.CtxErrorf(ctx, "storage delete token [%s] error: %+v", token, err)
		}
	}

	// 清除所有设备信息和登录信息
	err = a.storage.ClearAllDevices(ctx, userID)
	if err != nil {
		logger.CtxErrorf(ctx, "storage clear all devices error: %+v", err)
		return err
	}

	logger.CtxInfof(ctx, "user [%d] logged out from all devices", userID)
	return nil
}

// GetUserDevices 获取用户的所有在线设备.
func GetUserDevices(ctx context.Context, userID int64) (*UserDevice, error) {
	return a.storage.GetUserDevices(ctx, userID)
}

func gengerateToken() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
