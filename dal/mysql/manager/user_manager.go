package manager

import (
	"sync"

	"gorm.io/gorm"

	mysqldb "github.com/Guo-Chenxu/pay-log/dal/mysql"
	"github.com/Guo-Chenxu/pay-log/dal/mysql/model"
)

type UserManager struct{ db *gorm.DB }

var (
	userMgr     *UserManager
	userMgrOnce sync.Once
)

func NewUserManager() *UserManager {
	userMgrOnce.Do(func() { userMgr = &UserManager{db: mysqldb.GetDB()} })
	return userMgr
}

func (m *UserManager) GetByUsername(username string) (*model.User, error) {
	var u model.User
	err := m.db.Where("username = ? AND is_deleted = ?", username, false).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (m *UserManager) Create(u *model.User) error {
	return m.db.Create(u).Error
}

func (m *UserManager) ListAll() ([]*model.User, error) {
	var users []*model.User
	err := m.db.Where("is_deleted = ?", false).Find(&users).Error
	return users, err
}
