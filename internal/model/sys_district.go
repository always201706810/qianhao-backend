package model

import "time"

type SysDistrict struct {
	Id         int       `gorm:"column:id;primaryKey;autoIncrement"`
	Name       string    `gorm:"column:name"`
	CreateTime time.Time `gorm:"column:create_time"`
}

func (SysDistrict) TableName() string {
	return "sys_district"
}
