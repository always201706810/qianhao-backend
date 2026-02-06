package manage

import (
	"context"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

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
	var phones []model.PhonePool
	var total int64

	// 1. 基础查询 PhonePool
	db := l.svcCtx.Db.Model(&model.PhonePool{}).Where("is_deleted = ?", 0)

	// 筛选条件
	if req.Category != "" {
		db = db.Where("category = ?", req.Category)
	}
	if req.Grade > 0 {
		db = db.Where("grade = ?", req.Grade)
	}
	// 新增：号码模糊搜索
	// 前端传过来 phone_number，这里用 LIKE 做模糊查询
	if req.PhoneNumber != "" {
		db = db.Where("phone_number LIKE ?", "%"+req.PhoneNumber+"%")
	}
	// 计算总数
	db.Count(&total)

	// 分页查询 (按创建时间倒序)
	offset := (req.Page - 1) * req.Size
	if err := db.Offset(offset).Limit(req.Size).Order("id desc").Scan(&phones).Error; err != nil {
		return nil, err
	}

	// ==========================================
	// 🚀 核心修改：从业务单填充区县信息
	// ==========================================

	// 2. 收集当前页所有号码的 ID
	var phoneIds []int
	for _, p := range phones {
		phoneIds = append(phoneIds, p.Id)
	}

	// 定义临时结构体接收查询结果
	type OrderDistrictInfo struct {
		PhoneId      int    `gorm:"column:phone_id"`
		DistrictId   int    `gorm:"column:district_id"`
		DistrictName string `gorm:"column:district_name"`
	}
	var orderDistricts []OrderDistrictInfo

	// 3. 如果有数据，去 business_order 表联查区县名
	// 逻辑：只查 "锁定(1)" 或 "已办理(2)" 的订单，因为空闲号码没有区县归属
	if len(phoneIds) > 0 {
		l.svcCtx.Db.Table("business_order").
			Select("business_order.phone_id, business_order.district_id, sys_district.name as district_name").
			Joins("LEFT JOIN sys_district ON business_order.district_id = sys_district.id").
			Where("business_order.phone_id IN ? AND business_order.status IN (1, 2)", phoneIds).
			Scan(&orderDistricts)
	}

	// 4. 转为 Map 方便后续 O(1) 查找 (Key: PhoneId)
	districtMap := make(map[int]OrderDistrictInfo)
	for _, item := range orderDistricts {
		districtMap[item.PhoneId] = item
	}

	// 5. 组装最终返回列表
	var list []types.PhoneInfo
	for _, p := range phones {
		var dId int
		var dName string

		// 如果号码不是空闲状态，尝试从 Map 里拿区县信息
		if p.Status != 0 {
			if val, ok := districtMap[p.Id]; ok {
				dId = val.DistrictId
				dName = val.DistrictName
			}
		}

		list = append(list, types.PhoneInfo{
			Id:          p.Id,
			PhoneNumber: p.PhoneNumber,
			Category:    p.Category,
			Grade:       p.Grade,
			Price:       p.Price,
			Status:      p.Status,
			// 这里填入查出来的值
			DistrictId:   dId,
			DistrictName: dName,
		})
	}

	return &types.PhoneListRes{
		Total: int(total),
		List:  list,
	}, nil
}
