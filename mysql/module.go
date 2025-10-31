package mysql

import (
	"database/sql"

	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"db.mysql",
		logger.WithNamedLogger("db.mysql"),
		fx.Provide(New),
		fx.Provide(func(w *wrapper) *sql.DB {
			return w.sql
		}),
	)
}
