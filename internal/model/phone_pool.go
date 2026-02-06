package model

import "time"

type PhonePool struct {
	Id          int    `gorm:"column:id;primaryKey;autoIncrement"`
	PhoneNumber string `gorm:"column:phone_number;unique"` // 唯一索引
	Category    string `gorm:"column:category"`
	Grade       int    `gorm:"column:grade"`
	// 新增
	Price        float64 `gorm:"column:price"`
	Status       int     `gorm:"column:status"` // 0-可选
	ImportUserId int     `gorm:"column:import_user_id"`
	IsDeleted    int     `gorm:"column:is_deleted"`
	//CreateTime   time.Time `gorm:"column:create_time"`
	//你在 phone_pool 表里定义了 create_time 字段，并且在 Go 的 model.PhonePool 结构体里定义了 CreateTime time.Time。
	//
	//当你初始化 newPhone 结构体时，没有给 CreateTime 赋值。
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime"`
}

func (PhonePool) TableName() string {
	return "phone_pool"
}
