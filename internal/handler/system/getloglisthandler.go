package system

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"qianhao-backend/internal/logic/system"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"
)

func GetLogListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LogListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := system.NewGetLogListLogic(r.Context(), svcCtx)
		resp, err := l.GetLogList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
