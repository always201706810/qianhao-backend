package manage

import (
	"context"
	"encoding/json"
	"time"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ImportPhonesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewImportPhonesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ImportPhonesLogic {
	return &ImportPhonesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ImportPhones 全量同步：Excel 有则加，Excel 无则删(仅限空闲)，重复则忽略
func (l *ImportPhonesLogic) ImportPhones(req *types.ImportPhonesReq) (resp *types.ImportPhonesRes, err error) {
	// 1. 获取当前操作员 ID
	// 保持你之前的写法
	userIdNumber, _ := l.ctx.Value("userId").(json.Number)
	userId, _ := userIdNumber.Int64()

	// 2. 整理 Excel 数据到 Map，方便比对
	// Map Key: PhoneNumber, Value: Item
	excelMap := make(map[string]*types.ImportItem)
	for _, item := range req.List {
		if item.PhoneNumber != "" {
			// 这里做一个临时变量存 item，防止循环指针问题
			temp := item
			excelMap[item.PhoneNumber] = &temp
		}
	}

	// 3. 查询数据库中【当前所有】的号码
	// 只查 phone_number 和 status 字段，减少内存消耗
	// 假设我们只比对未被软删除的数据 (IsDeleted = 0)
	type PhoneSimple struct {
		PhoneNumber string
		Status      int
	}
	var dbPhones []PhoneSimple
	// 注意：这里要查全量，不要 limit
	if err := l.svcCtx.Db.Model(&model.PhonePool{}).
		Where("is_deleted = ?", 0).
		Select("phone_number, status").
		Scan(&dbPhones).Error; err != nil {
		return nil, err
	}

	// 构造 DB Map
	dbMap := make(map[string]int)
	for _, p := range dbPhones {
		dbMap[p.PhoneNumber] = p.Status
	}

	// 4. 核心比对：分类
	var toInsert []model.PhonePool
	var toDeletePhones []string

	// A. 遍历 Excel，找出【新增】(DB中没有的)
	for phone, item := range excelMap {
		if _, exists := dbMap[phone]; !exists {
			// 数据库没有 -> 插入
			toInsert = append(toInsert, model.PhonePool{
				PhoneNumber:  item.PhoneNumber,
				Category:     item.Category,
				Grade:        item.Grade,
				Status:       0, // 默认空闲
				ImportUserId: int(userId),
				IsDeleted:    0,
				CreateTime:   time.Now(),
			})
		}
		// 如果 exists -> 忽略（根据需求：重复则不管）
	}

	// B. 遍历 DB，找出【删除】(Excel中没有的)
	for phone, status := range dbMap {
		if _, exists := excelMap[phone]; !exists {
			// Excel 没有，数据库有 -> 准备删除
			// ⚠️ 保护机制：只有 status == 0 (空闲) 的才删
			// 如果 status == 1 (锁定/待办) 或 2 (已办)，不能删
			if status == 0 {
				toDeletePhones = append(toDeletePhones, phone)
			}
		}
	}

	// 5. 执行事务 (批量操作)
	err = l.svcCtx.Db.Transaction(func(tx *gorm.DB) error {
		// 批量删除 (Excel 没传且是空闲的)
		if len(toDeletePhones) > 0 {
			// 每次删 1000 条，防止 IN 语句过长
			batchSize := 1000
			for i := 0; i < len(toDeletePhones); i += batchSize {
				end := i + batchSize
				if end > len(toDeletePhones) {
					end = len(toDeletePhones)
				}
				// 这里执行硬删除或者软删除，取决于你的 Model 定义
				// 如果你希望只是标记删除，可以用 .Update("is_deleted", 1)
				// 这里用 Delete，配合 Where
				if err := tx.Where("phone_number IN ?", toDeletePhones[i:end]).Delete(&model.PhonePool{}).Error; err != nil {
					return err
				}
			}
		}

		// 批量插入 (Excel 新增的)
		if len(toInsert) > 0 {
			// Gorm 的 CreateInBatches 自动处理分批
			if err := tx.CreateInBatches(toInsert, 500).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		l.Logger.Errorf("全量导入失败: %v", err)
		return nil, err
	}

	return &types.ImportPhonesRes{
		SuccessCount: int(len(toInsert)), // 这里返回新增的数量
	}, nil
}
