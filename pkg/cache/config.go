package cache

import "time"

type Config struct {
	Provider string       `json:"provider" yaml:"Provider"`
	Redis    RedisConfig  `json:"redis,optional" yaml:"Redis,optional"`
	Memory   MemoryConfig `json:"memory,optional" yaml:"Memory,optional"`
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
