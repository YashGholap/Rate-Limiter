package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.env")

	content := `REDIS_URL=redis://localhost:6379/0
HTTP_ADDR=:9090
RATE_LIMIT_CAPACITY=20
RATE_LIMIT_REFILL_RATE=4
`

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.RedisURL != "redis://localhost:6379/0" {
		t.Errorf("unexpected Redis URL: %s", cfg.RedisURL)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Errorf("unexpected HTTP address: %s", cfg.HTTPAddr)
	}
	if cfg.RateLimitCapacity != 20 {
		t.Errorf("unexpected capacity: %v", cfg.RateLimitCapacity)
	}
	if cfg.RateLimitRefill != 4 {
		t.Errorf("unexpected refill rate: %v", cfg.RateLimitRefill)
	}
}

func TestLoadDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.env")

	if err := os.WriteFile(
		path,
		[]byte("REDIS_URL=redis://localhost:6379/0\n"),
		0600,
	); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.HTTPAddr != ":8082" {
		t.Errorf("unexpected default HTTP address: %s", cfg.HTTPAddr)
	}
	if cfg.RateLimitCapacity != 10 {
		t.Errorf("unexpected default capacity: %v", cfg.RateLimitCapacity)
	}
	if cfg.RateLimitRefill != 2 {
		t.Errorf("unexpected default refill rate: %v", cfg.RateLimitRefill)
	}
}

func TestLoadRejectsInvalidCapacity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.env")

	content := `REDIS_URL=redis://localhost:6379/0
RATE_LIMIT_CAPACITY=0
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected invalid capacity to return an error")
	}
}

func TestLoadRejectsMissingRedisURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.env")

	if err := os.WriteFile(path, []byte("HTTP_ADDR=:8082\n"), 0600); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected missing REDIS_URL to return an error")
	}
}
