package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	useraccount "github.com/zero-net-panel/zero-net-panel/internal/logic/user/account"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// UserBalanceHandler 用户余额查询。
func UserBalanceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserBalanceRequest
		if err := httpx.Parse(r, &req); err != nil {
			respondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := useraccount.NewBalanceLogic(r.Context(), svcCtx)
		resp, err := logic.Balance(&req)
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
