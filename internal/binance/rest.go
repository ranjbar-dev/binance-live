package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/binance-live/internal/config"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RESTClient handles HTTP requests to Binance REST API
type RESTClient struct {
	baseURL    string
	httpClient *http.Client
	limiter    *rate.Limiter
	logger     *zap.Logger
}

// NewRESTClient creates a new Binance REST API client
func NewRESTClient(cfg *config.BinanceConfig, logger *zap.Logger) *RESTClient {

	// Create rate limiter based on config (requests per minute)
	requestsPerSecond := float64(cfg.RestRateLimit) / 60.0
	limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), cfg.RestRateLimit)

	return &RESTClient{
		baseURL: cfg.APIURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		limiter: limiter,
		logger:  logger,
	}
}

// doRequest performs an HTTP GET request with rate limiting
func (c *RESTClient) doRequest(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {

	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {

		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Build URL
	reqURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	if params != nil {

		reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {

		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {

		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {

		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil {

			return nil, &apiErr
		}
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetExchangeInfo retrieves exchange information including trading pairs
func (c *RESTClient) GetExchangeInfo(ctx context.Context) (*ExchangeInfoResponse, error) {

	body, err := c.doRequest(ctx, "/api/v3/exchangeInfo", nil)
	if err != nil {

		return nil, err
	}

	var info ExchangeInfoResponse
	if err := json.Unmarshal(body, &info); err != nil {

		return nil, fmt.Errorf("failed to unmarshal exchange info: %w", err)
	}

	return &info, nil
}

// GetKlines retrieves kline/candlestick data
func (c *RESTClient) GetKlines(ctx context.Context, symbol, interval string, startTime, endTime *time.Time, limit int) ([]KlineResponse, error) {

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("interval", interval)

	if startTime != nil {

		params.Set("startTime", strconv.FormatInt(startTime.UnixMilli(), 10))
	}

	if endTime != nil {

		params.Set("endTime", strconv.FormatInt(endTime.UnixMilli(), 10))
	}

	if limit > 0 {

		params.Set("limit", strconv.Itoa(limit))
	}

	body, err := c.doRequest(ctx, "/api/v3/klines", params)
	if err != nil {

		return nil, err
	}

	var klines []KlineResponse
	if err := json.Unmarshal(body, &klines); err != nil {

		return nil, fmt.Errorf("failed to unmarshal klines: %w", err)
	}

	return klines, nil
}

// GetTicker24hr retrieves 24hr ticker price change statistics
func (c *RESTClient) GetTicker24hr(ctx context.Context, symbol string) (*Ticker24hrResponse, error) {

	params := url.Values{}
	params.Set("symbol", symbol)

	body, err := c.doRequest(ctx, "/api/v3/ticker/24hr", params)
	if err != nil {

		return nil, err
	}

	var ticker Ticker24hrResponse
	if err := json.Unmarshal(body, &ticker); err != nil {

		return nil, fmt.Errorf("failed to unmarshal ticker: %w", err)
	}

	return &ticker, nil
}

// GetAllTickers24hr retrieves 24hr ticker for all symbols
func (c *RESTClient) GetAllTickers24hr(ctx context.Context) ([]Ticker24hrResponse, error) {

	body, err := c.doRequest(ctx, "/api/v3/ticker/24hr", nil)
	if err != nil {

		return nil, err
	}

	var tickers []Ticker24hrResponse
	if err := json.Unmarshal(body, &tickers); err != nil {

		return nil, fmt.Errorf("failed to unmarshal tickers: %w", err)
	}

	return tickers, nil
}

// GetDepth retrieves order book depth
func (c *RESTClient) GetDepth(ctx context.Context, symbol string, limit int) (*DepthResponse, error) {

	params := url.Values{}
	params.Set("symbol", symbol)

	if limit > 0 {

		params.Set("limit", strconv.Itoa(limit))
	}

	body, err := c.doRequest(ctx, "/api/v3/depth", params)
	if err != nil {

		return nil, err
	}

	var depth DepthResponse
	if err := json.Unmarshal(body, &depth); err != nil {

		return nil, fmt.Errorf("failed to unmarshal depth: %w", err)
	}

	return &depth, nil
}

// GetAggTrades retrieves aggregated trades
func (c *RESTClient) GetAggTrades(ctx context.Context, symbol string, startTime, endTime *time.Time, limit int) ([]AggTradeResponse, error) {

	params := url.Values{}
	params.Set("symbol", symbol)

	if startTime != nil {

		params.Set("startTime", strconv.FormatInt(startTime.UnixMilli(), 10))
	}

	if endTime != nil {

		params.Set("endTime", strconv.FormatInt(endTime.UnixMilli(), 10))
	}

	if limit > 0 {

		params.Set("limit", strconv.Itoa(limit))
	}

	body, err := c.doRequest(ctx, "/api/v3/aggTrades", params)
	if err != nil {

		return nil, err
	}

	var trades []AggTradeResponse
	if err := json.Unmarshal(body, &trades); err != nil {

		return nil, fmt.Errorf("failed to unmarshal trades: %w", err)
	}

	return trades, nil
}

// GetServerTime retrieves the server time
func (c *RESTClient) GetServerTime(ctx context.Context) (time.Time, error) {

	body, err := c.doRequest(ctx, "/api/v3/time", nil)
	if err != nil {

		return time.Time{}, err
	}

	var result struct {
		ServerTime int64 `json:"serverTime"`
	}

	if err := json.Unmarshal(body, &result); err != nil {

		return time.Time{}, fmt.Errorf("failed to unmarshal server time: %w", err)
	}

	return time.UnixMilli(result.ServerTime), nil
}

// Ping tests connectivity to the REST API
func (c *RESTClient) Ping(ctx context.Context) error {

	_, err := c.doRequest(ctx, "/api/v3/ping", nil)

	return err
}
