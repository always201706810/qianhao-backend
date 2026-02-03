package svc

import (
	"qianhao-backend/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	Db     *gorm.DB // 全局数据库实例
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 连接数据库
	db, err := gorm.Open(mysql.Open(c.DB.DataSource), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败: " + err.Error())
	}

	return &ServiceContext{
		Config: c,
		Db:     db,
	}
}
