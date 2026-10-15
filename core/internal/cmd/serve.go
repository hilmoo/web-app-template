package cmd

import (
	"context"
	"log/slog"

	"web-app-template-core/internal/app"
	"web-app-template-core/internal/config"

	"github.com/urfave/cli/v3"
)

func serveCommand(cfg config.Config, logger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "Run the application server",
		Action: func(ctx context.Context, _ *cli.Command) error {
			logger.InfoContext(ctx, "Starting application server...")
			if err := app.Main(cfg, logger); err != nil {
				return err
			}
			return nil
		},
	}
}
