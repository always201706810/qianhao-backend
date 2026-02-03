package model

import "time"

type SysUser struct {
	Id         int       `gorm:"column:id;primaryKey;autoIncrement"`
	Username   string    `gorm:"column:username;unique"`
	RealName   string    `gorm:"column:real_name"`
	Password   string    `gorm:"column:password"`
	Role       string    `gorm:"column:role"`
	DistrictId *int      `gorm:"column:district_id"`
	Status     int       `gorm:"column:status"`
	CreateTime time.Time `gorm:"column:create_time"`
}

func (SysUser) TableName() string {
	return "sys_user"
}
