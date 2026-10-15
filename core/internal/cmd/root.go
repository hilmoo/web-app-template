package cmd

import (
	"context"
	"log/slog"
	"os"

	"web-app-template-core/internal/config"

	"github.com/hilmoo/gomox/log"
	"github.com/urfave/cli/v3"
)

func RootCmd() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger, err = log.New(cfg.LogLevel, cfg.LogOutput)
	if err != nil {
		logger.Error("failed to initialize logger", "error", err)
		os.Exit(1)
	}

	cmd := &cli.Command{
		Name:  "web-app-template",
		Usage: "Web app template server",
		Commands: []*cli.Command{
			serveCommand(cfg, logger),
			migrateCommand(cfg, logger),
			healthcheckCommand(cfg),
		},
	}

	if runErr := cmd.Run(context.Background(), os.Args); runErr != nil {
		logger.Error("command failed", "error", runErr)
		os.Exit(1)
	}
}
