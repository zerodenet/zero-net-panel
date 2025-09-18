package cli

import (
	"context"
	"fmt"
	"sync"

	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/rest"

	"github.com/zero-net-panel/zero-net-panel/internal/config"
	"github.com/zero-net-panel/zero-net-panel/internal/handler"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
)

// RunAPIServer 保留旧名称以兼容现有入口，内部委托给 RunServices。
func RunAPIServer(ctx context.Context, cfg config.Config) error {
	return RunServices(ctx, cfg)
}

// RunServices 启动 HTTP 与 gRPC 服务，并在任一退出或外部取消时统一回收资源。
func RunServices(ctx context.Context, cfg config.Config) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	svcCtx, err := svc.NewServiceContext(cfg)
	if err != nil {
		return err
	}
	defer svcCtx.Cleanup()

	proc.AddShutdownListener(func() {
		cancel()
	})

	errCh := make(chan error, 2)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := runHTTPServer(runCtx, cfg, svcCtx); err != nil {
			errCh <- err
		}
	}()

	if cfg.GRPC.Enabled() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := runGRPCServer(runCtx, cfg, svcCtx); err != nil {
				errCh <- err
			}
		}()
	}

	var runErr error
	select {
	case <-runCtx.Done():
	case runErr = <-errCh:
		cancel()
	}

	wg.Wait()

	return runErr
}

func runHTTPServer(ctx context.Context, cfg config.Config, svcCtx *svc.ServiceContext) error {
	server := rest.MustNewServer(cfg.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, svcCtx)

	fmt.Printf("Starting HTTP API at %s:%d...\n", cfg.Host, cfg.Port)

	done := make(chan struct{})
	go func() {
		server.Start()
		close(done)
	}()

	select {
	case <-ctx.Done():
		server.Stop()
		<-done
		return nil
	case <-done:
		return nil
	}
}
