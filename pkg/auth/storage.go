package auth

import (
	"context"
	"time"
)

// 设备类型.
type DeviceType string

const (
	DeviceUnknown DeviceType = "unknown" // unknown设备无限制登录
	DeviceWeChat  DeviceType = "wechat"  // 其他设备只能同时登录一个
	DeviceWeb     DeviceType = "web"
	DeviceiOS     DeviceType = "ios"
	DeviceAndroid DeviceType = "android"
	DeviceHarmony DeviceType = "harmony"
)

type DeviceInfo struct {
	LoginTime  time.Time  `json:"login_time"`
	DeviceType DeviceType `json:"device_type"`
}

type UserDevice struct {
	Tokens  []string     `json:"tokens"` // 存储所有token，与devices数组一一对应
	Devices []DeviceInfo `json:"devices"`
	UserID  int64        `json:"user_id"`
}

type IStorage interface {
	// Token相关操作
	SetToken(ctx context.Context, token string, userID int64, deviceType DeviceType, expiration time.Duration) error
	GetUserIDByToken(ctx context.Context, token string) (int64, error)
	GetDeviceTypeByToken(ctx context.Context, token string) (DeviceType, error)
	DeleteToken(ctx context.Context, token string) error

	// 用户登录信息操作 - 改为按用户ID和设备类型存储
	SetLoginInfo(ctx context.Context, userID int64, deviceType DeviceType, info *LoginInfo, expiration time.Duration) error
	GetLoginInfo(ctx context.Context, userID int64, deviceType DeviceType) (*LoginInfo, error)
	DeleteLoginInfo(ctx context.Context, userID int64, deviceType DeviceType) error
	GetAllLoginInfos(ctx context.Context, userID int64) (map[DeviceType]*LoginInfo, error)
	GetLoginInfoByToken(ctx context.Context, token string) (*LoginInfo, error)

	// 设备管理相关操作
	AddDevice(ctx context.Context, userID int64, deviceInfo DeviceInfo, token string) error
	RemoveDevice(ctx context.Context, userID int64, token string) error
	GetUserDevices(ctx context.Context, userID int64) (*UserDevice, error)
	IsDeviceOnline(ctx context.Context, userID int64, deviceType DeviceType) (bool, error)

	// Token续期
	RenewToken(ctx context.Context, token string, expiration time.Duration) error
	RenewLoginInfo(ctx context.Context, userID int64, deviceType DeviceType, expiration time.Duration) error

	// 清除用户所有设备
	ClearAllDevices(ctx context.Context, userID int64) error
}
