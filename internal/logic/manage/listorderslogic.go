package manage

import (
	"context"
	"encoding/json"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListOrdersLogic) ListOrders(req *types.OrderListReq) (resp *types.OrderListRes, err error) {
	// 1. 获取当前登录用户ID (保留你的原有逻辑)
	var userId int64
	parseUid := func(key string) int64 {
		val := l.ctx.Value(key)
		if val == nil {
			return 0
		}
		if v, ok := val.(json.Number); ok {
			id, _ := v.Int64()
			return id
		}
		if v, ok := val.(float64); ok {
			return int64(v)
		}
		if v, ok := val.(int); ok {
			return int64(v)
		}
		return 0
	}

	userId = parseUid("userId")
	if userId == 0 {
		userId = parseUid("uid")
	}

	if userId == 0 {
		return &types.OrderListRes{List: []types.OrderItem{}, Total: 0}, nil
	}

	// 2. 查询用户信息 (保留你的原有逻辑)
	var currentUser model.SysUser
	if err := l.svcCtx.Db.First(&currentUser, userId).Error; err != nil {
		return &types.OrderListRes{List: []types.OrderItem{}, Total: 0}, nil
	}

	// 3. 构建查询
	type Result struct {
		model.BusinessOrder
		PhoneNumber string `gorm:"column:phone_number"`
		Category    string `gorm:"column:category"`
		// 改动1：增加接收区县名称的字段
		DistrictName string `gorm:"column:district_name"`
	}
	var results []Result
	var total int64

	// 改动2：增加 Join sys_district 表，并 select 出 district_name
	db := l.svcCtx.Db.Table("business_order").
		Select("business_order.*, phone_pool.phone_number, phone_pool.category, sys_district.name as district_name").
		Joins("left join phone_pool on business_order.phone_id = phone_pool.id").
		Joins("left join sys_district on business_order.district_id = sys_district.id")

	// 4. 权限隔离 (保留你的原有逻辑)
	if currentUser.Role == "district_admin" {
		if currentUser.DistrictId != nil {
			db = db.Where("business_order.district_id = ?", *currentUser.DistrictId)
		} else {
			return &types.OrderListRes{List: []types.OrderItem{}, Total: 0}, nil
		}
	}

	// 5. 状态筛选 (保留你的原有逻辑)
	if req.Status > 0 {
		db = db.Where("business_order.status = ?", req.Status)
	} else if req.Status == -1 {
		db = db.Where("business_order.status > ?", 1)
	}

	// 6. 模糊查询 (保留你的原有逻辑)
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		db = db.Where("business_order.customer_name LIKE ? OR business_order.customer_phone LIKE ? OR phone_pool.phone_number LIKE ?", keyword, keyword, keyword)
	}

	// 7. 新增：独立字段精确/模糊查询 (保留你的原有逻辑)
	if req.PhoneNumber != "" {
		db = db.Where("phone_pool.phone_number LIKE ?", "%"+req.PhoneNumber+"%")
	}
	if req.CustomerName != "" {
		db = db.Where("business_order.customer_name LIKE ?", "%"+req.CustomerName+"%")
	}
	if req.CustomerPhone != "" {
		db = db.Where("business_order.customer_phone LIKE ?", "%"+req.CustomerPhone+"%")
	}

	db.Count(&total)

	offset := (req.Page - 1) * req.Size
	if err := db.Offset(offset).Limit(req.Size).Order("business_order.apply_time desc").Scan(&results).Error; err != nil {
		return nil, err
	}

	var list []types.OrderItem
	for _, item := range results {
		// 改动3：拼装地址，加上区县名
		// 如果 DistrictName 有值，就显示 "[区县] 详细地址"，否则只显示详细地址
		displayAddress := item.CustomerAddress
		if item.DistrictName != "" {
			displayAddress = "[" + item.DistrictName + "] " + item.CustomerAddress
		}

		list = append(list, types.OrderItem{
			Id:              item.Id,
			CustomerName:    item.CustomerName,
			CustomerPhone:   item.CustomerPhone,
			CustomerAddress: displayAddress, // 使用拼接后的地址
			PhoneNumber:     item.PhoneNumber,
			Category:        item.Category,
			Status:          item.Status,
			ApplyTime:       item.ApplyTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.OrderListRes{
		Total: int(total),
		List:  list,
	}, nil
}
