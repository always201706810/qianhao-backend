package mini

import (
	"context"
	"time"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserStatsLogic {
	return &GetUserStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserStatsLogic) GetUserStats(req *types.UserStatsReq) (resp *types.UserStatsRes, err error) {
	// 1. 获取今日凌晨的时间点
	// 格式：2026-02-08 00:00:00
	todayStr := time.Now().Format("2006-01-02")
	location, _ := time.LoadLocation("Local")
	startTime, _ := time.ParseInLocation("2006-01-02 15:04:05", todayStr+" 00:00:00", location)

	// 2. 统计该 OpenID 今日申请的订单总数
	var count int64
	err = l.svcCtx.Db.Model(&model.BusinessOrder{}).
		Where("openid = ? AND apply_time >= ?", req.Openid, startTime).
		Count(&count).Error

	if err != nil {
		l.Errorf("统计用户预约数失败: %v", err)
		return nil, err
	}

	// 3. 返回数据
	return &types.UserStatsRes{
		TodayCount: int(count),
		Limit:      5, // 这里的 5 可以写死，也可以以后从数据库配置表里读
	}, nil
}
