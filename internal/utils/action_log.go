package utils

import (
	"time"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"

	"gorm.io/gorm"
)

// AddLog 异步记录操作日志
func AddLog(svcCtx *svc.ServiceContext, userId int, username string, action string, targetId string, ip string) {
	go func() {
		// 如果没传 IP，给个默认值
		if ip == "" {
			ip = "127.0.0.1"
		}

		logEntry := model.SysLog{
			UserId:     userId,
			Username:   username,
			Action:     action,
			TargetId:   targetId,
			Ip:         ip,
			Ua:         "System/Admin", // 后台操作默认UA
			CreateTime: time.Now(),
		}

		// 使用新的 DB 会话，防止上下文取消导致写入失败
		svcCtx.Db.Session(&gorm.Session{NewDB: true}).Create(&logEntry)
	}()
}
