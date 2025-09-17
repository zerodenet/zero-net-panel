package database

import "time"

type Config struct {
	Driver          string        `json:"driver" yaml:"Driver"`
	DSN             string        `json:"dsn" yaml:"DSN"`
	MaxOpenConns    int           `json:"maxOpenConns" yaml:"MaxOpenConns"`
	MaxIdleConns    int           `json:"maxIdleConns" yaml:"MaxIdleConns"`
	ConnMaxLifetime time.Duration `json:"connMaxLifetime" yaml:"ConnMaxLifetime"`
	LogLevel        string        `json:"logLevel" yaml:"LogLevel"`
}

func (c Config) IsEmpty() bool {
	return c.Driver == "" || c.DSN == ""
}
