package manage

import (
	"context"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAnalysisLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAnalysisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAnalysisLogic {
	return &GetAnalysisLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAnalysisLogic) GetAnalysis() (resp *types.AnalysisRes, err error) {
	resp = &types.AnalysisRes{}

	// 1. 统计号码池数据
	l.svcCtx.Db.Model(&model.PhonePool{}).Count(&resp.TotalPhones)
	l.svcCtx.Db.Model(&model.PhonePool{}).Where("status = ?", 2).Count(&resp.SoldPhones)

	// 2. 统计订单数据 (1-待沟通, 3-已拒绝)
	l.svcCtx.Db.Model(&model.BusinessOrder{}).Where("status = ?", 1).Count(&resp.PendingOrders)
	l.svcCtx.Db.Model(&model.BusinessOrder{}).Where("status = ?", 3).Count(&resp.RejectedOrders)

	// 3. 统计各区县数据 (稍微复杂一点的 SQL)
	// 我们需要查出所有区县，并统计它们各自的 待处理(1) 和 已办理(2) 订单数
	type StatResult struct {
		DistrictId   int
		DistrictName string
		Status       int
		Count        int64
	}
	var results []StatResult

	// 关联查询：sys_district left join business_order
	// Select: district_id, district_name, order_status, count(*)
	err = l.svcCtx.Db.Table("sys_district").
		Select("sys_district.id as district_id, sys_district.name as district_name, business_order.status, count(business_order.id) as count").
		Joins("LEFT JOIN business_order ON sys_district.id = business_order.district_id").
		Group("sys_district.id, sys_district.name, business_order.status").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// 4. 数据清洗：把数据库的扁平数据转为前端需要的结构
	// Map: DistrictID -> DistrictStat
	statMap := make(map[string]*types.DistrictStat)

	for _, row := range results {
		if _, ok := statMap[row.DistrictName]; !ok {
			statMap[row.DistrictName] = &types.DistrictStat{
				DistrictName:  row.DistrictName,
				PendingCount:  0,
				SoldCount:     0,
				RejectedCount: 0, // 初始化
			}
		}
		if row.Status == 1 {
			statMap[row.DistrictName].PendingCount = row.Count
		} else if row.Status == 2 {
			statMap[row.DistrictName].SoldCount = row.Count
		} else if row.Status == 3 {
			// 统计已拒绝 (Status=3)
			statMap[row.DistrictName].RejectedCount = row.Count
		}
	}

	// 转为切片
	for _, v := range statMap {
		resp.DistrictStats = append(resp.DistrictStats, *v)
	}

	return resp, nil
}
