package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Guo-Chenxu/pay-log/pkg/logger"
)

type RedisStorage struct {
	client *redis.Client
}

var _ IStorage = (*RedisStorage)(nil)

func NewRedisStorage(client *redis.Client) *RedisStorage {
	return &RedisStorage{
		client: client,
	}
}

const (
	// #nosec G101
	tokenKeyPrefix = "auth:token:%s" // token -> login id:deviceType
	// #nosec G101
	infoKeyPrefix = "auth:info:%d:%s" // loginID:deviceType -> info
	// #nosec G101
	deviceKeyPrefix = "auth:devices:%d" // loginID -> devices
)

// SetToken 存储token和对应的用户ID、设备类型.
func (r *RedisStorage) SetToken(ctx context.Context, token string, userID int64, deviceType DeviceType, expiration time.Duration) error {
	tKey := fmt.Sprintf(tokenKeyPrefix, token)
	value := fmt.Sprintf("%d:%s", userID, deviceType)

	err := r.client.Set(ctx, tKey, value, expiration).Err()
	if err != nil {
		logger.CtxErrorf(ctx, "redis set token [%s]:[%s] error: %+v", tKey, value, err)
		return err
	}

	return nil
}

// GetUserIDByToken 通过token获取用户ID.
func (r *RedisStorage) GetUserIDByToken(ctx context.Context, token string) (int64, error) {
	tKey := fmt.Sprintf(tokenKeyPrefix, token)
	val, err := r.client.Get(ctx, tKey).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.CtxErrorf(ctx, "redis get [%s] error: %+v", tKey, err)
		}
		return 0, err
	}

	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		logger.CtxErrorf(ctx, "invalid token value format [%s]", val)
		return 0, fmt.Errorf("invalid token value format")
	}

	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		logger.CtxErrorf(ctx, "parse user id [%s] error: %+v", parts[0], err)
		return 0, err
	}

	return userID, nil
}

// GetDeviceTypeByToken 通过token获取设备类型.
func (r *RedisStorage) GetDeviceTypeByToken(ctx context.Context, token string) (DeviceType, error) {
	tKey := fmt.Sprintf(tokenKeyPrefix, token)
	val, err := r.client.Get(ctx, tKey).Result()
	if err != nil {
		logger.CtxErrorf(ctx, "redis get [%s] error: %+v", tKey, err)
		return "", err
	}

	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		logger.CtxErrorf(ctx, "invalid token value format [%s]", val)
		return "", fmt.Errorf("invalid token value format")
	}

	return DeviceType(parts[1]), nil
}

// DeleteToken 删除token.
func (r *RedisStorage) DeleteToken(ctx context.Context, token string) error {
	tKey := fmt.Sprintf(tokenKeyPrefix, token)

	err := r.client.Del(ctx, tKey).Err()
	if err != nil {
		logger.CtxErrorf(ctx, "redis del token [%s] error: %+v", tKey, err)
		return err
	}

	return nil
}

// SetLoginInfo 存储用户登录信息.
func (r *RedisStorage) SetLoginInfo(ctx context.Context, userID int64, deviceType DeviceType, info *LoginInfo, expiration time.Duration) error {
	iKey := fmt.Sprintf(infoKeyPrefix, userID, deviceType)
	infoByte, err := json.Marshal(info)
	if err != nil {
		logger.CtxErrorf(ctx, "json marshal [%+v] error: %+v", info, err)
		return err
	}

	err = r.client.Set(ctx, iKey, string(infoByte), expiration).Err()
	if err != nil {
		logger.CtxErrorf(ctx, "redis set info [%s]:[%s] error: %+v", iKey, string(infoByte), err)
		return err
	}

	return nil
}

// GetLoginInfo 获取用户登录信息.
func (r *RedisStorage) GetLoginInfo(ctx context.Context, userID int64, deviceType DeviceType) (*LoginInfo, error) {
	iKey := fmt.Sprintf(infoKeyPrefix, userID, deviceType)
	val, err := r.client.Get(ctx, iKey).Result()
	if err != nil {
		logger.CtxErrorf(ctx, "redis get [%s] error: %+v", iKey, err)
		return nil, err
	}

	var info LoginInfo
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		logger.CtxErrorf(ctx, "json unmarshal [%s] error: %+v", val, err)
		return nil, err
	}

	return &info, nil
}

