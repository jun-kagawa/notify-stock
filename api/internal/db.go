package notifystock

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bunslog"
)

func NewDB(dsn string, logger *slog.Logger) *bun.DB {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn), pgdriver.WithTimeout(2*time.Second)))
	db := bun.NewDB(sqldb, pgdialect.New())
	hook := bunslog.NewQueryHook(
		bunslog.WithLogger(logger),
		bunslog.WithQueryLogLevel(slog.LevelDebug),
	)
	return db.WithQueryHook(hook)
}
