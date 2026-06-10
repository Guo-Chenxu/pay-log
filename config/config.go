package config

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v3"

	"github.com/Guo-Chenxu/pay-log/dal/mysql"
	"github.com/Guo-Chenxu/pay-log/dal/redis"
	"github.com/Guo-Chenxu/pay-log/pkg/auth"
	"github.com/Guo-Chenxu/pay-log/pkg/logger"
)

type ModelConfig struct {
	BaseURL string `yaml:"base_url"`
	ModelID string `yaml:"model_id"`
	APIKey  string `yaml:"api_key"`
}

type serverConfig struct {
	Port int `yaml:"port"`
}

type appConfig struct {
	Server serverConfig        `yaml:"server"`
	Mysql  mysql.MysqlConfig   `yaml:"mysql"`
	Redis  redis.RedisConfig   `yaml:"redis"`
	Logger logger.LoggerConfig `yaml:"logger"`
	Model  ModelConfig         `yaml:"model"`
	Auth   auth.AuthConfig     `yaml:"auth"`
}

var c appConfig

func Init(filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("failed to read config: %v", err))
	}
	if err = yaml.Unmarshal(data, &c); err != nil {
		panic(fmt.Sprintf("failed to parse config: %v", err))
	}
}

func GetServerConfig() serverConfig        { return c.Server }
func GetMysqlConfig() mysql.MysqlConfig    { return c.Mysql }
func GetRedisConfig() redis.RedisConfig    { return c.Redis }
func GetLoggerConfig() logger.LoggerConfig { return c.Logger }
func GetModelConfig() ModelConfig          { return c.Model }
func GetAuthConfig() auth.AuthConfig       { return c.Auth }
