package auth_test

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Guo-Chenxu/pay-log/config"
	"github.com/Guo-Chenxu/pay-log/dal/redis"
	"github.com/Guo-Chenxu/pay-log/pkg/auth"
	"github.com/Guo-Chenxu/pay-log/pkg/logger"
)

func init() {
	ctx := context.Background()
	config.Init("../../config.yaml")
	logger.Init(config.GetLoggerConfig())
	redis.Init(config.GetRedisConfig())
	auth.InitWithRedis(ctx, redis.GetClient(), config.GetAuthConfig())
}

func TestAuth(t *testing.T) {
	userID := int64(1001)
	info, err := auth.Login(t.Context(), userID, auth.DeviceWeChat)
	if err != nil {
		t.Error(err)
	}
	t.Logf("login info: %+v", info)

	info, err = auth.ValidateToken(t.Context(), info.Token)
	if err != nil {
		t.Error(err)
	}
	t.Logf("validate info: %+v", info)
}

func TestValidateTokenInvalid(t *testing.T) {
	userID := int64(1001)
	info, err := auth.LoginWithoutDevice(t.Context(), userID)
	if err != nil {
		t.Error(err)
	}
	t.Logf("login info: %+v", info)

	c := &gin.Context{}
	info, err = auth.ValidateToken(c, info.Token)
	if err != nil {
		t.Error(err)
	}
	t.Logf("validate info: %+v", info)

	id, err := auth.GetLoginID(c)
	if err != nil {
		t.Error(err)
	}
	t.Logf("login id: %d, isValid: %+v", id, id == userID)
}
