package notifystock

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type SymbolDetail struct {
	bun.BaseModel `bun:"table:symbols"`

	Symbol        string          `bun:"symbol,pk"`
	ShortName     string          `bun:"short_name"`
	LongName      string          `bun:"long_name"`
	MarketPrice   decimal.Decimal `bun:"market_price"`
	PreviousClose decimal.Decimal `bun:"previous_close"`
	Volume        sql.NullInt64   `bun:"volume"`
	MarketCap     sql.NullInt64   `bun:"market_cap"`
	Currency      *Currency       `bun:"currency"`
}

func NewSymbolDetail(symbol, shortName, longName, currency string,
	marketPrice, previousClose decimal.Decimal,
	options ...SymbolDetailOption) *SymbolDetail {
	cur, err := CurrencyString(currency)
	if err != nil {
		return nil
	}
	detail := &SymbolDetail{
		Symbol:        symbol,
		ShortName:     shortName,
		LongName:      longName,
		MarketPrice:   marketPrice.Round(2),
		PreviousClose: previousClose.Round(2),
		Currency:      &cur,
	}
	for _, option := range options {
		option(detail)
	}
	return detail
}

func (s *SymbolDetail) Change() string {
	change := s.MarketPrice.Sub(s.PreviousClose)
	if change.IsPositive() {
		return "+" + change.String()
	}
	return change.String()
}

func (s *SymbolDetail) ChangePercent() string {
	if s.PreviousClose.IsZero() {
		return "0%"
	}
	p := s.MarketPrice.Sub(s.PreviousClose).Div(s.PreviousClose).
		Mul(decimal.New(100, 0)).Round(2)
	if p.IsPositive() {
		return "+" + p.String() + "%"
	}
	return p.String() + "%"
}

func (s *SymbolDetail) Key() string {
	return s.Symbol
}

type SymbolDetailOption func(detail *SymbolDetail) *SymbolDetail

func WithVolume(volume int64) SymbolDetailOption {
	return func(detail *SymbolDetail) *SymbolDetail {
		detail.Volume = sql.NullInt64{Int64: volume, Valid: true}
		return detail
	}
}
func WithMarketCap(marketCap int64) SymbolDetailOption {
	return func(detail *SymbolDetail) *SymbolDetail {
		detail.MarketCap = sql.NullInt64{Int64: marketCap, Valid: true}
		return detail
	}
}

type SymbolRepository struct {
	db *bun.DB
}

func NewSymbolRepository(db *bun.DB) *SymbolRepository {
	return &SymbolRepository{
		db: db,
	}
}

func (r *SymbolRepository) Save(ctx context.Context, details []SymbolDetail) error {
	_, err := r.db.NewInsert().Model(&details).
		On("CONFLICT (symbol) DO UPDATE").
		Set(strings.Join([]string{
			"short_name = EXCLUDED.short_name",
			"long_name = EXCLUDED.long_name",
			"market_price = EXCLUDED.market_price",
			"previous_close = EXCLUDED.previous_close",
			"volume = EXCLUDED.volume",
			"market_cap = EXCLUDED.market_cap",
			"currency = EXCLUDED.currency",
		}, ",")).
		Exec(ctx)
	return err
}

func (r *SymbolRepository) Get(ctx context.Context, symbol string) (*SymbolDetail, error) {
	var detail SymbolDetail
	err := r.db.NewSelect().Model(&detail).
		Where("symbol = ?", symbol).
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func (r *SymbolRepository) GetBySymbols(
	ctx context.Context, symbols []string) ([]SymbolDetail, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("required len")
	}
	var details []SymbolDetail
	err := r.db.NewSelect().Model(&details).
		Where("symbol IN (?)", bun.In(symbols)).
		Order("symbol ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return details, nil
}

func (r *SymbolRepository) GetAll(ctx context.Context) ([]SymbolDetail, error) {
	var details []SymbolDetail
	err := r.db.NewSelect().Model(&details).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return details, nil
}
