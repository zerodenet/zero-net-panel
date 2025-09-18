package dashboard

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	handlercommon "github.com/zero-net-panel/zero-net-panel/internal/handler/common"
	admindashboard "github.com/zero-net-panel/zero-net-panel/internal/logic/admin/dashboard"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
)

// AdminDashboardHandler returns the module overview for the administration console.
func AdminDashboardHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logic := admindashboard.NewDashboardLogic(r.Context(), svcCtx)
		resp, err := logic.Modules()
		if err != nil {
			handlercommon.RespondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
