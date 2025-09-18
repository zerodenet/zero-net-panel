package cli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	errCh := make(chan error, 3)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := runHTTPServer(runCtx, cfg, svcCtx); err != nil {
			errCh <- err
		}
	}()

	if cfg.Metrics.Enabled() && cfg.Metrics.Standalone() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := runMetricsServer(runCtx, cfg.Metrics); err != nil {
				errCh <- err
			}
		}()
	}

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

	if cfg.Metrics.Enabled() && !cfg.Metrics.Standalone() {
		metricsHandler := promhttp.Handler()
		server.AddRoute(rest.Route{
			Method:  http.MethodGet,
			Path:    cfg.Metrics.Path,
			Handler: metricsHandler.ServeHTTP,
		})
		fmt.Printf("Prometheus metrics available at http://%s:%d%s\n", cfg.Host, cfg.Port, cfg.Metrics.Path)
	}

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

func runMetricsServer(ctx context.Context, cfg config.MetricsConfig) error {
	mux := http.NewServeMux()
	mux.Handle(cfg.Path, promhttp.Handler())

	server := &http.Server{
		Addr:    cfg.ListenOn,
		Handler: mux,
	}

	fmt.Printf("Starting Prometheus metrics server at %s%s...\n", cfg.ListenOn, cfg.Path)

	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		if err, ok := <-errCh; ok && err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		if err != nil {
			return err
		}
		return nil
	}
}
