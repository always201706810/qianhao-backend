package model

import "time"

type SysLog struct {
	Id         int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserId     int       `gorm:"column:user_id"`
	Username   string    `gorm:"column:username"` // 新加的
	Action     string    `gorm:"column:action"`   // 例如："用户登录"
	TargetId   string    `gorm:"column:target_id"`
	Ip         string    `gorm:"column:ip"`
	Ua         string    `gorm:"column:ua"` // 新加的
	CreateTime time.Time `gorm:"column:create_time"`
}

func (SysLog) TableName() string {
	return "sys_log"
}
