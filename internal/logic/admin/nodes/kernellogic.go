package nodes

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// KernelLogic 查询节点内核配置。
type KernelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewKernelLogic 构造函数。
func NewKernelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KernelLogic {
	return &KernelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Kernels 返回节点的协议列表。
func (l *KernelLogic) Kernels(req *types.AdminNodeKernelsRequest) (*types.AdminNodeKernelResponse, error) {
	kernels, err := l.svcCtx.Repositories.Node.GetKernels(l.ctx, req.NodeID)
	if err != nil {
		return nil, err
	}

	resp := &types.AdminNodeKernelResponse{
		NodeID:  req.NodeID,
		Kernels: make([]types.NodeKernelSummary, 0, len(kernels)),
	}

	for _, kernel := range kernels {
		resp.Kernels = append(resp.Kernels, types.NodeKernelSummary{
			Protocol:     kernel.Protocol,
			Endpoint:     kernel.Endpoint,
			Revision:     kernel.Revision,
			Status:       kernel.Status,
			Config:       kernel.Config,
			LastSyncedAt: kernel.LastSyncedAt.Unix(),
		})
	}

	return resp, nil
}
