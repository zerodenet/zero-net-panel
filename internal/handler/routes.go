package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"

	"github.com/zero-net-panel/zero-net-panel/internal/svc"
)

func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
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
		},
		rest.WithPrefix("/api/v1/admin"),
	)

	server.AddRoutes(
		[]rest.Route{
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
		},
		rest.WithPrefix("/api/v1/user"),
	)
}
