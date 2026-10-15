package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"web-app-template-core/internal/config"
	"web-app-template-core/internal/helper"
)

const (
	// errChBufferSize allows both the HTTP server goroutine and future background
	// goroutines to report a failure without blocking on send.
	errChBufferSize = 2

	gracefulShutdownTimeout = 10 * time.Second
)

func Main(cfg config.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	deps, cleanup, err := setupDependencies(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}
	defer cleanup()

	errCh := make(chan error, errChBufferSize)

	srv := startServer(deps, errCh)

	return awaitShutdown(ctx, deps.Logger, errCh, srv)
}

func startServer(deps *helper.HandlerParams, errCh chan<- error) *http.Server {
	// HTTP/1.1 for browsers and Connect, cleartext HTTP/2 for gRPC clients.
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	addr := fmt.Sprintf("%s:%d", deps.Config.ListenHost, deps.Config.ListenPort)
	srv := &http.Server{
		Addr:              addr,
		Handler:           Handler(*deps),
		Protocols:         protocols,
		ReadHeaderTimeout: gracefulShutdownTimeout,
	}

	go func() {
		deps.Logger.Info("Starting server", slog.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("server failed: %w", err)
		}
	}()

	return srv
}

func awaitShutdown(ctx context.Context, logger *slog.Logger, errCh <-chan error, servers ...*http.Server) error {
	select {
	case err := <-errCh:
		return err

	case <-ctx.Done():
		logger.InfoContext(ctx, "Interrupt received, beginning graceful shutdown...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()

		for _, srv := range servers {
			if err := srv.Shutdown(shutdownCtx); err != nil {
				logger.WarnContext(ctx, "Server shutdown timed out, forcing stop",
					slog.String("addr", srv.Addr), slog.Any("error", err))
				_ = srv.Close()
			} else {
				logger.InfoContext(ctx, "Server shut down cleanly", slog.String("addr", srv.Addr))
			}
		}

		return nil
	}
}
