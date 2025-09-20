package orders

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	handlercommon "github.com/zero-net-panel/zero-net-panel/internal/handler/common"
	adminorders "github.com/zero-net-panel/zero-net-panel/internal/logic/admin/orders"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// AdminListOrdersHandler lists orders with administrative filters and pagination.
func AdminListOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminListOrdersRequest
		if err := httpx.Parse(r, &req); err != nil {
			handlercommon.RespondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := adminorders.NewListLogic(r.Context(), svcCtx)
		resp, err := logic.List(&req)
		if err != nil {
			handlercommon.RespondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// AdminGetOrderHandler returns a detailed order snapshot.
func AdminGetOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminGetOrderRequest
		if err := httpx.Parse(r, &req); err != nil {
			handlercommon.RespondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := adminorders.NewGetLogic(r.Context(), svcCtx)
		resp, err := logic.Get(&req)
		if err != nil {
			handlercommon.RespondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// AdminMarkOrderPaidHandler marks an order as paid by an administrator.
func AdminMarkOrderPaidHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminMarkOrderPaidRequest
		if err := httpx.Parse(r, &req); err != nil {
			handlercommon.RespondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := adminorders.NewMarkPaidLogic(r.Context(), svcCtx)
		resp, err := logic.MarkPaid(&req)
		if err != nil {
			handlercommon.RespondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// AdminCancelOrderHandler cancels an order manually.
func AdminCancelOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminCancelOrderRequest
		if err := httpx.Parse(r, &req); err != nil {
			handlercommon.RespondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := adminorders.NewCancelLogic(r.Context(), svcCtx)
		resp, err := logic.Cancel(&req)
		if err != nil {
			handlercommon.RespondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// AdminRefundOrderHandler performs a balance refund for an order.
func AdminRefundOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminRefundOrderRequest
		if err := httpx.Parse(r, &req); err != nil {
			handlercommon.RespondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := adminorders.NewRefundLogic(r.Context(), svcCtx)
		resp, err := logic.Refund(&req)
		if err != nil {
			handlercommon.RespondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// AdminPaymentCallbackHandler processes external payment callbacks and updates order states.
func AdminPaymentCallbackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminPaymentCallbackRequest
		if err := httpx.Parse(r, &req); err != nil {
			handlercommon.RespondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := adminorders.NewPaymentCallbackLogic(r.Context(), svcCtx)
		resp, err := logic.Process(&req)
		if err != nil {
			handlercommon.RespondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
