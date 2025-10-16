package publisher

import (
	"context"
	"fmt"

	"github.com/binance-live/internal/models"
	"github.com/binance-live/internal/redis"
	binanceProto "github.com/binance-live/proto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// ProtobufPublisher handles publishing live data to Redis using protobuf
type ProtobufPublisher struct {
	redis  *redis.Client
	logger *zap.Logger
}

// NewProtobufPublisher creates a new protobuf publisher
func NewProtobufPublisher(redisClient *redis.Client, logger *zap.Logger) *ProtobufPublisher {
	return &ProtobufPublisher{
		redis:  redisClient,
		logger: logger,
	}
}

// PublishKline publishes kline data to Redis using protobuf
func (p *ProtobufPublisher) PublishKline(ctx context.Context, kline *models.Kline) error {
	// Create protobuf kline data
	klineData := &binanceProto.KlineData{
		Interval:            kline.Interval,
		OpenTime:            kline.OpenTime / 1000,  // Convert milliseconds to seconds
		CloseTime:           kline.CloseTime / 1000, // Convert milliseconds to seconds
		OpenPrice:           kline.OpenPrice,
		HighPrice:           kline.HighPrice,
		LowPrice:            kline.LowPrice,
		ClosePrice:          kline.ClosePrice,
		Volume:              kline.Volume,
		QuoteVolume:         kline.QuoteVolume,
		TradesCount:         int32(kline.TradesCount),
		TakerBuyVolume:      kline.TakerBuyVolume,
		TakerBuyQuoteVolume: kline.TakerBuyQuoteVolume,
	}

	// Create live data message
	liveData := &binanceProto.LiveData{
		Type:      binanceProto.DataType_DATA_TYPE_KLINE,
		Symbol:    kline.Symbol,
		Timestamp: kline.OpenTime,
		Data: &binanceProto.LiveData_Kline{
			Kline: klineData,
		},
	}

	// Publish to channel
	channel := fmt.Sprintf("binance:kline:%s:%s", kline.Symbol, kline.Interval)
	if err := p.redis.PublishProtobuf(ctx, channel, liveData); err != nil {
		return fmt.Errorf("failed to publish kline: %w", err)
	}

	// Also store latest kline in Redis for quick access
	key := fmt.Sprintf("binance:latest:kline:%s:%s", kline.Symbol, kline.Interval)
	if err := p.redis.SetProtobuf(ctx, key, liveData, 0); err != nil {
		p.logger.Warn("Failed to cache kline in Redis",
			zap.String("symbol", kline.Symbol),
			zap.String("interval", kline.Interval),
			zap.Error(err),
		)
	}

	return nil
}

// PublishTicker publishes ticker data to Redis using protobuf
func (p *ProtobufPublisher) PublishTicker(ctx context.Context, ticker *models.Ticker) error {
	// Create protobuf ticker data
	tickerData := &binanceProto.TickerData{
		Price: ticker.Price,
	}

	// Set optional fields
	if ticker.BidPrice != nil {
		tickerData.BidPrice = ticker.BidPrice
	}
	if ticker.BidQty != nil {
		tickerData.BidQty = ticker.BidQty
	}
	if ticker.AskPrice != nil {
		tickerData.AskPrice = ticker.AskPrice
	}
	if ticker.AskQty != nil {
		tickerData.AskQty = ticker.AskQty
	}
	if ticker.Volume24h != nil {
		tickerData.Volume_24H = ticker.Volume24h
	}
	if ticker.QuoteVolume24h != nil {
		tickerData.QuoteVolume_24H = ticker.QuoteVolume24h
	}
	if ticker.PriceChange24h != nil {
		tickerData.PriceChange_24H = ticker.PriceChange24h
	}
	if ticker.PriceChangePercent24h != nil {
		tickerData.PriceChangePercent_24H = ticker.PriceChangePercent24h
	}
	if ticker.High24h != nil {
		tickerData.High_24H = ticker.High24h
	}
	if ticker.Low24h != nil {
		tickerData.Low_24H = ticker.Low24h
	}
	if ticker.TradesCount24h != nil {
		tickerData.TradesCount_24H = proto.Int32(int32(*ticker.TradesCount24h))
	}

	// Create live data message
	liveData := &binanceProto.LiveData{
		Type:      binanceProto.DataType_DATA_TYPE_TICKER,
		Symbol:    ticker.Symbol,
		Timestamp: ticker.Timestamp,
		Data: &binanceProto.LiveData_Ticker{
			Ticker: tickerData,
		},
	}

	// Publish to channel
	channel := fmt.Sprintf("binance:ticker:%s", ticker.Symbol)
	if err := p.redis.PublishProtobuf(ctx, channel, liveData); err != nil {
		return fmt.Errorf("failed to publish ticker: %w", err)
	}

	// Cache in Redis
	key := fmt.Sprintf("binance:latest:ticker:%s", ticker.Symbol)
	if err := p.redis.SetProtobuf(ctx, key, liveData, 0); err != nil {
		p.logger.Warn("Failed to cache ticker in Redis",
			zap.String("symbol", ticker.Symbol),
			zap.Error(err),
		)
	}

	return nil
}

// PublishDepth publishes depth data to Redis using protobuf
func (p *ProtobufPublisher) PublishDepth(ctx context.Context, depth *models.DepthSnapshot) error {
	// Parse bids and asks from JSON string
	// Note: This assumes the Bids/Asks fields contain JSON strings
	// You might need to adjust this based on your actual data format
	bids, err := parsePriceLevels(depth.Bids)
	if err != nil {
		return fmt.Errorf("failed to parse bids: %w", err)
	}

	asks, err := parsePriceLevels(depth.Asks)
	if err != nil {
		return fmt.Errorf("failed to parse asks: %w", err)
	}

	// Create protobuf depth data
	depthData := &binanceProto.DepthData{
		LastUpdateId: depth.LastUpdateID,
		Bids:         bids,
		Asks:         asks,
	}

	// Create live data message
	liveData := &binanceProto.LiveData{
		Type:      binanceProto.DataType_DATA_TYPE_DEPTH,
		Symbol:    depth.Symbol,
		Timestamp: depth.Timestamp,
		Data: &binanceProto.LiveData_Depth{
			Depth: depthData,
		},
	}

	// Publish to channel
	channel := fmt.Sprintf("binance:depth:%s", depth.Symbol)
	if err := p.redis.PublishProtobuf(ctx, channel, liveData); err != nil {
		return fmt.Errorf("failed to publish depth: %w", err)
	}

	// Cache in Redis
	key := fmt.Sprintf("binance:latest:depth:%s", depth.Symbol)
	if err := p.redis.SetProtobuf(ctx, key, liveData, 0); err != nil {
		p.logger.Warn("Failed to cache depth in Redis",
			zap.String("symbol", depth.Symbol),
			zap.Error(err),
		)
	}

	return nil
}

// PublishTrade publishes trade data to Redis using protobuf
func (p *ProtobufPublisher) PublishTrade(ctx context.Context, trade *models.Trade) error {
	// Create protobuf trade data
	tradeData := &binanceProto.TradeData{
		TradeId:       trade.TradeID,
		Price:         trade.Price,
		Quantity:      trade.Quantity,
		QuoteQuantity: trade.QuoteQuantity,
		IsBuyerMaker:  trade.IsBuyerMaker,
	}

	// Create live data message
	liveData := &binanceProto.LiveData{
		Type:      binanceProto.DataType_DATA_TYPE_TRADE,
		Symbol:    trade.Symbol,
		Timestamp: trade.Timestamp,
		Data: &binanceProto.LiveData_Trade{
			Trade: tradeData,
		},
	}

	// Publish to channel
	channel := fmt.Sprintf("binance:trade:%s", trade.Symbol)
	if err := p.redis.PublishProtobuf(ctx, channel, liveData); err != nil {
		return fmt.Errorf("failed to publish trade: %w", err)
	}

	return nil
}

// PublishAllSymbols publishes the list of all active symbols using protobuf
func (p *ProtobufPublisher) PublishAllSymbols(ctx context.Context, symbols []models.Symbol) error {
	symbolList := make([]string, len(symbols))
	for i, s := range symbols {
		symbolList[i] = s.Symbol
	}

	// Create protobuf symbol list
	symbolListData := &binanceProto.SymbolList{
		Symbols:   symbolList,
		Timestamp: 0, // You might want to set this to current timestamp
	}

	key := "binance:symbols:active"
	if err := p.redis.SetProtobuf(ctx, key, symbolListData, 0); err != nil {
		return fmt.Errorf("failed to publish symbols: %w", err)
	}

	return nil
}

// Helper function to parse price levels from JSON string
// This is a placeholder - you'll need to implement based on your actual data format
func parsePriceLevels(jsonData string) ([]*binanceProto.PriceLevel, error) {
	// This is a simplified implementation
	// You'll need to parse the actual JSON format of your bids/asks
	// For now, return empty slice
	return []*binanceProto.PriceLevel{}, nil
}
