package nodes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// SyncLogic 触发节点内核同步。
type SyncLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSyncLogic 构造函数。
func NewSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncLogic {
	return &SyncLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Sync 执行同步流程。
func (l *SyncLogic) Sync(req *types.AdminSyncNodeKernelRequest) (*types.AdminSyncNodeKernelResponse, error) {
	protocol := strings.ToLower(strings.TrimSpace(req.Protocol))
	if protocol == "" {
		protocol = l.svcCtx.Kernel.DefaultProtocol()
	}

	provider, err := l.svcCtx.Kernel.Provider(protocol)
	if err != nil {
		return nil, err
	}

	config, err := provider.FetchNodeConfig(l.ctx, fmt.Sprintf("%d", req.NodeID))
	if err != nil {
		return nil, err
	}

	if config.Protocol == "" {
		config.Protocol = protocol
	}
	if config.RetrievedAt.IsZero() {
		config.RetrievedAt = time.Now().UTC()
	}

	record := repository.NodeKernel{
		Protocol:     config.Protocol,
		Endpoint:     config.Endpoint,
		Revision:     config.Revision,
		Status:       "synced",
		Config:       config.Payload,
		LastSyncedAt: config.RetrievedAt,
	}

	stored, err := l.svcCtx.Repositories.Node.RecordKernelSync(l.ctx, req.NodeID, record)
	if err != nil {
		return nil, err
	}

	message := "同步完成"
	if config.RetrievedAt.Sub(stored.LastSyncedAt) > time.Minute {
		message = "同步完成（注意：返回时间与存储存在偏差）"
	}

	resp := &types.AdminSyncNodeKernelResponse{
		NodeID:   req.NodeID,
		Protocol: stored.Protocol,
		Revision: stored.Revision,
		SyncedAt: stored.LastSyncedAt.Unix(),
		Message:  message,
	}

	return resp, nil
}
