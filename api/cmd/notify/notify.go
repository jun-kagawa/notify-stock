package notify

import (
	"context"
	"log"
	"log/slog"

	"github.com/spf13/cobra"

	notifyapp "github.com/heyjun3/notify-stock/internal"
)

var NotifyCommand = &cobra.Command{
	Use:   "notify",
	Short: "Stock summary notification for yesterday",
	Run: func(cmd *cobra.Command, args []string) {
		notifyStock(symbols)
	},
}

var symbols []string

func init() {
	NotifyCommand.Flags().StringSliceVarP(&symbols, "symbol", "s", []string{}, "array of symbol")
}

func notifyStock(symbols []string) {
	db := notifyapp.NewDB(notifyapp.Cfg.DBDSN, notifyapp.CreateLogger(slog.LevelDebug))
	notifier, err := notifyapp.InitStockNotifier(
		context.Background(),
		db,
		notifyapp.MailGunClientConfig{
			Domain: notifyapp.Cfg.MailDomain,
			ApiKey: notifyapp.Cfg.MailGunAPIKey,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := notifier.Notify(symbols); err != nil {
		log.Fatal(err)
	}
}
