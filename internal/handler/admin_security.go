package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	securitylogic "github.com/zero-net-panel/zero-net-panel/internal/logic/admin/security"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// AdminGetSecuritySettingHandler 查询第三方安全配置。
func AdminGetSecuritySettingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := securitylogic.NewGetLogic(r.Context(), svcCtx)
		resp, err := l.Get()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// AdminUpdateSecuritySettingHandler 更新第三方安全配置。
func AdminUpdateSecuritySettingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminUpdateSecuritySettingRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := securitylogic.NewUpdateLogic(r.Context(), svcCtx)
		resp, err := l.Update(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
