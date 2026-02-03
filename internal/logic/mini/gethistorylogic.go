package mini

import (
	"context"
	"fmt"
	"strconv"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHistoryLogic {
	return &GetHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHistoryLogic) GetHistory(req *types.HistoryReq) (resp *types.HistoryRes, err error) {
	// 定义一个结构体来接收 Join 查询结果
	type OrderResult struct {
		model.BusinessOrder
		PhoneNumber string  `gorm:"column:phone_number"`
		Grade       int     `gorm:"column:grade"` // 注意：之前确认你的 Grade 是 int
		Category    string  `gorm:"column:category"`
		Price       float64 `gorm:"column:price"`
	}

	var results []OrderResult
	var total int64

	// 1. 关联查询：查询 business_order 表，同时把 phone_pool 表的信息带出来
	db := l.svcCtx.Db.Table("business_order").
		Select("business_order.*, phone_pool.phone_number, phone_pool.grade, phone_pool.category, phone_pool.price").
		Joins("left join phone_pool on business_order.phone_id = phone_pool.id").
		Where("business_order.openid = ?", req.Openid) // 只查当前用户的

	// 2. 状态筛选 (pending, processed, rejected, expired)
	if req.Status != "" {
		statusMap := map[string]int{
			"pending":   1,
			"processed": 2,
			"rejected":  3,
			"expired":   4,
		}
		if code, ok := statusMap[req.Status]; ok {
			db = db.Where("business_order.status = ?", code)
		}
	}

	// 3. 计算总数
	db.Count(&total)

	// 4. 分页取数据
	offset := (req.Page - 1) * req.PageSize
	if err := db.Offset(offset).Limit(req.PageSize).Order("business_order.apply_time desc").Scan(&results).Error; err != nil {
		return nil, err
	}

	// 5. 格式化返回
	var list []types.HistoryOrder
	for _, item := range results {
		// 转换状态文案
		statusStr := "pending"
		statusText := "待办理"
		switch item.Status {
		case 1:
			statusStr = "pending"
			statusText = "待办理"
		case 2:
			statusStr = "processed"
			statusText = "已办理"
		case 3:
			statusStr = "rejected"
			statusText = "已拒绝"
		case 4:
			statusStr = "expired"
			statusText = "已过期"
		}

		// 格式化等级 (例如 "2级")
		levelDisplay := fmt.Sprintf("%d级", item.Grade)

		// 格式化时间戳 (秒级)
		timeStr := strconv.FormatInt(item.ApplyTime.Unix(), 10)

		list = append(list, types.HistoryOrder{
			Id:         fmt.Sprintf("history_%d", item.Id),
			OrderId:    fmt.Sprintf("ORD-%s", item.ApplyTime.Format("20060102150405")),
			Number:     item.PhoneNumber,
			Type:       item.Category,
			Level:      levelDisplay,
			Status:     statusStr,
			StatusText: statusText,
			Price:      item.Price,
			Time:       timeStr,
		})
	}

	return &types.HistoryRes{
		List:  list,
		Total: int(total),
	}, nil
}
