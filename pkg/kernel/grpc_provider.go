package kernel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCDiscoveryProvider 提供 gRPC 方式的节点发现。
type GRPCDiscoveryProvider struct {
	opts GRPCOptions

	mu   sync.Mutex
	conn *grpc.ClientConn
}

// NewGRPCDiscoveryProvider 创建 gRPC Provider。
func NewGRPCDiscoveryProvider(opts GRPCOptions) (*GRPCDiscoveryProvider, error) {
	if opts.Endpoint == "" {
		return nil, fmt.Errorf("kernel grpc provider: endpoint required")
	}

	return &GRPCDiscoveryProvider{opts: opts}, nil
}

// Name 返回 Provider 名称。
func (p *GRPCDiscoveryProvider) Name() string {
	return "grpc"
}

func (p *GRPCDiscoveryProvider) ensureConn(ctx context.Context) (*grpc.ClientConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.conn != nil {
		return p.conn, nil
	}

	timeout := p.opts.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var opts []grpc.DialOption
	if p.opts.TLSCert != "" {
		cred, err := credentials.NewClientTLSFromFile(p.opts.TLSCert, "")
		if err != nil {
			return nil, fmt.Errorf("kernel grpc provider: load tls: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(cred))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.DialContext(dialCtx, p.opts.Endpoint, opts...)
	if err != nil {
		return nil, err
	}

	p.conn = conn
	return conn, nil
}

// FetchNodeConfig 目前仅完成连通性校验，后续接入具体 proto 实现。
func (p *GRPCDiscoveryProvider) FetchNodeConfig(ctx context.Context, nodeID string) (NodeConfig, error) {
	if _, err := p.ensureConn(ctx); err != nil {
		return NodeConfig{}, err
	}

	// 预留实际调用逻辑，待 proto 定义完成后补充。
	return NodeConfig{}, ErrNotImplemented
}

// Close 关闭 gRPC 连接。
func (p *GRPCDiscoveryProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.conn != nil {
		err := p.conn.Close()
		p.conn = nil
		return err
	}

	return nil
}
