package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	adminannouncements "github.com/zero-net-panel/zero-net-panel/internal/logic/admin/announcements"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// AdminListAnnouncementsHandler 公告列表。
func AdminListAnnouncementsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminListAnnouncementsRequest
		if err := httpx.Parse(r, &req); err != nil {
			respondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := adminannouncements.NewListLogic(r.Context(), svcCtx)
		resp, err := logic.List(&req)
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// AdminCreateAnnouncementHandler 创建公告。
func AdminCreateAnnouncementHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminCreateAnnouncementRequest
		if err := httpx.Parse(r, &req); err != nil {
			respondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := adminannouncements.NewCreateLogic(r.Context(), svcCtx)
		resp, err := logic.Create(&req)
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// AdminPublishAnnouncementHandler 发布公告。
func AdminPublishAnnouncementHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminPublishAnnouncementRequest
		if err := httpx.Parse(r, &req); err != nil {
			respondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := adminannouncements.NewPublishLogic(r.Context(), svcCtx)
		resp, err := logic.Publish(&req)
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
