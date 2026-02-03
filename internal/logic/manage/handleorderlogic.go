package manage

import (
	"context"
	"encoding/json" // 引入 json
	"errors"
	"time"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type HandleOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleOrderLogic {
	return &HandleOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleOrderLogic) HandleOrder(req *types.HandleOrderReq) (resp *types.LoginRes, err error) {
	// 1. 获取当前用户 (修复版)
	var userId int64

	// 辅助解析逻辑 (同上)
	parseUid := func(key string) int64 {
		val := l.ctx.Value(key)
		if val == nil {
			return 0
		}
		if v, ok := val.(json.Number); ok {
			id, _ := v.Int64()
			return id
		}
		if v, ok := val.(float64); ok {
			return int64(v)
		}
		if v, ok := val.(int); ok {
			return int64(v)
		}
		return 0
	}

	userId = parseUid("userId")
	if userId == 0 {
		userId = parseUid("uid")
	}

	if userId == 0 {
		return nil, errors.New("登录失效，无法获取用户ID")
	}

	var currentUser model.SysUser
	if err := l.svcCtx.Db.First(&currentUser, userId).Error; err != nil {
		return nil, errors.New("用户异常，请重新登录")
	}

	var transactionErr error
	transactionErr = l.svcCtx.Db.Transaction(func(tx *gorm.DB) error {
		// 2. 查订单
		var order model.BusinessOrder
		if err := tx.First(&order, req.OrderId).Error; err != nil {
			return errors.New("订单不存在")
		}

		// 3. 【核心逻辑】权限校验
		// 如果是区县管理员，必须确保订单属于该管理员的区县
		if currentUser.Role == "district_admin" {
			if currentUser.DistrictId == nil || order.DistrictId != *currentUser.DistrictId {
				return errors.New("无权操作非本区县的订单")
			}
		}

		if order.Status == req.Action {
			return nil
		}

		// 4. 查号码 (加锁)
		var phone model.PhonePool
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&phone, order.PhoneId).Error; err != nil {
			return errors.New("关联号码不存在")
		}

		// --- 状态流转逻辑 (保持不变) ---
		if req.Action == 3 { // 拒绝/释放
			if phone.Status == 1 || (phone.Status == 2 && order.Status == 2) {
				if err := tx.Model(&phone).Update("status", 0).Error; err != nil {
					return err
				}
			}
		}

		if req.Action == 2 || req.Action == 1 { // 办理 或 回退
			isMine := (phone.Status == 2 && order.Status == 2) || (phone.Status == 1 && order.Status == 1)
			isFree := phone.Status == 0

			if !isMine && !isFree {
				return errors.New("无法变更：该号码已被其他订单(用户)抢占！")
			}

			targetPhoneStatus := 1
			if req.Action == 2 {
				targetPhoneStatus = 2
			}
			if err := tx.Model(&phone).Update("status", targetPhoneStatus).Error; err != nil {
				return err
			}
		}

		// 5. 更新订单
		updates := map[string]interface{}{
			"status":      req.Action,
			"handle_time": time.Now(),
		}
		if req.DistrictId > 0 {
			// 如果是区县管理员，通常不允许把订单转给别的区县（除非业务允许），这里暂不限制，假设他只能看见自己的
			updates["district_id"] = req.DistrictId
		}

		if err := tx.Model(&order).Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})

	if transactionErr != nil {
		return nil, transactionErr
	}

	return &types.LoginRes{Token: "handle_success"}, nil
}
