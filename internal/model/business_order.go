package model

import "time"

type BusinessOrder struct {
	Id              int    `gorm:"column:id;primaryKey;autoIncrement"`
	CustomerName    string `gorm:"column:customer_name"`
	CustomerPhone   string `gorm:"column:customer_phone"`
	CustomerAddress string `gorm:"column:customer_address"`
	PhoneId         int    `gorm:"column:phone_id"`
	DistrictId      int    `gorm:"column:district_id"`
	// ✅ 新增
	Openid      string     `gorm:"column:openid"`
	AdminId     *int       `gorm:"column:admin_id"` // 指针类型，允许为 null
	Status      int        `gorm:"column:status"`
	ApplyTime   time.Time  `gorm:"column:apply_time;autoCreateTime"` // 自动生成时间
	ExpireTime  time.Time  `gorm:"column:expire_time"`
	HandleTime  *time.Time `gorm:"column:handle_time"`
	AdminRemark string     `gorm:"column:admin_remark"`
}

func (BusinessOrder) TableName() string {
	return "business_order"
}
