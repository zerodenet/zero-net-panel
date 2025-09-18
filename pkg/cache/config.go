package cache

import "time"

type Config struct {
	Provider string       `json:"provider" yaml:"Provider"`
	Redis    RedisConfig  `json:"redis,omitempty" yaml:"Redis,omitempty"`
	Memory   MemoryConfig `json:"memory,omitempty" yaml:"Memory,omitempty"`
}

type RedisConfig struct {
	Host        string        `json:"host" yaml:"Host"`
	Type        string        `json:"type" yaml:"Type"`
	Password    string        `json:"password" yaml:"Password"`
	TLS         bool          `json:"tls" yaml:"TLS"`
	NonBlock    bool          `json:"nonBlock" yaml:"NonBlock"`
	PingTimeout time.Duration `json:"pingTimeout" yaml:"PingTimeout"`
}

type MemoryConfig struct {
	Size int `json:"size" yaml:"Size"`
}
