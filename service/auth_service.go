package service

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Guo-Chenxu/pay-log/dal/mysql/manager"
	"github.com/Guo-Chenxu/pay-log/dal/mysql/model"
	"github.com/Guo-Chenxu/pay-log/pkg/auth"
	"github.com/Guo-Chenxu/pay-log/pkg/bizerr"
	"github.com/Guo-Chenxu/pay-log/pkg/snowflake"
)

type AuthService struct {
	userMgr *manager.UserManager
}

var (
	authSvc     *AuthService
	authSvcOnce sync.Once
)

func NewAuthService() *AuthService {
	authSvcOnce.Do(func() {
		authSvc = &AuthService{userMgr: manager.NewUserManager()}
	})
	return authSvc
}

// Login authenticates the user and returns a token string.
func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.userMgr.GetByUsername(username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", bizerr.NewCustomErrorWithExtra(bizerr.Unauthorized.Code, "用户名或密码错误")
	}
	if err != nil {
		return "", err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", bizerr.NewCustomErrorWithExtra(bizerr.Unauthorized.Code, "用户名或密码错误")
	}
	// auth.Login returns (*LoginInfo, error); extract the token string.
	info, err := auth.Login(ctx, user.ID, auth.DeviceWeb)
	if err != nil {
		return "", err
	}
	return info.Token, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return auth.Logout(ctx, token)
}

func (s *AuthService) CreateUser(username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userMgr.Create(&model.User{
		ID:       snowflake.GenerateID(),
		Username: username,
		Password: string(hash),
	})
}
