package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	adminplans "github.com/zero-net-panel/zero-net-panel/internal/logic/admin/plans"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// AdminListPlansHandler 套餐列表。
func AdminListPlansHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminListPlansRequest
		if err := httpx.Parse(r, &req); err != nil {
			respondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := adminplans.NewListLogic(r.Context(), svcCtx)
		resp, err := logic.List(&req)
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// AdminCreatePlanHandler 创建套餐。
func AdminCreatePlanHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminCreatePlanRequest
		if err := httpx.Parse(r, &req); err != nil {
			respondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := adminplans.NewCreateLogic(r.Context(), svcCtx)
		resp, err := logic.Create(&req)
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// AdminUpdatePlanHandler 更新套餐。
func AdminUpdatePlanHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminUpdatePlanRequest
		if err := httpx.Parse(r, &req); err != nil {
			respondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := adminplans.NewUpdateLogic(r.Context(), svcCtx)
		resp, err := logic.Update(&req)
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
