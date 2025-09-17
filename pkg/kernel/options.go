package kernel

import "time"

// Options 描述注册表初始化所需的配置。
type Options struct {
	DefaultProtocol string
	HTTP            HTTPOptions
	GRPC            GRPCOptions
}

// HTTPOptions 是 HTTP Provider 所需配置。
type HTTPOptions struct {
	BaseURL string
	Token   string
	Timeout time.Duration
}

// GRPCOptions 是 gRPC Provider 所需配置。
type GRPCOptions struct {
	Endpoint string
	TLSCert  string
	Timeout  time.Duration
}