// GetLoginInfoByToken 获取用户登录信息通过token.
func (r *RedisStorage) GetLoginInfoByToken(ctx context.Context, token string) (*LoginInfo, error) {
	// First, get the userID and deviceType from the token
	userID, err := r.GetUserIDByToken(ctx, token)
	if err != nil {
		logger.CtxErrorf(ctx, "get user id by token error: %+v", err)
		return nil, err
	}
	deviceType, err := r.GetDeviceTypeByToken(ctx, token)
	if err != nil {
		logger.CtxErrorf(ctx, "get device type by token error: %+v", err)
		return nil, err
	}

	// Then get the login info using userID:deviceType
	iKey := fmt.Sprintf(infoKeyPrefix, userID, deviceType)
	val, err := r.client.Get(ctx, iKey).Result()
	if err != nil {
		logger.CtxErrorf(ctx, "redis get [%s] error: %+v", iKey, err)
		return nil, err
	}

	var info LoginInfo
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		logger.CtxErrorf(ctx, "json unmarshal [%s] error: %+v", val, err)
		return nil, err
	}

	return &info, nil
}

// GetAllLoginInfos 获取用户所有设备类型的登录信息.
func (r *RedisStorage) GetAllLoginInfos(ctx context.Context, userID int64) (map[DeviceType]*LoginInfo, error) {
	pattern := fmt.Sprintf(infoKeyPrefix, userID, "*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		logger.CtxErrorf(ctx, "redis keys [%s] error: %+v", pattern, err)
		return nil, err
	}

	result := make(map[DeviceType]*LoginInfo)
	for _, key := range keys {
		val, err := r.client.Get(ctx, key).Result()
		if err != nil {
			logger.CtxErrorf(ctx, "redis get [%s] error: %+v", key, err)
			continue
		}

		var info LoginInfo
		if err := json.Unmarshal([]byte(val), &info); err != nil {
			logger.CtxErrorf(ctx, "json unmarshal [%s] error: %+v", val, err)
			continue
		}

		result[info.DeviceType] = &info
	}

	return result, nil
}

// DeleteLoginInfo 删除用户登录信息.
func (r *RedisStorage) DeleteLoginInfo(ctx context.Context, userID int64, deviceType DeviceType) error {
	iKey := fmt.Sprintf(infoKeyPrefix, userID, deviceType)
	err := r.client.Del(ctx, iKey).Err()
	if err != nil {
		logger.CtxErrorf(ctx, "redis del info [%s] error: %+v", iKey, err)
		return err
	}
	return nil
}

// AddDevice 添加设备到用户设备列表.
func (r *RedisStorage) AddDevice(ctx context.Context, userID int64, deviceInfo DeviceInfo, token string) error {
	dKey := fmt.Sprintf(deviceKeyPrefix, userID)

	userDevice, err := r.GetUserDevices(ctx, userID)
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	if userDevice == nil {
		userDevice = &UserDevice{
			UserID:  userID,
			Devices: []DeviceInfo{},
			Tokens:  []string{},
		}
	}

	// 检查设备类型是否已存在，如果存在则更新
	found := false
	for i, device := range userDevice.Devices {
		if device.DeviceType == deviceInfo.DeviceType {
			userDevice.Devices[i] = deviceInfo
			userDevice.Tokens[i] = token
			found = true
			break
		}
	}

	if !found {
		userDevice.Devices = append(userDevice.Devices, deviceInfo)
		userDevice.Tokens = append(userDevice.Tokens, token)
	}

	deviceByte, err := json.Marshal(userDevice)
	if err != nil {
		logger.CtxErrorf(ctx, "json marshal user device [%+v] error: %+v", userDevice, err)
		return err
	}

	err = r.client.Set(ctx, dKey, string(deviceByte), 0).Err() // 不设置过期时间
	if err != nil {
		logger.CtxErrorf(ctx, "redis set device [%s]:[%s] error: %+v", dKey, string(deviceByte), err)
		return err
	}

	return nil
}

