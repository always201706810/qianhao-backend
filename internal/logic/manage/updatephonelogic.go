package manage

import (
	"context"
	"errors"
	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdatePhoneLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatePhoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePhoneLogic {
	return &UpdatePhoneLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePhoneLogic) UpdatePhone(req *types.UpdatePhoneReq) (resp *types.LoginRes, err error) {
	// 开启事务，因为可能要同时改号码信息和订单信息
	err = l.svcCtx.Db.Transaction(func(tx *gorm.DB) error {
		var phone model.PhonePool
		if err := tx.First(&phone, req.Id).Error; err != nil {
			return errors.New("号码不存在")
		}

		// 1. 更新号码基础信息
		phone.PhoneNumber = req.PhoneNumber
		phone.Category = req.Category
		phone.Grade = req.Grade
		if err := tx.Save(&phone).Error; err != nil {
			return err
		}

		// 2. 如果传了 district_id 且号码处于“锁定/待沟通(1)”状态，说明要改单
		if req.DistrictId > 0 && phone.Status == 1 {
			// 找到关联的待沟通订单
			var order model.BusinessOrder
			if err := tx.Where("phone_id = ? AND status = 1", phone.Id).First(&order).Error; err == nil {
				// 更新订单的区县
				order.DistrictId = req.DistrictId
				if err := tx.Save(&order).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &types.LoginRes{Token: "update_success"}, nil
}
