package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"web-app-template-core/internal/config"
	"web-app-template-core/internal/db/migration"

	"github.com/golang-migrate/migrate/v4"
	"github.com/hilmoo/gomox/sqlx"
	"github.com/urfave/cli/v3"
)

func migrateCommand(cfg config.Config, logger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Database migration commands",
		Commands: []*cli.Command{
			{
				Name:  "up",
				Usage: "Migrate the DB to the most recent version available",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runMigration(ctx, cfg, logger, "up", cmd.Args())
				},
			},
			{
				Name:  "up-to",
				Usage: "Migrate the DB to a specific version",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runMigration(ctx, cfg, logger, "up-to", cmd.Args())
				},
			},
			{
				Name:  "down",
				Usage: "Roll back the version by 1",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runMigration(ctx, cfg, logger, "down", cmd.Args())
				},
			},
			{
				Name:  "down-to",
				Usage: "Roll back to a specific version",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runMigration(ctx, cfg, logger, "down-to", cmd.Args())
				},
			},
			{
				Name:  "version",
				Usage: "Print the current migration version",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runMigration(ctx, cfg, logger, "version", cmd.Args())
				},
			},
			{
				Name:  "force",
				Usage: "Force set a migration version (clears dirty state)",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runMigration(ctx, cfg, logger, "force", cmd.Args())
				},
			},
		},
	}
}

func runMigration(ctx context.Context, cfg config.Config, logger *slog.Logger, command string, args cli.Args) error {
	m, cleanup, err := sqlx.NewMigratePgx(
		ctx,
		cfg.DatabaseURL,
		migration.MigrationsFS,
		logger,
	)
	if err != nil {
		return err
	}
	defer cleanup()

	switch command {
	case "up":
		return migrateUp(m)
	case "up-to":
		return migrateToVersion(m, args, "up-to")
	case "down":
		return migrateDown(m)
	case "down-to":
		return migrateToVersion(m, args, "down-to")
	case "version":
		return migrateVersion(m)
	case "force":
		return migrateForce(m, args)
	default:
		return fmt.Errorf("unknown migration command: %s", command)
	}
}

func runStep(err error, msg string) error {
	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println(msg) //nolint:forbidigo // CLI progress output
		return nil
	}
	return err
}

func migrateUp(m *migrate.Migrate) error {
	return runStep(m.Up(), "migrate up: no changes")
}

func migrateDown(m *migrate.Migrate) error {
	return runStep(m.Steps(-1), "migrate down: no changes")
}

func migrateToVersion(m *migrate.Migrate, args cli.Args, label string) error {
	if args.Len() < 1 {
		return fmt.Errorf("missing version argument for 'migrate %s'", label)
	}
	version, err := strconv.ParseUint(args.First(), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid version %q: %w", args.First(), err)
	}
	return runStep(m.Migrate(uint(version)), fmt.Sprintf("migrate %s %d: no changes", label, version))
}

func migrateVersion(m *migrate.Migrate) error {
	v, dirty, err := m.Version()
	if err != nil {
		return fmt.Errorf("version: %w", err)
	}
	fmt.Printf("current version: %d (dirty=%v)\n", v, dirty) //nolint:forbidigo // CLI output
	return nil
}

func migrateForce(m *migrate.Migrate, args cli.Args) error {
	if args.Len() < 1 {
		return fmt.Errorf("missing version argument for 'migrate force'")
	}
	version, err := strconv.ParseInt(args.First(), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid version %q: %w", args.First(), err)
	}
	return m.Force(int(version))
}
