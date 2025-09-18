package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	userorder "github.com/zero-net-panel/zero-net-panel/internal/logic/user/order"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// UserCreateOrderHandler 提交套餐订单。
func UserCreateOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserCreateOrderRequest
		if err := httpx.Parse(r, &req); err != nil {
			respondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := userorder.NewCreateLogic(r.Context(), svcCtx)
		resp, err := logic.Create(&req)
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// UserListOrdersHandler 列出用户订单。
func UserListOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserOrderListRequest
		if err := httpx.Parse(r, &req); err != nil {
			respondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := userorder.NewListLogic(r.Context(), svcCtx)
		resp, err := logic.List(&req)
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// UserGetOrderHandler 获取单个订单详情。
func UserGetOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserGetOrderRequest
		if err := httpx.Parse(r, &req); err != nil {
			respondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := userorder.NewGetLogic(r.Context(), svcCtx)
		resp, err := logic.Get(&req)
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
