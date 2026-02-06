package system

import (
	"context"
	"encoding/json"
	"errors"
	"qianhao-backend/internal/model"

	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteUserLogic) DeleteUser(req *types.DeleteReq) (resp *types.LoginRes, err error) {
	// 1. 获取当前登录用户的 ID (从 JWT Context 中读取)
	var currentUid int64
	val := l.ctx.Value("userId") // 或者是 "uid"，取决于你 JWT 插件设置的 Key
	if v, ok := val.(json.Number); ok {
		currentUid, _ = v.Int64()
	} else if v, ok := val.(float64); ok {
		currentUid = int64(v)
	}

	// 2. 安全检查：禁止删除自己
	if int64(req.Id) == currentUid {
		return nil, errors.New("禁止删除当前正在登录的账号")
	}

	// 3. 执行删除操作
	// 如果你的表支持软删除（有 DeletedAt 字段），GORM 会自动执行软删除
	// 如果是硬删除，执行后数据会从数据库消失
	res := l.svcCtx.Db.Model(&model.SysUser{}).Where("id = ?", req.Id).Delete(&model.SysUser{})

	if res.Error != nil {
		l.Errorf("删除用户失败: %v", res.Error)
		return nil, errors.New("数据库操作失败")
	}

	if res.RowsAffected == 0 {
		return nil, errors.New("该用户不存在或已被删除")
	}

	// 4. 返回 LoginRes (虽然名字叫 LoginRes，但由于接口定义限制，我们返回空结构体即可)
	return &types.LoginRes{}, nil
}
