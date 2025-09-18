package dashboard

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	adminroutes "github.com/zero-net-panel/zero-net-panel/internal/admin/routes"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// DashboardLogic 管理后台概览信息。
type DashboardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDashboardLogic 构造函数。
func NewDashboardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DashboardLogic {
	return &DashboardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Modules 返回可用的管理后台模块。
func (l *DashboardLogic) Modules() (*types.AdminDashboardResponse, error) {
	modules, err := l.svcCtx.Repositories.AdminModule.ListModules(l.ctx)
	if err != nil {
		return nil, err
	}

	resp := &types.AdminDashboardResponse{Modules: make([]types.AdminModule, 0, len(modules))}
	prefix := l.svcCtx.Config.Admin.RoutePrefix
	for _, module := range modules {
		resp.Modules = append(resp.Modules, mapModule(module, prefix))
	}

	return resp, nil
}

func mapModule(module repository.AdminModule, prefix string) types.AdminModule {
	return types.AdminModule{
		Key:         module.Key,
		Name:        module.Name,
		Description: module.Description,
		Icon:        module.Icon,
		Route:       adminroutes.APIPath(module.Route, prefix),
		Permissions: append([]string(nil), module.Permissions...),
	}
}
