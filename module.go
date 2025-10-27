package sqlfx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-core-fx/fxutil"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"db",
		fxutil.WithNamedLogger("db"),
		fx.Provide(New),
		fx.Invoke(func(lc fx.Lifecycle, db *sql.DB, logger *zap.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("starting database")
					if err := db.PingContext(ctx); err != nil {
						return fmt.Errorf("ping database: %w", err)
					}
					logger.Info("database started")

					return nil
				},
				OnStop: func(_ context.Context) error {
					logger.Info("shutting down database")
					if err := db.Close(); err != nil {
						return fmt.Errorf("close database: %w", err)
					}
					logger.Info("database shutdown completed")
					return nil
				},
			})
		}),
	)
}
