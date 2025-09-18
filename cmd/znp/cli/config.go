package cli

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"

	"github.com/zero-net-panel/zero-net-panel/internal/config"
)

func loadConfig(configFile string) (config.Config, error) {
	var cfg config.Config
	if err := conf.Load(configFile, &cfg); err != nil {
		return config.Config{}, fmt.Errorf("load config %q: %w", configFile, err)
	}
	cfg.Normalize()
	return cfg, nil
}
