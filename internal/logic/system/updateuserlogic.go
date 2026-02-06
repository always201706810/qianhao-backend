package system

import (
	"context"
	"errors"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UpdateUserReq) (resp *types.LoginRes, err error) {
	var user model.SysUser
	if err := l.svcCtx.Db.First(&user, req.Id).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	if req.Role == "district_admin" && req.DistrictId == 0 {
		return nil, errors.New("区县管理员必须选择归属区县")
	}

	// --- 修复: 指针赋值 ---
	var distIdPtr *int
	if req.DistrictId != 0 {
		val := req.DistrictId
		distIdPtr = &val
	} else {
		distIdPtr = nil
	}

	// 修改这里：更新 RealName
	user.RealName = req.NickName
	user.Role = req.Role
	user.DistrictId = distIdPtr

	if req.Password != "" {
		user.Password = req.Password
	}

	if err := l.svcCtx.Db.Save(&user).Error; err != nil {
		return nil, errors.New("更新失败")
	}

	return &types.LoginRes{Token: "update_success"}, nil
}
