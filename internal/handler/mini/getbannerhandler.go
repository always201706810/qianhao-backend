package mini

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"qianhao-backend/internal/logic/mini"
	"qianhao-backend/internal/svc"
)

func GetBannerHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := mini.NewGetBannerLogic(r.Context(), svcCtx)
		resp, err := l.GetBanner()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
