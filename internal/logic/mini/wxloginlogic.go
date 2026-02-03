package mini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WxLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWxLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WxLoginLogic {
	return &WxLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 微信返回的数据结构
type WxSessionResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func (l *WxLoginLogic) WxLogin(req *types.WxLoginReq) (resp *types.WxLoginRes, err error) {
	if req.Code == "" {
		return nil, errors.New("缺少 code 参数")
	}

	appID := l.svcCtx.Config.MiniProgram.AppID
	appSecret := l.svcCtx.Config.MiniProgram.AppSecret

	if appID == "" || appSecret == "" {
		return nil, errors.New("后端未配置 AppID 或 Secret")
	}

	// 1. 调用微信官方接口 jscode2session
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, appSecret, req.Code)

	apiResp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("微信接口调用失败: %v", err)
	}
	defer apiResp.Body.Close()

	// 2. 解析返回结果
	var wxResp WxSessionResp
	if err := json.NewDecoder(apiResp.Body).Decode(&wxResp); err != nil {
		return nil, fmt.Errorf("解析微信响应失败: %v", err)
	}

	// 3. 检查是否有错误
	if wxResp.ErrCode != 0 {
		return nil, fmt.Errorf("微信登录失败 [%d]: %s", wxResp.ErrCode, wxResp.ErrMsg)
	}

	// 4. 返回 OpenID
	// 注意：这里我们直接把 OpenID 当作 Token 返回给前端
	// 在正规逻辑中，应该生成一个自定义 Token (JWT) 并在 Redis 里映射 OpenID
	// 但在这个简单项目里，直接用 OpenID 交互是可行的
	return &types.WxLoginRes{
		Openid: wxResp.OpenID,
		Token:  wxResp.OpenID, // 简单粗暴，前端拿着这个当 Token 用
		Expire: 7200,          // 假装有效期 2 小时
	}, nil
}
