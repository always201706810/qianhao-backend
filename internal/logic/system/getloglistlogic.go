package system

import (
	"context"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLogListLogic {
	return &GetLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLogListLogic) GetLogList(req *types.LogListReq) (resp *types.LogListRes, err error) {
	var logs []model.SysLog
	var total int64

	db := l.svcCtx.Db.Model(&model.SysLog{})

	// 搜索逻辑
	if req.Username != "" {
		db = db.Where("username LIKE ?", "%"+req.Username+"%")
	}

	db.Count(&total)

	offset := (req.Page - 1) * req.Size
	// 按时间倒序
	if err := db.Offset(offset).Limit(req.Size).Order("create_time desc").Scan(&logs).Error; err != nil {
		return nil, err
	}

	var list []types.LogItem
	for _, item := range logs {
		list = append(list, types.LogItem{
			Id:         item.Id,
			Username:   item.Username,
			Action:     item.Action,
			Ip:         item.Ip,
			Ua:         item.Ua,
			CreateTime: item.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.LogListRes{
		List:  list,
		Total: int(total),
	}, nil
}