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
				Path:    "/api/v1/ping",
				Handler: PingHandler(svcCtx),
			},
		},
		rest.WithPrefix(""),
	)
}
