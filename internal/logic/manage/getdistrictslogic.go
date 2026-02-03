package manage

import (
	"context"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDistrictsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDistrictsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDistrictsLogic {
	return &GetDistrictsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDistrictsLogic) GetDistricts() (resp *types.DistrictListRes, err error) {
	// 1. 定义数据库模型切片
	var districts []model.SysDistrict

	// 2. 查询所有区县 (GORM 查询)
	result := l.svcCtx.Db.Find(&districts)
	if result.Error != nil {
		l.Logger.Errorf("查询区县列表失败: %v", result.Error)
		return nil, result.Error
	}

	// 3. 转换格式 (从 Model 转为 API Types)
	var list []types.DistrictInfo
	for _, d := range districts {
		list = append(list, types.DistrictInfo{
			Id:   d.Id,
			Name: d.Name,
		})
	}

	// 4. 返回结果
	return &types.DistrictListRes{
		List: list,
	}, nil
}
