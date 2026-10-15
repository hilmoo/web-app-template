package config

import (
	"fmt"
	"net"
	"strconv"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Base
	ListenAddr string `env:"LISTEN_ADDR" envDefault:"127.0.0.1:8080"`
	ListenHost string `env:"-"`
	ListenPort int    `env:"-"`
	LogLevel   string `env:"LOG_LEVEL"   envDefault:"INFO"`
	LogOutput  string `env:"LOG_OUTPUT"  envDefault:"stdout"`

	// Secret is used to hash session tokens.
	Secret string `env:"SECRET,required"`
	// ServiceSecret is the bearer token trusted services present (Authorization: Bearer ...).
	ServiceSecret string `env:"SERVICE_SECRET,required"`
	AppEnv        string `env:"APP_ENV"`
	IsProd        bool

	DatabaseURL string `env:"DATABASE_URL,required"`
}

func LoadConfig() (Config, error) {
	var c Config
	if err := env.Parse(&c); err != nil {
		return c, fmt.Errorf("env config: %w", err)
	}

	host, portStr, err := net.SplitHostPort(c.ListenAddr)
	if err != nil {
		return c, fmt.Errorf("invalid LISTEN_ADDR %q: %w", c.ListenAddr, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return c, fmt.Errorf("invalid port in LISTEN_ADDR %q: %w", c.ListenAddr, err)
	}
	if port < 0 || port > 65535 {
		return c, fmt.Errorf("port in LISTEN_ADDR %q out of range (0-65535)", c.ListenAddr)
	}

	c.ListenHost = host
	c.ListenPort = port

	c.IsProd = c.AppEnv == "production"

	return c, nil
}
