package notifystock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

type HTTPClientInterface interface {
	Do(*http.Request) (*http.Response, error)
}

type FinanceClient struct {
	Client HTTPClientInterface
}

func NewFinanceClient(client HTTPClientInterface) *FinanceClient {
	return &FinanceClient{
		Client: client,
	}
}

type ChartResponse struct {
	Chart Chart `json:"chart"`
}
type Pre struct {
	Timezone  string `json:"timezone"`
	Start     int    `json:"start"`
	End       int    `json:"end"`
	Gmtoffset int    `json:"gmtoffset"`
}
type Regular struct {
	Timezone  string `json:"timezone"`
	Start     int    `json:"start"`
	End       int    `json:"end"`
	Gmtoffset int    `json:"gmtoffset"`
}
type Post struct {
	Timezone  string `json:"timezone"`
	Start     int    `json:"start"`
	End       int    `json:"end"`
	Gmtoffset int    `json:"gmtoffset"`
}
type CurrentTradingPeriod struct {
	Pre     Pre     `json:"pre"`
	Regular Regular `json:"regular"`
	Post    Post    `json:"post"`
}
type Meta struct {
	Currency             string               `json:"currency"`
	Symbol               string               `json:"symbol"`
	ExchangeName         string               `json:"exchangeName"`
	FullExchangeName     string               `json:"fullExchangeName"`
	InstrumentType       string               `json:"instrumentType"`
	FirstTradeDate       int                  `json:"firstTradeDate"`
	RegularMarketTime    int                  `json:"regularMarketTime"`
	HasPrePostMarketData bool                 `json:"hasPrePostMarketData"`
	Gmtoffset            int                  `json:"gmtoffset"`
	Timezone             string               `json:"timezone"`
	ExchangeTimezoneName string               `json:"exchangeTimezoneName"`
	RegularMarketPrice   float64              `json:"regularMarketPrice"`
	FiftyTwoWeekHigh     float64              `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow      float64              `json:"fiftyTwoWeekLow"`
	RegularMarketDayHigh float64              `json:"regularMarketDayHigh"`
	RegularMarketDayLow  float64              `json:"regularMarketDayLow"`
	RegularMarketVolume  int                  `json:"regularMarketVolume"`
	LongName             string               `json:"longName"`
	ShortName            string               `json:"shortName"`
	ChartPreviousClose   float64              `json:"chartPreviousClose"`
	PreviousClose        float64              `json:"previousClose"`
	PriceHint            int                  `json:"priceHint"`
	CurrentTradingPeriod CurrentTradingPeriod `json:"currentTradingPeriod"`
	DataGranularity      string               `json:"dataGranularity"`
	Range                string               `json:"range"`
	ValidRanges          []string             `json:"validRanges"`
}
type Quote struct {
	Volume []int     `json:"volume"`
	Close  []float64 `json:"close"`
	High   []float64 `json:"high"`
	Low    []float64 `json:"low"`
	Open   []float64 `json:"open"`
}
type Adjclose struct {
	Adjclose []float64 `json:"adjclose"`
}
type Indicators struct {
	Quote    []Quote    `json:"quote"`
	Adjclose []Adjclose `json:"adjclose"`
}
type Result struct {
	Meta       Meta       `json:"meta"`
	Timestamp  []int      `json:"timestamp"`
	Indicators Indicators `json:"indicators"`
}
type Chart struct {
	Result []Result `json:"result"`
	Error  any      `json:"error"`
}

func IsSameLen[T any](array ...[]T) bool {
	for i := range len(array) - 1 {
		if !(len(array[i]) == len(array[i+1])) {
			return false
		}
	}
	return true
}
func ConvertResponseToStock(res ChartResponse) (*Stocks, error) {
	result := res.Chart.Result
	if len(result) == 0 {
		return nil, fmt.Errorf("result is nil")
	}
	quote := result[0].Indicators.Quote
	if len(quote) == 0 {
		return nil, fmt.Errorf("quote is nil")
	}
	timestamp := result[0].Timestamp
	open := quote[0].Open
	close := quote[0].Close
	high := quote[0].High
	low := quote[0].Low
	if !IsSameLen(open, close, high, low) || len(timestamp) != len(open) {
		logger.Error(
			"same len error", "timestamp", len(timestamp), "open", len(open),
			"close", len(close), "high", len(high), "low", len(low))
		return nil, fmt.Errorf("don't same length error")
	}
	symbol := result[0].Meta.Symbol

	stocks := make([]Stock, 0, len(timestamp))
	for i, t := range timestamp {
		stock, err := NewStock(
			symbol, time.Unix(int64(t), 0),
			open[i], close[i], high[i], low[i],
		)
		if err != nil {
			logger.Error("new stock error", "error", err)
			continue
		}
		stocks = append(stocks, stock)
	}
	detail, err := ConvertResponseToSymbol(&res)
	if err != nil {
		return nil, err
	}
	s, err := NewStocks(*detail, stocks)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func ConvertResponseToSymbol(res *ChartResponse) (*SymbolDetail, error) {
	result := res.Chart.Result
	if len(result) == 0 {
		return nil, fmt.Errorf("result is nil")
	}
	meta := result[0].Meta

	previousClose, err := parsePreviousClose(res)
	if err != nil {
		return nil, err
	}
	detail := NewSymbolDetail(meta.Symbol, meta.ShortName, meta.LongName, meta.Currency,
		decimal.NewFromFloat(meta.RegularMarketPrice), decimal.NewFromFloat(previousClose))
	return detail, nil
}
func parsePreviousClose(res *ChartResponse) (float64, error) {
	result := res.Chart.Result
	if len(result) == 0 {
		return 0, fmt.Errorf("result is nil")
	}
	meta := result[0].Meta
	if meta.PreviousClose != 0 {
		return meta.PreviousClose, nil
	}
	adjclose := result[0].Indicators.Adjclose
	if len(adjclose) != 0 && len(adjclose[0].Adjclose) > 1 {
		return adjclose[0].Adjclose[len(adjclose[0].Adjclose)-2], nil
	}

	logger.Warn("failed to determine previous close", "symbol", meta.Symbol)
	return 0, nil
}

type Option func(URL *url.URL) *url.URL

func WithInterval(interval string) Option {
	return func(URL *url.URL) *url.URL {
		query := URL.Query()
		query.Add("interval", interval)
		URL.RawQuery = query.Encode()
		return URL
	}
}

func (c *FinanceClient) FetchCurrentStock(symbol string) (*Stocks, error) {
	now := time.Now()
	return c.FetchStock(symbol, now, now)
}

func (c *FinanceClient) FetchStock(
	symbol string, beggingOfPeriod, endOfPeriod time.Time, opts ...Option) (
	*Stocks, error) {
	URL, err := url.Parse(fmt.Sprintf("https://query2.finance.yahoo.com/v8/finance/chart/%s", symbol))
	if err != nil {
		return nil, err
	}
	query := URL.Query()
	query.Add("period1", strconv.Itoa(int(beggingOfPeriod.Unix())))
	query.Add("period2", strconv.Itoa(int(endOfPeriod.Unix())))
	query.Add("region", "US")

	URL.RawQuery = query.Encode()
	for _, opt := range opts {
		URL = opt(URL)
	}

	logger.Info("request", "url", URL.String())
	req, err := http.NewRequest(http.MethodGet, URL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:136.0) Gecko/20100101 Firefox/136.0")
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	logger.Info("request status", "code", res.StatusCode)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var chart ChartResponse
	if err := json.Unmarshal(body, &chart); err != nil {
		return nil, err
	}
	time.Sleep(time.Second * 1)
	return ConvertResponseToStock(chart)
}
