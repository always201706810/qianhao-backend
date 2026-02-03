package manage

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"qianhao-backend/internal/logic/manage"
	"qianhao-backend/internal/svc"
	"qianhao-backend/internal/types"
)

func UpdatePhoneHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdatePhoneReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := manage.NewUpdatePhoneLogic(r.Context(), svcCtx)
		resp, err := l.UpdatePhone(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
