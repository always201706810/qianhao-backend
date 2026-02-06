package system

import (
	"context"

	"qianhao-backend/internal/model"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserListLogic {
	return &UserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserListLogic) UserList(req *types.PageReq) (resp *types.UserListRes, err error) {
	// 定义一个临时结构体来接收 Join 结果
	type Result struct {
		model.SysUser
		DistrictName string `gorm:"column:district_name"`
	}
	var results []Result
	var total int64

	db := l.svcCtx.Db.Table("sys_user").
		Select("sys_user.*, sys_district.name as district_name").
		Joins("LEFT JOIN sys_district ON sys_user.district_id = sys_district.id")

	db.Count(&total)

	offset := (req.Page - 1) * req.Size
	if err := db.Offset(offset).Limit(req.Size).Scan(&results).Error; err != nil {
		return nil, err
	}
	var list []types.UserInfo
	for _, item := range results {
		var distId int
		if item.DistrictId != nil {
			distId = *item.DistrictId
		}

		list = append(list, types.UserInfo{
			Id:       item.Id,
			Username: item.Username,
			// 修改这里：数据库取出来的是 RealName，赋值给 API 的 NickName
			NickName:     item.RealName,
			Role:         item.Role,
			DistrictId:   distId,
			DistrictName: item.DistrictName,
			CreateTime:   item.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.UserListRes{
		List:  list,
		Total: int(total),
	}, nil
}
