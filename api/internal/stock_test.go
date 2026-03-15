package notifystock_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	notify "github.com/heyjun3/notify-stock/internal"
	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	db := openDB(t)
	repo := notify.NewStockRepository(db)

	tests := []struct {
		name   string
		stocks []notify.Stock
		err    error
	}{
		{
			stocks: []notify.Stock{{Symbol: "N225", Timestamp: time.Now(), Open: 1000, Close: 2000, High: 2500, Low: 500}},
			err:    nil,
		},
		{
			stocks: []notify.Stock{
				{Symbol: "N225", Timestamp: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)},
				{Symbol: "N225", Timestamp: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)},
				{Symbol: "SP500", Timestamp: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)},
				{Symbol: "SP500", Timestamp: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)},
			},
			err: nil,
		},
		{
			stocks: []notify.Stock{},
			err:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Save(context.Background(), tt.stocks)

			assert.Equal(t, tt.err, err)
		})
	}
}

func TestGetStockByPeriod(t *testing.T) {
	db := openDB(t)
	repo := notify.NewStockRepository(db)
	if err := repo.Save(
		context.Background(),
		[]notify.Stock{{Symbol: "N225", Timestamp: time.Now().AddDate(0, -2, 0),
			Open: 1000, Close: 2000, High: 2500, Low: 500}}); err != nil {
		panic(err)
	}
	tests := []struct {
		name      string
		symbol    string
		begging   time.Time
		end       time.Time
		err       error
		minLength int
	}{{
		symbol: "N225",
		err:    fmt.Errorf("symbol N225 not found"),
	}, {
		symbol:    "N225",
		begging:   time.Now().AddDate(-1, 0, 0),
		end:       time.Now(),
		err:       nil,
		minLength: 1,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stocks, err := repo.GetStockByPeriod(context.Background(), tt.symbol, tt.begging, tt.end)

			assert.Equal(t, tt.err, err)
			assert.GreaterOrEqual(t, len(stocks), tt.minLength)
			for _, stock := range stocks {
				assert.True(t, tt.begging.Before(stock.Timestamp))
				assert.True(t, tt.end.After(stock.Timestamp))
			}
		})
	}
}

func TestGetLatestStock(t *testing.T) {
	db := openDB(t)
	repo := notify.NewStockRepository(db)
	stocks := make([]notify.Stock, 0, 100)
	now := time.Now().UTC().Round(time.Millisecond)
	for i := range 100 {
		t := now.AddDate(0, 0, -i)
		stocks = append(stocks, notify.Stock{
			Symbol: "S&P500", Timestamp: t, Open: 10, Close: 10, High: 10, Low: 10,
		})
	}
	if err := repo.Save(
		context.Background(),
		stocks); err != nil {
		panic(err)
	}
	tests := []struct {
		name      string
		symbol    string
		timestamp time.Time
		err       error
	}{{
		symbol:    "S&P500",
		timestamp: now,
		err:       nil,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stock, err := repo.GetLatestStock(context.Background(), tt.symbol)

			assert.NoError(t, err)
			assert.Equal(t, tt.timestamp, stock.Timestamp)
		})
	}
}

func TestStocksLatest(t *testing.T) {
	stocks := []notify.Stock{
		{
			Timestamp: time.Now(),
			Close:     10000,
		},
		{
			Timestamp: time.Now().AddDate(-1, 0, 0),
			Close:     90000,
		},
	}

	t.Run("", func(t *testing.T) {
		s, err := notify.NewStocks(notify.SymbolDetail{}, stocks)
		assert.NoError(t, err)
		latest := s.Latest()

		assert.Equal(t, float64(10000), latest.Close)
	})
}

func TestTimeCompare(t *testing.T) {
	t.Run("after", func(t *testing.T) {
		assert.True(t, time.Now().After(time.Now().AddDate(-1, 0, 0)))
	})
	t.Run("before", func(t *testing.T) {
		assert.True(t, time.Now().Before(time.Now().AddDate(0, 0, 1)))
	})
}

func TestRoundTime(t *testing.T) {
	t.Run("round", func(t *testing.T) {
		tt := time.Date(2023, 10, 1, 23, 2, 2, 2, time.UTC)
		assert.Equal(t, "2023-10-02T00:00:00Z", tt.Round(time.Hour*24).Format(time.RFC3339))
	})
	t.Run("truncate", func(t *testing.T) {
		tt := time.Date(2023, 10, 1, 23, 2, 2, 2, time.UTC)
		assert.Equal(t, "2023-10-01T00:00:00Z", tt.Truncate(time.Hour*24).Format(time.RFC3339))
	})
}
