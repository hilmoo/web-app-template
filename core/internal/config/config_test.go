package config_test

import (
	"maps"
	"strings"
	"testing"

	"web-app-template-core/internal/config"
)

const (
	strongSecretA = "0123456789abcdef0123456789abcdef"
	strongSecretB = "fedcba9876543210fedcba9876543210"
)

func setEnv(t *testing.T, env map[string]string) {
	t.Helper()
	base := map[string]string{
		"DATABASE_URL":   "postgresql://localhost/db",
		"SECRET":         "supersecret",
		"SERVICE_SECRET": "supersecret",
		"LISTEN_ADDR":    "",
		"APP_ENV":        "",
	}
	maps.Copy(base, env)
	for k, v := range base {
		t.Setenv(k, v)
	}
}

func TestLoadConfigDefaults(t *testing.T) { //nolint:paralleltest // t.Setenv forbids t.Parallel
	setEnv(t, nil)

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.ListenHost != "127.0.0.1" || cfg.ListenPort != 8080 {
		t.Errorf("listen = %s:%d, want 127.0.0.1:8080", cfg.ListenHost, cfg.ListenPort)
	}
	if cfg.IsProd {
		t.Error("IsProd = true, want false")
	}
}

func TestLoadConfigInvalidListenAddr(t *testing.T) { //nolint:paralleltest // t.Setenv forbids t.Parallel
	for _, addr := range []string{"nonsense", "127.0.0.1:abc", "127.0.0.1:70000"} { //nolint:paralleltest // t.Setenv forbids t.Parallel
		t.Run(addr, func(t *testing.T) {
			setEnv(t, map[string]string{"LISTEN_ADDR": addr})
			if _, err := config.LoadConfig(); err == nil {
				t.Errorf("LoadConfig() with LISTEN_ADDR=%q: want error", addr)
			}
		})
	}
}

func TestLoadConfigProductionSecrets(t *testing.T) { //nolint:paralleltest // t.Setenv forbids t.Parallel
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{
			name:    "placeholder secrets",
			env:     map[string]string{},
			wantErr: "SECRET",
		},
		{
			name:    "short service secret",
			env:     map[string]string{"SECRET": strongSecretA, "SERVICE_SECRET": "short"},
			wantErr: "SERVICE_SECRET",
		},
		{
			name:    "identical secrets",
			env:     map[string]string{"SECRET": strongSecretA, "SERVICE_SECRET": strongSecretA},
			wantErr: "must differ",
		},
		{
			name:    "strong distinct secrets",
			env:     map[string]string{"SECRET": strongSecretA, "SERVICE_SECRET": strongSecretB},
			wantErr: "",
		},
	}
	for _, tt := range tests { //nolint:paralleltest // t.Setenv forbids t.Parallel
		t.Run(tt.name, func(t *testing.T) {
			tt.env["APP_ENV"] = "production"
			setEnv(t, tt.env)

			_, err := config.LoadConfig()
			switch {
			case tt.wantErr == "" && err != nil:
				t.Fatalf("LoadConfig() error = %v, want nil", err)
			case tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)):
				t.Fatalf("LoadConfig() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}
