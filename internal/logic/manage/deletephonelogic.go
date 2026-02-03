package manage

import (
	"context"
	"errors"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeletePhoneLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeletePhoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePhoneLogic {
	return &DeletePhoneLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeletePhoneLogic) DeletePhone(req *types.DeletePhoneReq) (resp *types.LoginRes, err error) {
	// 1. 先查询这个号码是否存在，且状态是否允许删除
	var phone model.PhonePool
	if err := l.svcCtx.Db.First(&phone, req.Id).Error; err != nil {
		return nil, errors.New("号码不存在")
	}

	// 2. 只有“可选(0)”状态的号码才能删除
	// 如果已经被锁定了(1)或者是已办理(2)，不能删，防止破坏订单数据
	if phone.Status != 0 {
		return nil, errors.New("该号码已被占用或锁定，无法删除")
	}

	// 3. 执行物理删除
	if err := l.svcCtx.Db.Delete(&phone).Error; err != nil {
		return nil, errors.New("删除失败")
	}

	// 复用 LoginRes 返回一个成功消息
	return &types.LoginRes{Token: "deleted_success"}, nil
}
