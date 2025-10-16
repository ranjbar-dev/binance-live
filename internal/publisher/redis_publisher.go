package publisher

import (
	"context"
	"fmt"

	"github.com/binance-live/internal/models"
	"github.com/binance-live/internal/redis"
	"go.uber.org/zap"
)

// Publisher interface defines the contract for publishing live data
type Publisher interface {
	PublishKline(ctx context.Context, kline *models.Kline) error
	PublishTicker(ctx context.Context, ticker *models.Ticker) error
	PublishDepth(ctx context.Context, depth *models.DepthSnapshot) error
	PublishTrade(ctx context.Context, trade *models.Trade) error
	PublishAllSymbols(ctx context.Context, symbols []models.Symbol) error
}

// JSONPublisher handles publishing live data to Redis using JSON
type JSONPublisher struct {
	redis  *redis.Client
	logger *zap.Logger
}

// NewJSONPublisher creates a new JSON publisher
func NewJSONPublisher(redisClient *redis.Client, logger *zap.Logger) *JSONPublisher {
	return &JSONPublisher{
		redis:  redisClient,
		logger: logger,
	}
}

// New creates a new publisher (defaults to protobuf for better performance)
func New(redisClient *redis.Client, logger *zap.Logger) Publisher {
	return NewProtobufPublisher(redisClient, logger)
}

// PublishKline publishes kline data to Redis
func (p *JSONPublisher) PublishKline(ctx context.Context, kline *models.Kline) error {
	// Create live data structure
	liveData := models.LiveData{
		Type:      "kline",
		Symbol:    kline.Symbol,
		Timestamp: kline.OpenTime,
		Data: map[string]interface{}{
			"interval":               kline.Interval,
			"open_time":              kline.OpenTime / 1000,  // Convert milliseconds to seconds
			"close_time":             kline.CloseTime / 1000, // Convert milliseconds to seconds
			"open_price":             kline.OpenPrice,
			"high_price":             kline.HighPrice,
			"low_price":              kline.LowPrice,
			"close_price":            kline.ClosePrice,
			"volume":                 kline.Volume,
			"quote_volume":           kline.QuoteVolume,
			"trades_count":           kline.TradesCount,
			"taker_buy_volume":       kline.TakerBuyVolume,
			"taker_buy_quote_volume": kline.TakerBuyQuoteVolume,
		},
	}

	// Publish to channel
	channel := fmt.Sprintf("binance:kline:%s:%s", kline.Symbol, kline.Interval)
	if err := p.redis.PublishJSON(ctx, channel, liveData); err != nil {
		return fmt.Errorf("failed to publish kline: %w", err)
	}

	// Also store latest kline in Redis for quick access
	key := fmt.Sprintf("binance:latest:kline:%s:%s", kline.Symbol, kline.Interval)
	if err := p.redis.SetJSON(ctx, key, liveData, 0); err != nil {
		p.logger.Warn("Failed to cache kline in Redis",
			zap.String("symbol", kline.Symbol),
			zap.String("interval", kline.Interval),
			zap.Error(err),
		)
	}

	return nil
}

// PublishTicker publishes ticker data to Redis
func (p *JSONPublisher) PublishTicker(ctx context.Context, ticker *models.Ticker) error {
	liveData := models.LiveData{
		Type:      "ticker",
		Symbol:    ticker.Symbol,
		Timestamp: ticker.Timestamp,
		Data: map[string]interface{}{
			"price":                    ticker.Price,
			"bid_price":                ticker.BidPrice,
			"bid_qty":                  ticker.BidQty,
			"ask_price":                ticker.AskPrice,
			"ask_qty":                  ticker.AskQty,
			"volume_24h":               ticker.Volume24h,
			"quote_volume_24h":         ticker.QuoteVolume24h,
			"price_change_24h":         ticker.PriceChange24h,
			"price_change_percent_24h": ticker.PriceChangePercent24h,
			"high_24h":                 ticker.High24h,
			"low_24h":                  ticker.Low24h,
			"trades_count_24h":         ticker.TradesCount24h,
		},
	}

	// Publish to channel
	channel := fmt.Sprintf("binance:ticker:%s", ticker.Symbol)
	if err := p.redis.PublishJSON(ctx, channel, liveData); err != nil {
		return fmt.Errorf("failed to publish ticker: %w", err)
	}

	// Cache in Redis
	key := fmt.Sprintf("binance:latest:ticker:%s", ticker.Symbol)
	if err := p.redis.SetJSON(ctx, key, liveData, 0); err != nil {
		p.logger.Warn("Failed to cache ticker in Redis",
			zap.String("symbol", ticker.Symbol),
			zap.Error(err),
		)
	}

	return nil
}

// PublishDepth publishes depth data to Redis
func (p *JSONPublisher) PublishDepth(ctx context.Context, depth *models.DepthSnapshot) error {
	liveData := models.LiveData{
		Type:      "depth",
		Symbol:    depth.Symbol,
		Timestamp: depth.Timestamp,
		Data: map[string]interface{}{
			"last_update_id": depth.LastUpdateID,
			"bids":           depth.Bids,
			"asks":           depth.Asks,
		},
	}

	// Publish to channel
	channel := fmt.Sprintf("binance:depth:%s", depth.Symbol)
	if err := p.redis.PublishJSON(ctx, channel, liveData); err != nil {
		return fmt.Errorf("failed to publish depth: %w", err)
	}

	// Cache in Redis
	key := fmt.Sprintf("binance:latest:depth:%s", depth.Symbol)
	if err := p.redis.SetJSON(ctx, key, liveData, 0); err != nil {
		p.logger.Warn("Failed to cache depth in Redis",
			zap.String("symbol", depth.Symbol),
			zap.Error(err),
		)
	}

	return nil
}

// PublishTrade publishes trade data to Redis
func (p *JSONPublisher) PublishTrade(ctx context.Context, trade *models.Trade) error {
	liveData := models.LiveData{
		Type:      "trade",
		Symbol:    trade.Symbol,
		Timestamp: trade.Timestamp,
		Data: map[string]interface{}{
			"trade_id":       trade.TradeID,
			"price":          trade.Price,
			"quantity":       trade.Quantity,
			"quote_quantity": trade.QuoteQuantity,
			"is_buyer_maker": trade.IsBuyerMaker,
		},
	}

	// Publish to channel
	channel := fmt.Sprintf("binance:trade:%s", trade.Symbol)
	if err := p.redis.PublishJSON(ctx, channel, liveData); err != nil {
		return fmt.Errorf("failed to publish trade: %w", err)
	}

	return nil
}

// PublishAllSymbols publishes the list of all active symbols
func (p *JSONPublisher) PublishAllSymbols(ctx context.Context, symbols []models.Symbol) error {
	symbolList := make([]string, len(symbols))
	for i, s := range symbols {
		symbolList[i] = s.Symbol
	}

	key := "binance:symbols:active"
	if err := p.redis.SetJSON(ctx, key, symbolList, 0); err != nil {
		return fmt.Errorf("failed to publish symbols: %w", err)
	}

	return nil
}
