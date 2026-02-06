// package manage
//
// import (
//
//	"context"
//	"errors"
//	"qianhao-backend/internal/model"
//	"qianhao-backend/internal/svc"
//	"qianhao-backend/internal/types"
//
//	"github.com/zeromicro/go-zero/core/logx"
//	"gorm.io/gorm"
//
// )
//
//	type UpdatePhoneLogic struct {
//		logx.Logger
//		ctx    context.Context
//		svcCtx *svc.ServiceContext
//	}
//
//	func NewUpdatePhoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePhoneLogic {
//		return &UpdatePhoneLogic{
//			Logger: logx.WithContext(ctx),
//			ctx:    ctx,
//			svcCtx: svcCtx,
//		}
//	}
//
//	func (l *UpdatePhoneLogic) UpdatePhone(req *types.UpdatePhoneReq) (resp *types.LoginRes, err error) {
//		// 开启事务，因为可能要同时改号码信息和订单信息
//		err = l.svcCtx.Db.Transaction(func(tx *gorm.DB) error {
//			var phone model.PhonePool
//			if err := tx.First(&phone, req.Id).Error; err != nil {
//				return errors.New("号码不存在")
//			}
//
//			// 1. 更新号码基础信息
//			phone.PhoneNumber = req.PhoneNumber
//			phone.Category = req.Category
//			phone.Grade = req.Grade
//			if err := tx.Save(&phone).Error; err != nil {
//				return err
//			}
//
//			// 2. 如果传了 district_id 且号码处于“锁定/待沟通(1)”状态，说明要改单
//			if req.DistrictId > 0 && phone.Status == 1 {
//				// 找到关联的待沟通订单
//				var order model.BusinessOrder
//				if err := tx.Where("phone_id = ? AND status = 1", phone.Id).First(&order).Error; err == nil {
//					// 更新订单的区县
//					order.DistrictId = req.DistrictId
//					if err := tx.Save(&order).Error; err != nil {
//						return err
//					}
//				}
//			}
//			return nil
//		})
//
//		if err != nil {
//			return nil, err
//		}
//
//		return &types.LoginRes{Token: "update_success"}, nil
//	}
package manage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"
	"qianhao-backend/internal/utils" // 引入日志工具

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
	// 开启事务
	err = l.svcCtx.Db.Transaction(func(tx *gorm.DB) error {
		var phone model.PhonePool
		if err := tx.First(&phone, req.Id).Error; err != nil {
			return errors.New("号码不存在")
		}

		// 1. 更新号码基础信息
		if req.PhoneNumber != "" {
			phone.PhoneNumber = req.PhoneNumber
		}
		if req.Category != "" {
			phone.Category = req.Category
		}
		if req.Grade > 0 {
			phone.Grade = req.Grade
		}
		// 新增：修改价格 (只要传了就改)
		// 注意：如果业务允许改为0，这里需要判断是否为 nil 或者特殊逻辑，
		// 简单起见，假设前端传了值我们就更新
		if req.Price >= 0 {
			phone.Price = req.Price
		}

		if err := tx.Save(&phone).Error; err != nil {
			return err
		}

		// 2. 如果传了 district_id 且号码处于“锁定/待沟通(1)”状态，说明要改单
		if req.DistrictId >= 0 && phone.Status == 1 {
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

	// ==========================================
	// 3. 记录操作日志
	// ==========================================
	// 获取当前操作人ID
	var userId int
	if uidNumber, ok := l.ctx.Value("userId").(json.Number); ok {
		uidInt64, _ := uidNumber.Int64()
		userId = int(uidInt64)
	}

	// 记录日志
	utils.AddLog(l.svcCtx, userId, "Admin", "编辑号码", fmt.Sprintf("ID:%d, 号码:%s", req.Id, req.PhoneNumber), "")

	return &types.LoginRes{Token: "update_success"}, nil
}
