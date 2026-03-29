package notifystock_test

import (
	"context"
	"testing"

	"github.com/uptrace/bun"

	notify "github.com/heyjun3/notify-stock/internal"
)

func openDB(t *testing.T) *bun.DB {
	t.Helper()
	dsn := "postgres://postgres:postgres@localhost:5555/notify-stock-test?sslmode=disable"
	db := notify.NewDB(dsn, notify.CreateLogger("DEBUG"))
	for _, table := range []any{
		(*notify.Stock)(nil),
		(*notify.Notification)(nil),
		(*notify.SymbolDetail)(nil),
		(*notify.Member)(nil),
		(*notify.GoogleMember)(nil),
	} {
		db.NewDelete().Model(table).Where("1 = 1").Exec(context.Background())
	}
	t.Cleanup(func() {
		db.Close()
	})
	return db
}
