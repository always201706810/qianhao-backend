package mini

import (
	"context"
	"fmt"
	//  删除了 strings 和 strconv，因为 Grade 是 int，不需要解析了

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNumbersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNumbersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNumbersLogic {
	return &GetNumbersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNumbersLogic) GetNumbers(req *types.MiniNumberListReq) (resp *types.MiniNumberListRes, err error) {
	var phones []model.PhonePool
	var total int64

	// 1. 构建查询
	db := l.svcCtx.Db.Model(&model.PhonePool{}).Where("is_deleted = ?", 0)

	// 筛选：类型
	if req.Category != "" {
		db = db.Where("category = ?", req.Category)
	}

	// 筛选：等级
	if req.Level > 0 {
		// 既然 Grade 是 int，直接精确查询
		db = db.Where("grade = ?", req.Level)
	}

	// 搜索：尾号或关键词
	if req.Keyword != "" {
		db = db.Where("phone_number LIKE ?", "%"+req.Keyword+"%")
	}

	// 2. 计算总数
	db.Count(&total)

	// 3. 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := db.Offset(offset).Limit(req.PageSize).Order("id asc").Scan(&phones).Error; err != nil {
		return nil, err
	}

	// 4. 转换数据格式
	var list []types.MiniNumberItem
	for _, p := range phones {
		// 修复：Grade 是 int，直接使用
		levelInt := p.Grade
		levelName := fmt.Sprintf("%d级", p.Grade) // 拼接显示名称

		// 判断锁定状态 (Status: 0-空闲, 1-锁定/待办, 2-已办)
		isLocked := p.Status > 0

		list = append(list, types.MiniNumberItem{
			Id:         p.Id,
			FullNumber: p.PhoneNumber,
			Level:      levelInt,  // 2
			LevelName:  levelName, // "2级"
			Price:      p.Price,
			Category:   p.Category,
			IsLocked:   isLocked,
			Tag:        levelName, // 标签也显示 "2级"
		})
	}

	return &types.MiniNumberListRes{
		List:    list,
		Total:   int(total),
		HasMore: int(total) > req.Page*req.PageSize,
	}, nil
}
