package auth

import (
	"context"

	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/mojocn/base64Captcha"
	"github.com/zeromicro/go-zero/core/logx"
)

// 使用默认的内存存储 (Simple Memory Store)
// 注意：单机部署没问题。如果是多实例集群，需要用 Redis Store，这里 MVP 版本用内存即可。
var store = base64Captcha.DefaultMemStore

type GetCaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCaptchaLogic {
	return &GetCaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCaptchaLogic) GetCaptcha() (resp *types.CaptchaRes, err error) {
	// 配置验证码参数: 高度, 宽度, 长度, 干扰强度
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)

	// 生成验证码
	c := base64Captcha.NewCaptcha(driver, store)
	id, b64s, _, err := c.Generate()
	if err != nil {
		return nil, err
	}

	return &types.CaptchaRes{
		CaptchaId: id,
		Base64:    b64s,
	}, nil
}
