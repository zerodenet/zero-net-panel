package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/zero-net-panel/zero-net-panel/internal/config"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
)

func runGRPCServer(ctx context.Context, cfg config.Config, svcCtx *svc.ServiceContext) error {
	grpcCfg := cfg.GRPC
	if !grpcCfg.Enabled() {
		return nil
	}

	_ = svcCtx

	addr := grpcCfg.ListenOn
	if strings.TrimSpace(addr) == "" {
		return errors.New("grpc listen address is required")
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("start gRPC listener: %w", err)
	}
	defer func() {
		if cerr := lis.Close(); cerr != nil && !errors.Is(cerr, net.ErrClosed) {
			fmt.Printf("failed to close gRPC listener: %v\n", cerr)
		}
	}()

	server := grpc.NewServer()

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	if grpcCfg.ReflectionEnabled() {
		reflection.Register(server)
	}

	fmt.Printf("Starting gRPC API at %s...\n", addr)

	errCh := make(chan error, 1)
	go func() {
		if serveErr := server.Serve(lis); serveErr != nil &&
			!errors.Is(serveErr, grpc.ErrServerStopped) &&
			!errors.Is(serveErr, net.ErrClosed) {
			errCh <- serveErr
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		stopped := make(chan struct{})
		go func() {
			server.GracefulStop()
			close(stopped)
		}()

		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			server.Stop()
		}
		return nil
	case err, ok := <-errCh:
		if !ok {
			return nil
		}
		return err
	}
}
