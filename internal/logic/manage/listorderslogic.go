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
	// 1. 获取当前登录用户ID
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

	// 2. 查询用户信息
	var currentUser model.SysUser
	if err := l.svcCtx.Db.First(&currentUser, userId).Error; err != nil {
		return &types.OrderListRes{List: []types.OrderItem{}, Total: 0}, nil
	}

	// 3. 构建查询
	type Result struct {
		model.BusinessOrder
		PhoneNumber string `gorm:"column:phone_number"`
		Category    string `gorm:"column:category"`
	}
	var results []Result
	var total int64

	db := l.svcCtx.Db.Table("business_order").
		Select("business_order.*, phone_pool.phone_number, phone_pool.category").
		Joins("left join phone_pool on business_order.phone_id = phone_pool.id")

	// 4. 权限隔离
	if currentUser.Role == "district_admin" {
		if currentUser.DistrictId != nil {
			db = db.Where("business_order.district_id = ?", *currentUser.DistrictId)
		} else {
			return &types.OrderListRes{List: []types.OrderItem{}, Total: 0}, nil
		}
	}

	// 5. 状态筛选
	if req.Status > 0 {
		// 指定查某种状态 (1, 2, 3, 4)
		db = db.Where("business_order.status = ?", req.Status)
	} else if req.Status == -1 {
		// ✅ 新增：-1 代表查询“非待沟通”的所有历史记录 (已办理2、已拒绝3、已过期4)
		db = db.Where("business_order.status > ?", 1)
	}

	// 6. 模糊查询 (原有 Keyword)
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		db = db.Where("business_order.customer_name LIKE ? OR business_order.customer_phone LIKE ? OR phone_pool.phone_number LIKE ?", keyword, keyword, keyword)
	}

	// ✅ 7. 新增：独立字段精确/模糊查询
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
	// ✅ 修正这里：使用 apply_time
	if err := db.Offset(offset).Limit(req.Size).Order("business_order.apply_time desc").Scan(&results).Error; err != nil {
		return nil, err
	}

	var list []types.OrderItem
	for _, item := range results {
		list = append(list, types.OrderItem{
			Id:              item.Id,
			CustomerName:    item.CustomerName,
			CustomerPhone:   item.CustomerPhone,
			CustomerAddress: item.CustomerAddress,
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
