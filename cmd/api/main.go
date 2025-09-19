package main

import (
	"context"
	"flag"
	"log"

	"github.com/zeromicro/go-zero/core/conf"

	"github.com/zero-net-panel/zero-net-panel/cmd/znp/cli"
	"github.com/zero-net-panel/zero-net-panel/internal/bootstrap"
	"github.com/zero-net-panel/zero-net-panel/internal/config"
)

var configFile = flag.String("f", "etc/znp-api.yaml", "the config file")

func main() {
	flag.Parse()

	var cfg config.Config
	if err := conf.Load(*configFile, &cfg); err != nil {
		log.Fatalf("failed to load config %s: %v", *configFile, err)
	}
	cfg.Normalize()

	if _, err := bootstrap.PrepareDatabase(context.Background(), cfg, bootstrap.DatabaseOptions{AutoMigrate: true, TargetVersion: 0}); err != nil {
		log.Fatalf("failed to prepare database: %v", err)
	}

	if err := cli.RunServices(context.Background(), cfg); err != nil {
		log.Fatalf("services exited with error: %v", err)
	}
}
