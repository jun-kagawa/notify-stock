package update

import (
	"net/http"
	"time"

	notify "github.com/heyjun3/notify-stock/internal"
	"github.com/spf13/cobra"
)

func init() {
	StockCommand.Flags().BoolVarP(&isAll, "all", "a", false,
		"register stock price data for the entire period")
}

var (
	isAll        bool
	StockCommand = &cobra.Command{
		Use:   "update",
		Short: "Update stock command",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			symbols, err := notify.GetSupportSymbols("config.yaml")
			if err != nil {
				panic(err)
			}
			start := time.Now().AddDate(0, 0, -7)
			if isAll {
				start = time.Now().AddDate(-5, 0, 0)
			}
			end := time.Now()
			register := notify.InitStockRegister(
				notify.NewDB(notify.Cfg.DBDSN, notify.CreateLogger("INFO")),
				&http.Client{},
			)
			if err := register.RegisterStockBySymbols(
				ctx,
				symbols.Symbols,
				start,
				end,
			); err != nil {
				panic(err)
			}
		},
	}
)
