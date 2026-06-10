package dal

import (
	"github.com/Guo-Chenxu/pay-log/config"
	"github.com/Guo-Chenxu/pay-log/dal/mysql"
	"github.com/Guo-Chenxu/pay-log/dal/redis"
)

func Init() {
	mysql.Init(config.GetMysqlConfig())
	redis.Init(config.GetRedisConfig())
}
