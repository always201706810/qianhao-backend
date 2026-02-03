package manage

import (
	"context"
	"errors"
	"time"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.OrderCreateReq) (resp *types.LoginRes, err error) {
	// 开启数据库事务
	err = l.svcCtx.Db.Transaction(func(tx *gorm.DB) error {
		// 1. 悲观锁查询：锁定该号码行，防止并发抢号
		// SQL: SELECT * FROM phone_pool WHERE id = ? FOR UPDATE
		var phone model.PhonePool
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&phone, req.PhoneId).Error; err != nil {
			return errors.New("号码不存在")
		}

		// 2. 检查号码状态 (0才是可选)
		if phone.Status != 0 {
			return errors.New("手慢了，该号码已被抢走！")
		}

		// 3. 更新号码状态为 1 (待沟通/锁定)
		if err := tx.Model(&phone).Update("status", 1).Error; err != nil {
			return err
		}

		// 4. 创建订单
		order := model.BusinessOrder{
			CustomerName:    req.CustomerName,
			CustomerPhone:   req.CustomerPhone,
			CustomerAddress: req.CustomerAddress,
			PhoneId:         req.PhoneId,
			DistrictId:      req.DistrictId,
			Status:          1, // 1-待沟通
			// 设置48小时后过期
			ExpireTime: time.Now().Add(48 * time.Hour),
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// 事务执行成功，返回 nil
		return nil
	})

	if err != nil {
		// 如果事务失败（包括被抢走、数据库错误），直接返回错误信息给前端
		return nil, err
	}

	// 返回成功标识 (复用了 LoginRes 结构体，这里 token 返回一个成功消息即可)
	return &types.LoginRes{Token: "order_created_success"}, nil
}
