package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"

	adminAnnouncements "github.com/zero-net-panel/zero-net-panel/internal/handler/admin/announcements"
	adminDashboard "github.com/zero-net-panel/zero-net-panel/internal/handler/admin/dashboard"
	adminNodes "github.com/zero-net-panel/zero-net-panel/internal/handler/admin/nodes"
	adminOrders "github.com/zero-net-panel/zero-net-panel/internal/handler/admin/orders"
	adminPlans "github.com/zero-net-panel/zero-net-panel/internal/handler/admin/plans"
	adminSecurity "github.com/zero-net-panel/zero-net-panel/internal/handler/admin/security"
	adminTemplates "github.com/zero-net-panel/zero-net-panel/internal/handler/admin/templates"
	authhandlers "github.com/zero-net-panel/zero-net-panel/internal/handler/auth"
	sharedhandlers "github.com/zero-net-panel/zero-net-panel/internal/handler/shared"
	userAccount "github.com/zero-net-panel/zero-net-panel/internal/handler/user/account"
	userAnnouncements "github.com/zero-net-panel/zero-net-panel/internal/handler/user/announcements"
	userOrders "github.com/zero-net-panel/zero-net-panel/internal/handler/user/orders"
	userPlans "github.com/zero-net-panel/zero-net-panel/internal/handler/user/plans"
	userSubscriptions "github.com/zero-net-panel/zero-net-panel/internal/handler/user/subscriptions"
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
				Handler: sharedhandlers.PingHandler(svcCtx),
			},
		},
		rest.WithPrefix("/api/v1"),
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/login",
				Handler: authhandlers.AuthLoginHandler(svcCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/refresh",
				Handler: authhandlers.AuthRefreshHandler(svcCtx),
			},
		},
		rest.WithPrefix("/api/v1/auth"),
	)

	adminRoutes := []rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/dashboard",
			Handler: adminDashboard.AdminDashboardHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/nodes",
			Handler: adminNodes.AdminListNodesHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/nodes/:id/kernels",
			Handler: adminNodes.AdminNodeKernelsHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/nodes/:id/kernels/sync",
			Handler: adminNodes.AdminSyncNodeKernelHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/subscription-templates",
			Handler: adminTemplates.AdminListSubscriptionTemplatesHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/subscription-templates",
			Handler: adminTemplates.AdminCreateSubscriptionTemplateHandler(svcCtx),
		},
		{
			Method:  http.MethodPatch,
			Path:    "/subscription-templates/:id",
			Handler: adminTemplates.AdminUpdateSubscriptionTemplateHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/subscription-templates/:id/publish",
			Handler: adminTemplates.AdminPublishSubscriptionTemplateHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/subscription-templates/:id/history",
			Handler: adminTemplates.AdminSubscriptionTemplateHistoryHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/plans",
			Handler: adminPlans.AdminListPlansHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/plans",
			Handler: adminPlans.AdminCreatePlanHandler(svcCtx),
		},
		{
			Method:  http.MethodPatch,
			Path:    "/plans/:id",
			Handler: adminPlans.AdminUpdatePlanHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/announcements",
			Handler: adminAnnouncements.AdminListAnnouncementsHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/announcements",
			Handler: adminAnnouncements.AdminCreateAnnouncementHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/announcements/:id/publish",
			Handler: adminAnnouncements.AdminPublishAnnouncementHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/security-settings",
			Handler: adminSecurity.AdminGetSecuritySettingHandler(svcCtx),
		},
		{
			Method:  http.MethodPatch,
			Path:    "/security-settings",
			Handler: adminSecurity.AdminUpdateSecuritySettingHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/orders",
			Handler: adminOrders.AdminListOrdersHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/orders/:id",
			Handler: adminOrders.AdminGetOrderHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/orders/:id/pay",
			Handler: adminOrders.AdminMarkOrderPaidHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/orders/:id/cancel",
			Handler: adminOrders.AdminCancelOrderHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/orders/:id/refund",
			Handler: adminOrders.AdminRefundOrderHandler(svcCtx),
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
			Handler: userSubscriptions.UserListSubscriptionsHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/subscriptions/:id/preview",
			Handler: userSubscriptions.UserSubscriptionPreviewHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/subscriptions/:id/template",
			Handler: userSubscriptions.UserUpdateSubscriptionTemplateHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/plans",
			Handler: userPlans.UserListPlansHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/announcements",
			Handler: userAnnouncements.UserListAnnouncementsHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/account/balance",
			Handler: userAccount.UserBalanceHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/orders",
			Handler: userOrders.UserCreateOrderHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/orders/:id/cancel",
			Handler: userOrders.UserCancelOrderHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/orders",
			Handler: userOrders.UserListOrdersHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/orders/:id",
			Handler: userOrders.UserGetOrderHandler(svcCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/orders/:id/cancel",
			Handler: userOrders.UserCancelOrderHandler(svcCtx),
		},
	}
	userRoutes = rest.WithMiddlewares([]rest.Middleware{authMiddleware.RequireRoles("user"), thirdPartyMiddleware.Handler}, userRoutes...)
	server.AddRoutes(userRoutes, rest.WithPrefix("/api/v1/user"))
}
