package mysql

import (
	"fmt"
	"sync"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/Guo-Chenxu/pay-log/dal/mysql/model"
)

type MysqlConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

var (
	db   *gorm.DB
	once sync.Once
)

func Init(cfg MysqlConfig) {
	once.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
		var err error
		db, err = gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
		if err != nil {
			panic(fmt.Sprintf("failed to connect mysql: %v", err))
		}
		if err = db.AutoMigrate(
			&model.User{},
			&model.BillRecord{},
			&model.MonthSummary{},
			&model.AISummary{},
		); err != nil {
			panic(fmt.Sprintf("failed to automigrate: %v", err))
		}
	})
}

func GetDB() *gorm.DB { return db }
