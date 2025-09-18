package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	admindashboard "github.com/zero-net-panel/zero-net-panel/internal/logic/admin/dashboard"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
)

// AdminDashboardHandler 返回管理后台模块信息。
func AdminDashboardHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logic := admindashboard.NewDashboardLogic(r.Context(), svcCtx)
		resp, err := logic.Modules()
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
