package cron

import (
	"time"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ExpireOrderJob struct {
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExpireOrderJob(svcCtx *svc.ServiceContext) *ExpireOrderJob {
	return &ExpireOrderJob{
		svcCtx: svcCtx,
		Logger: logx.WithContext(nil), // Cron 任务没有 Request Context
	}
}

// Run 核心逻辑：找出超时订单 -> 标记过期 -> 释放号码
func (j *ExpireOrderJob) Run() {
	// 定义超时时间：48小时前
	expireTime := time.Now().Add(-48 * time.Hour)

	// 测试用：如果你想立刻看到效果，可以改成 1 分钟前
	//expireTime := time.Now().Add(-2 * time.Minute)

	j.Info("开始执行订单超时扫描...")

	err := j.svcCtx.Db.Transaction(func(tx *gorm.DB) error {
		// 1. 查出所有超时且状态为 "待沟通(1)" 的订单
		var expiredOrders []model.BusinessOrder
		if err := tx.Where("status = ? AND apply_time < ?", 1, expireTime).Find(&expiredOrders).Error; err != nil {
			return err
		}

		if len(expiredOrders) == 0 {
			return nil // 没有超时订单
		}

		j.Infof("发现 %d 个超时订单，准备处理...", len(expiredOrders))

		for _, order := range expiredOrders {
			// 2. 更新订单状态为 "已过期(4)"
			if err := tx.Model(&model.BusinessOrder{}).
				Where("id = ?", order.Id).
				Update("status", 4).Error; err != nil {
				return err
			}

			// 3. 释放关联的号码 (Status 1 -> 0)
			// 注意：要确保该号码当前确实是被这个订单占用的 (Status=1)
			if err := tx.Model(&model.PhonePool{}).
				Where("id = ? AND status = ?", order.PhoneId, 1). // 只有锁定中的才释放
				Update("status", 0).Error; err != nil {
				return err
			}

			// 可选：记录一条系统操作日志 (如果需要)
		}

		return nil
	})

	if err != nil {
		j.Errorf("订单超时任务执行异常: %v", err)
	} else {
		// j.Info("订单超时扫描结束") // 日志太多可以注释掉
	}
}
