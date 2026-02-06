package mini

import (
	"context"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckStatusLogic {
	return &CheckStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckStatusLogic) CheckStatus(req *types.CheckStatusReq) (resp *types.CheckStatusRes, err error) {
	// 1. 查询号码当前状态
	var phone model.PhonePool
	// 使用 NumberId 查询最准
	if err := l.svcCtx.Db.First(&phone, req.NumberId).Error; err != nil {
		return nil, err // 号码不存在
	}

	isLocked := false
	lockedByMe := false

	// Status: 0-空闲, 1-锁定/待办, 2-已办
	if phone.Status > 0 {
		isLocked = true

		// 2. 如果被锁了，检查是不是我自己锁的
		// Status: 0-空闲, 1-锁定/待办, 2-已办
		if phone.Status > 0 {
			isLocked = true

			// 修改逻辑：只要是非空闲状态，都去查一下是不是我的
			var count int64
			// 查询条件：号码ID + 我的OpenID + 状态是(1或2)
			l.svcCtx.Db.Model(&model.BusinessOrder{}).
				Where("phone_id = ? AND openid = ? AND status IN ?", phone.Id, req.Openid, []int{1, 2}).
				Count(&count)

			if count > 0 {
				lockedByMe = true
			}
		}
	}

	return &types.CheckStatusRes{
		IsLocked:   isLocked,
		LockedByMe: lockedByMe,
		Price:      phone.Price, // 把价格带回去，前端可能需要校验
	}, nil
}
