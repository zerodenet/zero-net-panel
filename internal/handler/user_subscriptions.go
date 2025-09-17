package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	usersub "github.com/zero-net-panel/zero-net-panel/internal/logic/user/subscription"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// UserListSubscriptionsHandler 用户订阅列表。
func UserListSubscriptionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserListSubscriptionsRequest
		if err := httpx.Parse(r, &req); err != nil {
			respondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := usersub.NewListLogic(r.Context(), svcCtx)
		resp, err := logic.List(&req)
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// UserSubscriptionPreviewHandler 用户订阅预览。
func UserSubscriptionPreviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserSubscriptionPreviewRequest
		if err := httpx.Parse(r, &req); err != nil {
			respondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := usersub.NewPreviewLogic(r.Context(), svcCtx)
		resp, err := logic.Preview(&req)
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// UserUpdateSubscriptionTemplateHandler 用户切换订阅模板。
func UserUpdateSubscriptionTemplateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserUpdateSubscriptionTemplateRequest
		if err := httpx.Parse(r, &req); err != nil {
			respondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := usersub.NewUpdateTemplateLogic(r.Context(), svcCtx)
		resp, err := logic.UpdateTemplate(&req)
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
