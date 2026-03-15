package notifystock

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

func NewStock(symbol string, timestamp time.Time,
	open, close, high, low float64) (Stock, error) {
	for _, v := range []float64{open, close, high, low} {
		if v <= 0 {
			return Stock{}, fmt.Errorf(
				"value is higher than zero. open: %v, close: %v, high: %v, low: %v, timestamp: %v, symbol: %v",
				open, close, high, low, timestamp, symbol)
		}
	}
	truncated := timestamp.Truncate(time.Hour * 24)
	return Stock{
		Symbol:    symbol,
		Timestamp: truncated,
		Open:      open,
		Close:     close,
		High:      high,
		Low:       low,
	}, nil
}

type Stock struct {
	bun.BaseModel `bun:"table:stocks"`

	Symbol    string    `bun:"symbol,type:text,pk"`
	Timestamp time.Time `bun:"timestamp,type:timestamp,pk"`
	Open      float64   `bun:"open,type:decimal,notnull"`
	Close     float64   `bun:"close,type:decimal,notnull"`
	High      float64   `bun:"high,type:decimal,notnull"`
	Low       float64   `bun:"low,type:decimal,notnull"`
}

type Stocks struct {
	symbol SymbolDetail
	stocks []Stock
}

func NewStocks(symbol SymbolDetail, stocks []Stock) (*Stocks, error) {
	return &Stocks{
		symbol: symbol,
		stocks: stocks,
	}, nil
}

func (s *Stocks) Latest() Stock {
	return slices.MaxFunc(s.stocks, func(a, b Stock) int {
		return a.Timestamp.Compare(b.Timestamp)
	})
}

func (s *Stocks) ClosingAverage() (decimal.Decimal, error) {
	close := make([]float64, 0, len(s.stocks))
	for _, v := range s.stocks {
		close = append(close, v.Close)
	}
	return CalcAVG(close)
}

func (s *Stocks) ClosingPriceToAVGRatio() (decimal.Decimal, error) {
	avg, err := s.ClosingAverage()
	if err != nil {
		return decimal.Decimal{}, err
	}
	latest := s.Latest()
	ratio := decimal.NewFromFloat(latest.Close).Div(avg)
	return ratio, nil
}

func (s *Stocks) GenerateNotificationMessage() (string, error) {
	avg, err := s.ClosingAverage()
	if err != nil {
		return "", err
	}
	latest := s.Latest()
	fmt.Println("latest", latest.Close)
	ratio, err := s.ClosingPriceToAVGRatio()
	if err != nil {
		return "", err
	}
	currency := s.symbol.Currency
	text := strings.Join([]string{
		s.symbol.ShortName,
		fmt.Sprintf("Closing Price: %v %s", int(latest.Close), currency),
		fmt.Sprintf("1-Year Moving Average: %v %s", avg.Ceil(), currency),
		fmt.Sprintf("Closing Price to 1-Year Moving Average Ratio: %v%s", ratio.Mul(decimal.New(100, 0)).RoundCeil(2), "%"),
	}, "\n")
	return text, nil
}

func (s *Stocks) Json() ([]byte, error) {
	t := struct {
		Symbol SymbolDetail `json:"symbol"`
		Stocks []Stock      `json:"stocks"`
	}{
		Symbol: s.symbol,
		Stocks: s.stocks,
	}
	return json.Marshal(t)
}

func CalcAVG[T cmp.Ordered](values []T) (decimal.Decimal, error) {
	d := make([]decimal.Decimal, 0, len(values))
	for _, v := range values {
		deci, err := decimal.NewFromString(fmt.Sprintf("%v", v))
		if err != nil {
			logger.Error("failed convert string to decimal")
			return decimal.Decimal{}, err
		}
		d = append(d, deci)
	}
	avg := decimal.Avg(d[0], d[1:]...)
	return avg, nil
}

type Quotes []Stock

func (qs Quotes) Unique() Quotes {
	m := make(map[string]Quotes)
	for _, stock := range qs {
		m[stock.Symbol] = append(m[stock.Symbol], stock)
	}
	quotes := make(Quotes, 0, len(qs))
	for _, stocks := range m {
		m := make(map[int64]Stock, len(stocks))
		for _, stock := range stocks {
			a := stock.Timestamp.Unix()
			m[a] = stock
		}
		for _, stock := range m {
			quotes = append(quotes, stock)
		}
	}
	return quotes
}

type StockRepository struct {
	db *bun.DB
}

func NewStockRepository(db *bun.DB) *StockRepository {
	return &StockRepository{
		db: db,
	}
}

func (r *StockRepository) Save(ctx context.Context, stocks []Stock) error {
	if len(stocks) == 0 {
		return nil
	}
	quotes := Quotes(stocks).Unique()
	_, err := r.db.NewInsert().
		Model(&quotes).
		On("CONFLICT (symbol, timestamp) DO UPDATE").
		Set(strings.Join([]string{
			"open = EXCLUDED.open",
			"close = EXCLUDED.close",
			"high = EXCLUDED.high",
			"low = EXCLUDED.low",
		}, ",")).
		Exec(ctx)
	return err
}

func (r *StockRepository) GetStockByPeriod(
	ctx context.Context, symbol string, begging, end time.Time) (
	[]Stock, error) {
	stocks, err := r.GetStockByPeriodAndSymbols(
		ctx, []string{symbol}, begging, end,
	)
	if err != nil {
		return nil, err
	}
	if stock, ok := stocks[symbol]; !ok {
		return nil, fmt.Errorf("symbol %s not found", symbol)
	} else {
		return stock, nil
	}
}

func (r *StockRepository) GetStockByPeriodAndSymbols(
	ctx context.Context, symbols []string, begging, end time.Time) (
	map[string][]Stock, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("symbols is empty")
	}
	var stocks []Stock
	if err := r.db.NewSelect().
		Model(&stocks).
		DistinctOn("timestamp::date, symbol").
		Where("symbol IN (?)", bun.In(symbols)).
		Where("timestamp::date BETWEEN ? AND ?", begging, end).
		OrderExpr("timestamp::date").
		Order("symbol").
		Scan(ctx); err != nil {
		return nil, err
	}
	result := make(map[string][]Stock)
	for _, stock := range stocks {
		if arr, ok := result[stock.Symbol]; !ok {
			result[stock.Symbol] = []Stock{stock}
		} else {
			arr = append(arr, stock)
			result[stock.Symbol] = arr
		}
	}
	return result, nil
}

func (r *StockRepository) GetLatestStock(ctx context.Context, symbol string) (*Stock, error) {
	var stock Stock
	if err := r.db.NewSelect().
		Model(&stock).
		Where("symbol = ?", symbol).
		Order("timestamp DESC").
		Limit(0).
		Scan(ctx); err != nil {
		return nil, err
	}
	return &stock, nil
}
