package config

import (
	"time"

	"github.com/zeromicro/go-zero/rest"

	"github.com/zero-net-panel/zero-net-panel/pkg/cache"
	"github.com/zero-net-panel/zero-net-panel/pkg/database"
)

type Config struct {
	rest.RestConf

	Project  ProjectConfig   `json:"project" yaml:"Project"`
	Database database.Config `json:"database" yaml:"Database"`
	Cache    cache.Config    `json:"cache" yaml:"Cache"`
	Kernel   KernelConfig    `json:"kernel" yaml:"Kernel"`
	Auth     AuthConfig      `json:"auth" yaml:"Auth"`
	Metrics  MetricsConfig   `json:"metrics" yaml:"Metrics"`
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
	Enable bool   `json:"enable" yaml:"Enable"`
	Path   string `json:"path" yaml:"Path"`
}
