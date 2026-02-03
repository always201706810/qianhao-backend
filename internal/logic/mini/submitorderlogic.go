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
	// ✅ 新增：校验区县ID是否合法
	if req.UserInfo.DistrictId <= 0 {
		return nil, errors.New("请选择业务办理区县")
	}

	// ✅ 新增：去数据库查一下这个 DistrictId 是否真的存在
	var districtCount int64
	l.svcCtx.Db.Model(&model.SysDistrict{}).Where("id = ?", req.UserInfo.DistrictId).Count(&districtCount)
	if districtCount == 0 {
		return nil, errors.New("选择的区县无效或不存在")
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

			// ✅ 【测试专用】在这里加延时！
			// 模拟数据库处理很慢，或者网络卡顿，强行持有锁 10秒
			//fmt.Println(">>> 模拟并发：用户抢到了锁，正在处理中 (Sleep 10s)...")
			//time.Sleep(2 * time.Second)
			// ------------------------------------------------

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
				// ✅ 修改这里：使用前端传入的 DistrictId
				DistrictId: req.UserInfo.DistrictId,
				Status:     1,
				Openid:     req.Openid,
				ApplyTime:  time.Now(),
				ExpireTime: expireTime,
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
