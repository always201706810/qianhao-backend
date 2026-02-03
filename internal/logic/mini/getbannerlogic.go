package mini

import (
	"context"

	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBannerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBannerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBannerLogic {
	return &GetBannerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBannerLogic) GetBanner() (resp *types.BannerRes, err error) {
	// 目前先写死，后期有需要可以建表存数据库
	return &types.BannerRes{
		Image:    "https://via.placeholder.com/600x300/2c3e50/ffffff?text=VIP+Number", // 替换成你真实的图片URL
		Title:    "靓号特惠季",
		Subtitle: "精选豹子号，限时抢购",
	}, nil
}
