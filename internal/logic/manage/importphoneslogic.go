// package manage
//
// import (
//
//	"context"
//	"encoding/json"
//	"time"
//
//	"qianhao-backend/internal/model"
//	"qianhao-backend/internal/svc"
//	"qianhao-backend/internal/types"
//	"qianhao-backend/internal/utils"
//
//	"github.com/zeromicro/go-zero/core/logx"
//	"gorm.io/gorm"
//
// )
//
//	type ImportPhonesLogic struct {
//		logx.Logger
//		ctx    context.Context
//		svcCtx *svc.ServiceContext
//	}
//
//	func NewImportPhonesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ImportPhonesLogic {
//		return &ImportPhonesLogic{
//			Logger: logx.WithContext(ctx),
//			ctx:    ctx,
//			svcCtx: svcCtx,
//		}
//	}
//
// // ImportPhones 全量同步：新增的插入，存在的更新(价格/分类/等级)，消失的删除(仅限空闲)
//
//	func (l *ImportPhonesLogic) ImportPhones(req *types.ImportPhonesReq) (resp *types.ImportPhonesRes, err error) {
//		// 1. 获取当前操作员 ID
//		userIdNumber, _ := l.ctx.Value("userId").(json.Number)
//		userId, _ := userIdNumber.Int64()
//
//		// 2. 整理 Excel 数据到 Map
//		excelMap := make(map[string]*types.ImportItem)
//		for _, item := range req.List {
//			if item.PhoneNumber != "" {
//				temp := item
//				excelMap[item.PhoneNumber] = &temp
//			}
//		}
//
//		// 3. 查询数据库中【当前所有】的号码
//		// 我们需要查出更多字段来做比对：Price, Grade, Category
//		var dbPhones []model.PhonePool
//		if err := l.svcCtx.Db.Model(&model.PhonePool{}).
//			Where("is_deleted = ?", 0).
//			Scan(&dbPhones).Error; err != nil {
//			return nil, err
//		}
//
//		// 构造 DB Map，Key是号码，Value是整个对象(为了拿ID和对比字段)
//		dbMap := make(map[string]model.PhonePool)
//		for _, p := range dbPhones {
//			dbMap[p.PhoneNumber] = p
//		}
//
//		// 4. 核心比对：分类
//		var toInsert []model.PhonePool
//		var toUpdate []model.PhonePool // 新增：需要更新的列表
//		var toDeletePhones []string
//
//		// A. 遍历 Excel -> 找出【新增】和【更新】
//		for phone, item := range excelMap {
//			if dbItem, exists := dbMap[phone]; !exists {
//				// case 1: 数据库没有 -> 插入
//				toInsert = append(toInsert, model.PhonePool{
//					PhoneNumber:  item.PhoneNumber,
//					Category:     item.Category,
//					Grade:        item.Grade,
//					Price:        item.Price, // 写入价格
//					Status:       0,
//					ImportUserId: int(userId),
//					IsDeleted:    0,
//					CreateTime:   time.Now(),
//				})
//			} else {
//				// case 2: 数据库有 -> 检查是否需要更新
//				// 只要 价格、等级、分类 有任意一个不一样，就更新
//				// 注意：浮点数比较 price 可能会有精度问题，但这里直接 != 0 即可满足大部分场景
//				needUpdate := false
//				if dbItem.Price != item.Price {
//					needUpdate = true
//				}
//				if dbItem.Grade != item.Grade {
//					needUpdate = true
//				}
//				if dbItem.Category != item.Category {
//					needUpdate = true
//				}
//
//				if needUpdate {
//					// 准备更新的数据
//					toUpdate = append(toUpdate, model.PhonePool{
//						Id:       dbItem.Id, // 关键：必须带上ID
//						Price:    item.Price,
//						Grade:    item.Grade,
//						Category: item.Category,
//					})
//				}
//			}
//		}
//
//		// B. 遍历 DB -> 找出【删除】(Excel没传且是空闲的)
//		for phone, item := range dbMap {
//			if _, exists := excelMap[phone]; !exists {
//				if item.Status == 0 {
//					toDeletePhones = append(toDeletePhones, phone)
//				}
//			}
//		}
//
//		// 5. 执行事务
//		err = l.svcCtx.Db.Transaction(func(tx *gorm.DB) error {
//			// 批量删除
//			if len(toDeletePhones) > 0 {
//				// 分批删除防止 SQL 过长
//				batchSize := 1000
//				for i := 0; i < len(toDeletePhones); i += batchSize {
//					end := i + batchSize
//					if end > len(toDeletePhones) {
//						end = len(toDeletePhones)
//					}
//					if err := tx.Where("phone_number IN ?", toDeletePhones[i:end]).Delete(&model.PhonePool{}).Error; err != nil {
//						return err
//					}
//				}
//			}
//
//			// 批量插入
//			if len(toInsert) > 0 {
//				if err := tx.CreateInBatches(toInsert, 500).Error; err != nil {
//					return err
//				}
//			}
//
//			// 批量更新 (逐个更新)
//			// Gorm 批量更新不同值比较麻烦，最稳妥的是遍历 Update
//			// 虽然效率不如 Batch Insert，但对于几千条数据来说完全没问题
//			for _, updateItem := range toUpdate {
//				if err := tx.Model(&model.PhonePool{}).
//					Where("id = ?", updateItem.Id).
//					Updates(map[string]interface{}{
//						"price":    updateItem.Price,
//						"grade":    updateItem.Grade,
//						"category": updateItem.Category,
//					}).Error; err != nil {
//					return err
//				}
//			}
//			return nil
//		})
//
//		if err != nil {
//			l.Logger.Errorf("全量导入失败: %v", err)
//			return nil, err
//		}
//
//		// 记录日志
//		utils.AddLog(l.svcCtx, int(userId), "Admin", "批量导入号码", "Excel同步(含更新)", "")
//
//		return &types.ImportPhonesRes{
//			SuccessCount: int(len(toInsert) + len(toUpdate)), // 返回变动的总数
//		}, nil
//	}
package manage

