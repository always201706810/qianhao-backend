package manage

import (
	"context"
	"qianhao-backend/internal/model"
	"qianhao-backend/internal/types"

	"qianhao-backend/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPhonesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPhonesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPhonesLogic {
	return &ListPhonesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPhonesLogic) ListPhones(req *types.PhoneListReq) (resp *types.PhoneListRes, err error) {
	// 定义临时结构体接收 Join 结果
	type Result struct {
		model.PhonePool
		DistrictName string `gorm:"column:district_name"`
		DistrictId   int    `gorm:"column:district_id"`
	}

	var results []Result
	var total int64

	// 构建查询：左连接 business_order (只连 status=1 待沟通的订单)，再连 sys_district
	// 注意：这里逻辑是“展示该号码当前活跃订单的区县”
	db := l.svcCtx.Db.Table("phone_pool").
		Select("phone_pool.*, sys_district.name as district_name, business_order.district_id").
		Joins("LEFT JOIN business_order ON phone_pool.id = business_order.phone_id AND business_order.status = 1").
		Joins("LEFT JOIN sys_district ON business_order.district_id = sys_district.id")

	// 筛选条件
	if req.Category != "" {
		db = db.Where("phone_pool.category = ?", req.Category)
	}
	if req.Grade > 0 {
		db = db.Where("phone_pool.grade = ?", req.Grade)
	}
	// 管理员可能想看所有状态，这里暂时不去掉 status=0 的限制？
	// 既然要管理“已被选”的号码，建议注释掉下面这行，或者允许前端传 status 查全部
	// db = db.Where("status = ?", 0) <--- 注释掉这行，以便能查出已锁定的号码

	db.Count(&total)

	offset := (req.Page - 1) * req.Size
	if err := db.Offset(offset).Limit(req.Size).Order("phone_pool.create_time desc").Scan(&results).Error; err != nil {
		return nil, err
	}

	var list []types.PhoneInfo
	for _, item := range results {
		list = append(list, types.PhoneInfo{
			Id:           item.Id,
			PhoneNumber:  item.PhoneNumber,
			Category:     item.Category,
			Grade:        item.Grade,
			Status:       item.Status,
			DistrictName: item.DistrictName, // 赋值
			DistrictId:   item.DistrictId,   // 赋值
		})
	}

	return &types.PhoneListRes{
		Total: int(total),
		List:  list,
	}, nil
}
