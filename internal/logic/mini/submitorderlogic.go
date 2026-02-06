package mini

import (
	"context"
	"errors"
	"fmt"
	"time"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type SubmitOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitOrderLogic {
	return &SubmitOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitOrderLogic) SubmitOrder(req *types.SubmitOrderReq) (resp *types.SubmitOrderRes, err error) {
	// 1. 基础校验
	if len(req.SelectedNumbers) == 0 {
		return nil, errors.New("请至少选择一个号码")
	}
	if req.Openid == "" {
		return nil, errors.New("用户信息异常(OpenID缺失)")
	}
	if req.UserInfo.DistrictId <= 0 {
		return nil, errors.New("请选择业务办理区县")
	}

	// 校验区县是否存在
	var districtCount int64
	l.svcCtx.Db.Model(&model.SysDistrict{}).Where("id = ?", req.UserInfo.DistrictId).Count(&districtCount)
	if districtCount == 0 {
		return nil, errors.New("选择的区县无效或不存在")
	}

	// 新增：每日限购检查 (限制 5 单)
	// 计算逻辑：已经提交的订单数 + 加上本次准备提交的 > 5 则报错
	var todayCount int64
	// 获取今天的 00:00:00
	todayStart := time.Now().Format("2006-01-02") + " 00:00:00"

	l.svcCtx.Db.Model(&model.BusinessOrder{}).
		Where("openid = ? AND apply_time >= ?", req.Openid, todayStart).
		Count(&todayCount)

	if todayCount+int64(len(req.SelectedNumbers)) > 5 {
		msg := fmt.Sprintf("您今日已发起 %d 次预约，每日限额 5 次，本次最多还可预约 %d 次",
			todayCount, 5-todayCount)
		return nil, errors.New(msg)
	}

	expireTime := time.Now().Add(48 * time.Hour)
	displayOrderId := fmt.Sprintf("ORD-%s", time.Now().Format("20060102150405"))

	// 2. 开启事务
	err = l.svcCtx.Db.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.SelectedNumbers {
			// ⚡ 乐观锁抢号
			result := tx.Model(&model.PhonePool{}).
				Where("id = ? AND status = 0", item.Id).
				Update("status", 1)

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected == 0 {
				var count int64
				tx.Model(&model.PhonePool{}).Where("id = ?", item.Id).Count(&count)
				if count == 0 {
					return fmt.Errorf("号码 %s 不存在或已下架", item.FullNumber)
				}
				return fmt.Errorf("手慢了！号码 %s 刚刚已被抢走", item.FullNumber)
			}

			// 📝 创建订单
			newOrder := model.BusinessOrder{
				PhoneId:         item.Id,
				CustomerName:    req.UserInfo.Name,
				CustomerPhone:   req.UserInfo.Phone,
				CustomerAddress: req.UserInfo.Address,
				DistrictId:      req.UserInfo.DistrictId,
				Status:          1,
				Openid:          req.Openid,
				ApplyTime:       time.Now(),
				ExpireTime:      expireTime,
			}

			if req.UserInfo.Remark != "" {
				newOrder.AdminRemark = "用户备注: " + req.UserInfo.Remark
			}

			if err := tx.Create(&newOrder).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &types.SubmitOrderRes{
		OrderId:    displayOrderId,
		Status:     "success",
		ExpireTime: expireTime.Format("2006-01-02 15:04:05"),
	}, nil
}
