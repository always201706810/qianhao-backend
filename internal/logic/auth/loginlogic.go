package auth

import (
	"context"
	"errors"
	"time"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

//var store = base64Captcha.DefaultMemStore

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginRes, err error) {

	// 1. 优先校验验证码 (Verify: id, answer, clear)
	// clear=true 表示验证后立即从内存删除，防止重放攻击
	if !store.Verify(req.CaptchaId, req.Captcha, true) {
		return nil, errors.New("验证码错误或已过期")
	}

	// 1. 查询用户
	var user model.SysUser
	// 使用 GORM 查询 username
	result := l.svcCtx.Db.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		// 为了安全，通常提示“账号或密码错误”，但按照你的需求提示“账号不存在”也可以
		return nil, errors.New("账号不存在")
	}

	// 2. 校验密码 (这里为了演示简单，使用的是明文比对。生产环境请务必使用 bcrypt 加密)
	if user.Password != req.Password {
		return nil, errors.New("密码错误")
	}

	// 3. 校验状态
	if user.Status == 0 {
		return nil, errors.New("账号已被禁用")
	}

	// 4. 生成 JWT Token
	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire

	// 处理 DistrictId 可能为 nil 的情况
	var districtId int64 = 0
	if user.DistrictId != nil {
		districtId = int64(*user.DistrictId)
	}

	token, err := l.getJwtToken(l.svcCtx.Config.Auth.AccessSecret, now, accessExpire, int64(user.Id), user.Role, districtId)
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	// 修改返回值：填入用户信息
	var distId int
	if user.DistrictId != nil {
		distId = *user.DistrictId
	}

	//插入日志 (异步)
	go func() {
		// 获取 IP (暂时写死，后面教你通过 Context 传)
		clientIp := "127.0.0.1"
		userAgent := "Unknown Browser"

		logEntry := model.SysLog{
			UserId:     user.Id,
			Username:   user.Username, // 存快照
			Action:     "用户登录",
			Ip:         clientIp,
			Ua:         userAgent,
			CreateTime: time.Now(),
		}
		l.svcCtx.Db.Create(&logEntry)
	}()

	return &types.LoginRes{
		Token:      token,
		Username:   user.Username,
		Role:       user.Role,     // 返回角色
		NickName:   user.RealName, // 注意这里是用 RealName
		DistrictId: distId,
	}, nil
}

// getJwtToken 生成 JWT 令牌
func (l *LoginLogic) getJwtToken(secretKey string, iat, seconds, userId int64, role string, districtId int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds     // 过期时间
	claims["iat"] = iat               // 签发时间
	claims["userId"] = userId         // 存入当前用户ID
	claims["role"] = role             // 存入当前用户角色
	claims["districtId"] = districtId // 存入区县ID

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}
