package shared

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/zero-net-panel/zero-net-panel/internal/logic"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
)

// PingHandler provides a lightweight health probe for the service runtime.
func PingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewPingLogic(r.Context(), svcCtx)
		resp, err := l.Ping()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
