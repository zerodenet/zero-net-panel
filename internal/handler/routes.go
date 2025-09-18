package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"

	"github.com/zero-net-panel/zero-net-panel/internal/middleware"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
)

func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
	authMiddleware := middleware.NewAuthMiddleware(svcCtx.Auth, svcCtx.Repositories.User)
	thirdPartyMiddleware := middleware.NewThirdPartyMiddleware(svcCtx.Repositories.Security)
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/ping",
				Handler: PingHandler(svcCtx),
			},
		},
		rest.WithPrefix("/api/v1"),
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/login",
				Handler: AuthLoginHandler(svcCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/refresh",
				Handler: AuthRefreshHandler(svcCtx),
			},
		},
		rest.WithPrefix("/api/v1/auth"),
	)

	adminRoutes := []rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/dashboard",
			Handler: AdminDashboardHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/nodes",
			Handler: AdminListNodesHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/nodes/:id/kernels",
			Handler: AdminNodeKernelsHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/nodes/:id/kernels/sync",
			Handler: AdminSyncNodeKernelHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/subscription-templates",
			Handler: AdminListSubscriptionTemplatesHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/subscription-templates",
			Handler: AdminCreateSubscriptionTemplateHandler(svcCtx),
		},
		{
			Method:  http.MethodPatch,
			Path:    "/subscription-templates/:id",
			Handler: AdminUpdateSubscriptionTemplateHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/subscription-templates/:id/publish",
			Handler: AdminPublishSubscriptionTemplateHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/subscription-templates/:id/history",
			Handler: AdminSubscriptionTemplateHistoryHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/plans",
			Handler: AdminListPlansHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/plans",
			Handler: AdminCreatePlanHandler(svcCtx),
		},
		{
			Method:  http.MethodPatch,
			Path:    "/plans/:id",
			Handler: AdminUpdatePlanHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/announcements",
			Handler: AdminListAnnouncementsHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/announcements",
			Handler: AdminCreateAnnouncementHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/announcements/:id/publish",
			Handler: AdminPublishAnnouncementHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/security-settings",
			Handler: AdminGetSecuritySettingHandler(svcCtx),
		},
		{
			Method:  http.MethodPatch,
			Path:    "/security-settings",
			Handler: AdminUpdateSecuritySettingHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/orders",
			Handler: AdminListOrdersHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/orders/:id",
			Handler: AdminGetOrderHandler(svcCtx),
		},
	}
	adminRoutes = rest.WithMiddlewares([]rest.Middleware{authMiddleware.RequireRoles("admin")}, adminRoutes...)
	adminPrefix := svcCtx.Config.Admin.RoutePrefix
	adminBase := "/api/v1"
	if adminPrefix != "" {
		adminBase += "/" + adminPrefix
	}
	server.AddRoutes(adminRoutes, rest.WithPrefix(adminBase))

	userRoutes := []rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/subscriptions",
			Handler: UserListSubscriptionsHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/subscriptions/:id/preview",
			Handler: UserSubscriptionPreviewHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/subscriptions/:id/template",
			Handler: UserUpdateSubscriptionTemplateHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/plans",
			Handler: UserListPlansHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/announcements",
			Handler: UserListAnnouncementsHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/account/balance",
			Handler: UserBalanceHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/orders",
			Handler: UserCreateOrderHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/orders",
			Handler: UserListOrdersHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/orders/:id",
			Handler: UserGetOrderHandler(svcCtx),
		},
	}
	userRoutes = rest.WithMiddlewares([]rest.Middleware{authMiddleware.RequireRoles("user"), thirdPartyMiddleware.Handler}, userRoutes...)
	server.AddRoutes(userRoutes, rest.WithPrefix("/api/v1/user"))
}
