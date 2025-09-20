package config

import (
	"testing"
)

func TestMetricsConfigNormalize(t *testing.T) {
	cfg := MetricsConfig{Enable: true, Path: "metrics", ListenOn: " 0.0.0.0:9100 "}
	cfg.Normalize()

	if cfg.Path != "/metrics" {
		t.Fatalf("expected /metrics path, got %q", cfg.Path)
	}
	if cfg.ListenOn != "0.0.0.0:9100" {
		t.Fatalf("expected trimmed listen address, got %q", cfg.ListenOn)
	}
}

func TestMetricsConfigNormalizeDisabled(t *testing.T) {
	cfg := MetricsConfig{Enable: false, Path: "  /custom ", ListenOn: "127.0.0.1:9000"}
	cfg.Normalize()

	if cfg.Path != "/custom" {
		t.Fatalf("expected sanitized path, got %q", cfg.Path)
	}
	if cfg.ListenOn != "" {
		t.Fatalf("expected listen address cleared when disabled, got %q", cfg.ListenOn)
	}
}

func TestConfigNormalizeSyncsMiddlewares(t *testing.T) {
	cfg := Config{}
	cfg.Metrics.Enable = true
	cfg.Metrics.Path = "metrics"

	cfg.Normalize()

	if !cfg.Middlewares.Prometheus {
		t.Fatal("prometheus middleware should be enabled when metrics is on")
	}
	if !cfg.Middlewares.Metrics {
		t.Fatal("metrics middleware should be enabled when metrics is on")
	}

	cfg.Metrics.Enable = false
	cfg.Normalize()

	if cfg.Middlewares.Prometheus {
		t.Fatal("prometheus middleware should be disabled when metrics is off")
	}
	if cfg.Middlewares.Metrics {
		t.Fatal("metrics middleware should be disabled when metrics is off")
	}
}