// RemoveDevice 从用户设备列表中移除设备.
func (r *RedisStorage) RemoveDevice(ctx context.Context, userID int64, token string) error {
	dKey := fmt.Sprintf(deviceKeyPrefix, userID)

	userDevice, err := r.GetUserDevices(ctx, userID)
	if err != nil {
		return err
	}

	if userDevice == nil {
		return nil
	}

	// 查找token对应的设备索引并移除
	for i, deviceToken := range userDevice.Tokens {
		if deviceToken == token {
			userDevice.Devices = append(userDevice.Devices[:i], userDevice.Devices[i+1:]...)
			userDevice.Tokens = append(userDevice.Tokens[:i], userDevice.Tokens[i+1:]...)
			break
		}
	}

	// 删除token
	err = r.DeleteToken(ctx, token)
	if err != nil {
		logger.CtxErrorf(ctx, "delete token [%s] error: %+v", token, err)
	}

	deviceByte, err := json.Marshal(userDevice)
	if err != nil {
		logger.CtxErrorf(ctx, "json marshal user device [%+v] error: %+v", userDevice, err)
		return err
	}

	err = r.client.Set(ctx, dKey, string(deviceByte), 0).Err()
	if err != nil {
		logger.CtxErrorf(ctx, "redis set device [%s]:[%s] error: %+v", dKey, string(deviceByte), err)
		return err
	}

	return nil
}

// GetUserDevices 获取用户的所有设备.
func (r *RedisStorage) GetUserDevices(ctx context.Context, userID int64) (*UserDevice, error) {
	dKey := fmt.Sprintf(deviceKeyPrefix, userID)
	val, err := r.client.Get(ctx, dKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		logger.CtxErrorf(ctx, "redis get [%s] error: %+v", dKey, err)
		return nil, err
	}

	var userDevice UserDevice
	if err := json.Unmarshal([]byte(val), &userDevice); err != nil {
		logger.CtxErrorf(ctx, "json unmarshal [%s] error: %+v", val, err)
		return nil, err
	}

	return &userDevice, nil
}

// IsDeviceOnline 检查设备是否在线.
func (r *RedisStorage) IsDeviceOnline(ctx context.Context, userID int64, deviceType DeviceType) (bool, error) {
	// 如果是未知设备类型（带或不带时间戳），允许无限登录，所以总是返回false（即不在线，可以登录）
	if strings.HasPrefix(string(deviceType), string(DeviceUnknown)+"-") || deviceType == DeviceUnknown {
		return false, nil
	}

	userDevice, err := r.GetUserDevices(ctx, userID)
	if err != nil {
		return false, err
	}

	if userDevice == nil {
		return false, nil
	}

	for i, device := range userDevice.Devices {
		if device.DeviceType == deviceType {
			// 检查对应的token是否还存在
			if i < len(userDevice.Tokens) {
				token := userDevice.Tokens[i]
				tKey := fmt.Sprintf(tokenKeyPrefix, token)
				_, err := r.client.Exists(ctx, tKey).Result()
				return err == nil, nil
			}
			return false, nil
		}
	}

	return false, nil
}

// RenewToken 续期token.
func (r *RedisStorage) RenewToken(ctx context.Context, token string, expiration time.Duration) error {
	tKey := fmt.Sprintf(tokenKeyPrefix, token)

	err := r.client.Expire(ctx, tKey, expiration).Err()
	if err != nil {
		logger.CtxErrorf(ctx, "redis expire token [%s] error: %+v", tKey, err)
		return err
	}

	return nil
}

// RenewLoginInfo 续期登录信息.
func (r *RedisStorage) RenewLoginInfo(ctx context.Context, userID int64, deviceType DeviceType, expiration time.Duration) error {
	iKey := fmt.Sprintf(infoKeyPrefix, userID, deviceType)
	err := r.client.Expire(ctx, iKey, expiration).Err()
	if err != nil {
		logger.CtxErrorf(ctx, "redis expire info [%s] error: %+v", iKey, err)
		return err
	}
	return nil
}

// ClearAllDevices 清除用户所有设备.
func (r *RedisStorage) ClearAllDevices(ctx context.Context, userID int64) error {
	dKey := fmt.Sprintf(deviceKeyPrefix, userID)
	err := r.client.Del(ctx, dKey).Err()
	if err != nil {
		logger.CtxErrorf(ctx, "redis delete user devices [%s] error: %+v", dKey, err)
		return err
	}

	// 清除所有登录信息
	pattern := fmt.Sprintf(infoKeyPrefix, userID, "*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		logger.CtxErrorf(ctx, "redis keys [%s] error: %+v", pattern, err)
		return err
	}

	if len(keys) > 0 {
		err = r.client.Del(ctx, keys...).Err()
		if err != nil {
			logger.CtxErrorf(ctx, "redis delete login infos [%+v] error: %+v", keys, err)
			return err
		}
	}

	return nil
}
