package manage

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"qianhao-backend/internal/logic/manage"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"
)

func HandleOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HandleOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := manage.NewHandleOrderLogic(r.Context(), svcCtx)
		resp, err := l.HandleOrder(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