import (
	"context"
	"encoding/json"
	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"
	"qianhao-backend/internal/utils"
	"time"

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

//	func (l *ImportPhonesLogic) ImportPhones(req *types.ImportPhonesReq) (resp *types.ImportPhonesRes, err error) {
//		userIdNumber, _ := l.ctx.Value("userId").(json.Number)
//		userId, _ := userIdNumber.Int64()
//
//		excelMap := make(map[string]*types.ImportItem)
//		for _, item := range req.List {
//			if item.PhoneNumber != "" {
//				temp := item
//				excelMap[item.PhoneNumber] = &temp
//			}
//		}
//
//		var dbPhones []model.PhonePool
//		// 仅查询未被删除的号码
//		if err := l.svcCtx.Db.Model(&model.PhonePool{}).
//			Where("is_deleted = ?", 0).
//			Scan(&dbPhones).Error; err != nil {
//			return nil, err
//		}
//
//		dbMap := make(map[string]model.PhonePool)
//		for _, p := range dbPhones {
//			dbMap[p.PhoneNumber] = p
//		}
//
//		var toInsert []model.PhonePool
//		var toUpdate []model.PhonePool
//		var toDeletePhones []string
//
//		for phone, item := range excelMap {
//			if dbItem, exists := dbMap[phone]; !exists {
//				toInsert = append(toInsert, model.PhonePool{
//					PhoneNumber:  item.PhoneNumber,
//					Category:     item.Category,
//					Grade:        item.Grade,
//					Price:        item.Price,
//					Status:       0,
//					ImportUserId: int(userId),
//					IsDeleted:    0,
//					CreateTime:   time.Now(),
//				})
//			} else {
//				needUpdate := false
//				if dbItem.Price != item.Price || dbItem.Grade != item.Grade || dbItem.Category != item.Category {
//					needUpdate = true
//				}
//				if needUpdate {
//					toUpdate = append(toUpdate, model.PhonePool{
//						Id:       dbItem.Id,
//						Price:    item.Price,
//						Grade:    item.Grade,
//						Category: item.Category,
//					})
//				}
//			}
//		}
//
//		// ✅ 只有当导入列表长度大于 1 时，才执行“消失即删除”逻辑
//		// 这样可以防止“单个录入”时意外清空整个号码池
//		if len(req.List) > 1 {
//			for phone, item := range dbMap {
//				if _, exists := excelMap[phone]; !exists {
//					// 仅限空闲号码才允许自动下架
//					if item.Status == 0 {
//						toDeletePhones = append(toDeletePhones, phone)
//					}
//				}
//			}
//		}
//
//		err = l.svcCtx.Db.Transaction(func(tx *gorm.DB) error {
//			// ✅ 修改点：使用软删除（Update is_deleted = 1）而非物理 Delete
//			if len(toDeletePhones) > 0 {
//				batchSize := 500
//				for i := 0; i < len(toDeletePhones); i += batchSize {
//					end := i + batchSize
//					if end > len(toDeletePhones) {
//						end = len(toDeletePhones)
//					}
//					// 使用 Updates 而不是 Delete 避开外键约束报错
//					if err := tx.Model(&model.PhonePool{}).
//						Where("phone_number IN ?", toDeletePhones[i:end]).
//						Update("is_deleted", 1).Error; err != nil {
//						return err
//					}
//				}
//			}
//
//			if len(toInsert) > 0 {
//				if err := tx.CreateInBatches(toInsert, 500).Error; err != nil {
//					return err
//				}
//			}
//
//			for _, updateItem := range toUpdate {
//				if err := tx.Model(&model.PhonePool{}).
//					Where("id = ?", updateItem.Id).
//					Updates(map[string]interface{}{
//						"price":      updateItem.Price,
//						"grade":      updateItem.Grade,
//						"category":   updateItem.Category,
//						"is_deleted": 0, // 确保如果是以前删掉的号又导进来了，能重新启用
//					}).Error; err != nil {
//					return err
//				}
//			}
//			return nil
//		})
//
//		if err != nil {
//			l.Logger.Errorf("号码导入失败: %v", err)
//			return nil, err
//		}
//
//		utils.AddLog(l.svcCtx, int(userId), "Admin", "号码库同步", "同步/更新号码数据", "")
//
//		return &types.ImportPhonesRes{
//			SuccessCount: int(len(toInsert) + len(toUpdate)),
//		}, nil
//	}
func (l *ImportPhonesLogic) ImportPhones(req *types.ImportPhonesReq) (resp *types.ImportPhonesRes, err error) {
	// 1. 获取当前操作员 ID
	userIdNumber, _ := l.ctx.Value("userId").(json.Number)
	userId, _ := userIdNumber.Int64()

	// ✅ 修复日志账号问题：查询当前管理员的真实用户名
	var currentUser model.SysUser
	l.svcCtx.Db.Select("username").First(&currentUser, userId)
	operatorName := currentUser.Username
	if operatorName == "" {
		operatorName = "Unknown" // 兜底
	}

	// 2. 整理 Excel 数据到 Map
	excelMap := make(map[string]*types.ImportItem)
	for _, item := range req.List {
		if item.PhoneNumber != "" {
			temp := item
			excelMap[item.PhoneNumber] = &temp
		}
	}

	// 3. 查询数据库中【所有】号码（包括已删除的，防止重复插入）
	var dbPhones []model.PhonePool
	// ✅ 移除 is_deleted = 0 限制，这样能识别出“已删除”的号，从而执行“复活”而不是“新增”
	if err := l.svcCtx.Db.Model(&model.PhonePool{}).Scan(&dbPhones).Error; err != nil {
		return nil, err
	}

	dbMap := make(map[string]model.PhonePool)
	for _, p := range dbPhones {
		dbMap[p.PhoneNumber] = p
	}

	var toInsert []model.PhonePool
	var toUpdate []model.PhonePool
	var toDeletePhones []string

	// 4. 比对逻辑
	for phone, item := range excelMap {
		if dbItem, exists := dbMap[phone]; !exists {
			// Case A: 彻底没见过这个号 -> 插入
			toInsert = append(toInsert, model.PhonePool{
				PhoneNumber:  item.PhoneNumber,
				Category:     item.Category,
				Grade:        item.Grade,
				Price:        item.Price,
				Status:       0,
				ImportUserId: int(userId),
				IsDeleted:    0,
				CreateTime:   time.Now(),
			})
		} else {
			// Case B: 见过这个号（可能是在用的，也可能是被删的）
			// 只要数据有变动，或者该号当前处于“删除”状态，就执行更新/复活
			needUpdate := false
			if dbItem.Price != item.Price || dbItem.Grade != item.Grade || dbItem.Category != item.Category || dbItem.IsDeleted == 1 {
				needUpdate = true
			}

			if needUpdate {
				toUpdate = append(toUpdate, model.PhonePool{
					Id:       dbItem.Id,
					Price:    item.Price,
					Grade:    item.Grade,
					Category: item.Category,
				})
			}
		}
	}

	// 只有批量导入时才自动下架消失的号码
	if len(req.List) > 1 {
		for phone, item := range dbMap {
			if _, exists := excelMap[phone]; !exists {
				// 仅限“未删除”且“空闲”的号码
				if item.IsDeleted == 0 && item.Status == 0 {
					toDeletePhones = append(toDeletePhones, phone)
				}
			}
		}
	}

	// 5. 事务执行
	err = l.svcCtx.Db.Transaction(func(tx *gorm.DB) error {
		// 软删除
		if len(toDeletePhones) > 0 {
			if err := tx.Model(&model.PhonePool{}).
				Where("phone_number IN ?", toDeletePhones).
				Update("is_deleted", 1).Error; err != nil {
				return err
			}
		}

		// 批量插入
		if len(toInsert) > 0 {
			if err := tx.CreateInBatches(toInsert, 500).Error; err != nil {
				return err
			}
		}

		// 批量更新/激活
		for _, updateItem := range toUpdate {
			if err := tx.Model(&model.PhonePool{}).
				Where("id = ?", updateItem.Id).
				Updates(map[string]interface{}{
					"price":      updateItem.Price,
					"grade":      updateItem.Grade,
					"category":   updateItem.Category,
					"is_deleted": 0, // ✅ 确保号码被激活/复活
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		l.Logger.Errorf("号码同步失败: %v", err)
		return nil, err
	}

	// ✅ 修复日志账号显示：使用 operatorName 而不是写死的 "Admin"
	utils.AddLog(l.svcCtx, int(userId), operatorName, "号码库同步", "同步/更新号码数据", "")

	return &types.ImportPhonesRes{
		SuccessCount: int(len(toInsert) + len(toUpdate)),
	}, nil
}
