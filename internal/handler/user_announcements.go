package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	userannouncement "github.com/zero-net-panel/zero-net-panel/internal/logic/user/announcement"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// UserListAnnouncementsHandler 用户公告列表。
func UserListAnnouncementsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserAnnouncementListRequest
		if err := httpx.Parse(r, &req); err != nil {
			respondError(w, r, repository.ErrInvalidArgument)
			return
		}

		logic := userannouncement.NewListLogic(r.Context(), svcCtx)
		resp, err := logic.List(&req)
		if err != nil {
			respondError(w, r, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
