package system

import (
	"context"
	"errors"
	"time"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddUserLogic {
	return &AddUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddUserLogic) AddUser(req *types.AddUserReq) (resp *types.LoginRes, err error) {
	var count int64
	l.svcCtx.Db.Model(&model.SysUser{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	if req.Role == "district_admin" && req.DistrictId == 0 {
		return nil, errors.New("区县管理员必须选择归属区县")
	}

	// --- 修复: 构建 DistrictId 指针 ---
	var distIdPtr *int
	if req.DistrictId != 0 {
		val := req.DistrictId
		distIdPtr = &val
	}

	user := model.SysUser{
		Username: req.Username,
		Password: req.Password,
		// ✅ 修改这里：把前端传来的 NickName 存入数据库的 RealName
		RealName:   req.NickName,
		Role:       req.Role,
		DistrictId: distIdPtr,
		Status:     1,          // 默认启用
		CreateTime: time.Now(), // 记得加上时间
	}

	if err := l.svcCtx.Db.Create(&user).Error; err != nil {
		return nil, errors.New("创建失败")
	}

	return &types.LoginRes{Token: "create_success"}, nil
}
