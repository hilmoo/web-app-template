package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"web-app-template-core/internal/config"

	"github.com/urfave/cli/v3"
)

const httpClientTimeout = 5 * time.Second

func healthcheckCommand(cfg config.Config) *cli.Command {
	return &cli.Command{
		Name:  "healthcheck",
		Usage: "Perform a healthcheck by hitting the /api/health endpoint",
		Action: func(ctx context.Context, _ *cli.Command) error {
			url := fmt.Sprintf("http://localhost:%d/api/health", cfg.ListenPort)
			client := http.Client{
				Timeout: httpClientTimeout,
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("healthcheck failed: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("healthcheck failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusNoContent {
				return fmt.Errorf("healthcheck failed: unexpected status code %d", resp.StatusCode)
			}

			return nil
		},
	}
}
