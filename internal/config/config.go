package config

import (
	"strings"
	"time"

	"github.com/zeromicro/go-zero/rest"

	"github.com/zero-net-panel/zero-net-panel/pkg/cache"
	"github.com/zero-net-panel/zero-net-panel/pkg/database"
)

type Config struct {
	rest.RestConf

	Project  ProjectConfig    `json:"project" yaml:"Project"`
	Database database.Config  `json:"database" yaml:"Database"`
	Cache    cache.Config     `json:"cache" yaml:"Cache"`
	Kernel   KernelConfig     `json:"kernel" yaml:"Kernel"`
	Auth     AuthConfig       `json:"auth" yaml:"Auth"`
	Metrics  MetricsConfig    `json:"metrics" yaml:"Metrics"`
	Admin    AdminConfig      `json:"admin" yaml:"Admin"`
	GRPC     GRPCServerConfig `json:"grpcServer" yaml:"GRPCServer"`
}

type ProjectConfig struct {
	Name        string `json:"name" yaml:"Name"`
	Description string `json:"description" yaml:"Description"`
	Version     string `json:"version" yaml:"Version"`
}

type KernelConfig struct {
	DefaultProtocol string           `json:"defaultProtocol" yaml:"DefaultProtocol"`
	HTTP            KernelHTTPConfig `json:"http" yaml:"HTTP"`
	GRPC            KernelGRPCConfig `json:"grpc" yaml:"GRPC"`
}

type KernelHTTPConfig struct {
	BaseURL string        `json:"baseUrl" yaml:"BaseURL"`
	Token   string        `json:"token" yaml:"Token"`
	Timeout time.Duration `json:"timeout" yaml:"Timeout"`
}

type KernelGRPCConfig struct {
	Endpoint string        `json:"endpoint" yaml:"Endpoint"`
	TLSCert  string        `json:"tlsCert" yaml:"TLSCert"`
	Timeout  time.Duration `json:"timeout" yaml:"Timeout"`
}

type AuthConfig struct {
	AccessSecret  string        `json:"accessSecret" yaml:"AccessSecret"`
	AccessExpire  time.Duration `json:"accessExpire" yaml:"AccessExpire"`
	RefreshSecret string        `json:"refreshSecret" yaml:"RefreshSecret"`
	RefreshExpire time.Duration `json:"refreshExpire" yaml:"RefreshExpire"`
}

type MetricsConfig struct {
	Enable   bool   `json:"enable" yaml:"Enable"`
	Path     string `json:"path" yaml:"Path"`
	ListenOn string `json:"listenOn" yaml:"ListenOn"`
}

// Normalize trims the path/listener and applies defaults.
func (m *MetricsConfig) Normalize() {
	path := strings.TrimSpace(m.Path)
	if path == "" {
		path = "/metrics"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	m.Path = path

	m.ListenOn = strings.TrimSpace(m.ListenOn)
	if !m.Enable {
		m.ListenOn = ""
	}
}

// Enabled returns whether metrics export is enabled.
func (m MetricsConfig) Enabled() bool {
	return m.Enable
}

// Standalone reports whether metrics should be served on an independent listener.
func (m MetricsConfig) Standalone() bool {
	return m.Enable && m.ListenOn != ""
}

// AdminConfig 控制管理端路由相关配置。
type AdminConfig struct {
	RoutePrefix string `json:"routePrefix" yaml:"RoutePrefix"`
}

// Normalize 统一前缀写法并设置默认值。
func (a *AdminConfig) Normalize() {
	prefix := strings.TrimSpace(a.RoutePrefix)
	prefix = strings.Trim(prefix, "/")
	if prefix == "" {
		prefix = "admin"
	}
	a.RoutePrefix = prefix
}

// APIBasePath 返回管理端挂载的完整 API 前缀。
func (a AdminConfig) APIBasePath() string {
	if a.RoutePrefix == "" {
		return "/api/v1"
	}
	return "/api/v1/" + a.RoutePrefix
}

// GRPCServerConfig 控制内建 gRPC 服务监听配置。
type GRPCServerConfig struct {
	Enable     *bool  `json:"enable" yaml:"Enable"`
	ListenOn   string `json:"listenOn" yaml:"ListenOn"`
	Reflection *bool  `json:"reflection" yaml:"Reflection"`
}

// Normalize 设置默认监听地址与开关。
func (g *GRPCServerConfig) Normalize() {
	if g.Enable == nil {
		g.Enable = boolPtr(true)
	}
	if g.Reflection == nil {
		g.Reflection = boolPtr(true)
	}
	if g.Enabled() && strings.TrimSpace(g.ListenOn) == "" {
		g.ListenOn = "0.0.0.0:8890"
	}
}

// Enabled 返回 gRPC 服务是否启用（默认为 true）。
func (g GRPCServerConfig) Enabled() bool {
	if g.Enable == nil {
		return true
	}
	return *g.Enable
}

// SetEnabled 修改 gRPC 启用状态。
func (g *GRPCServerConfig) SetEnabled(enabled bool) {
	g.Enable = boolPtr(enabled)
}

// ReflectionEnabled 返回是否开放 gRPC reflection（默认为 true）。
func (g GRPCServerConfig) ReflectionEnabled() bool {
	if g.Reflection == nil {
		return true
	}
	return *g.Reflection
}

// Normalize 将配置补齐默认值。
func (c *Config) Normalize() {
	c.Metrics.Normalize()
	c.Admin.Normalize()
	c.GRPC.Normalize()
	c.RestConf.Middlewares.Prometheus = c.Metrics.Enabled()
	c.RestConf.Middlewares.Metrics = c.Metrics.Enabled()
}

func boolPtr(v bool) *bool {
	return &v
}
