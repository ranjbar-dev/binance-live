package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/binance-live/internal/binance"
	"github.com/binance-live/internal/models"
	"github.com/binance-live/internal/publisher"
	"github.com/binance-live/internal/repository"
	"go.uber.org/zap"
)

// StreamService handles real-time data streaming from Binance WebSocket
type StreamService struct {
	binanceClient  *binance.Client
	klineRepo      *repository.KlineRepository
	tickerRepo     *repository.TickerRepository
	syncStatusRepo *repository.SyncStatusRepository
	publisher      publisher.Publisher
	logger         *zap.Logger
}

// NewStreamService creates a new stream service
func NewStreamService(
	binanceClient *binance.Client,
	klineRepo *repository.KlineRepository,
	tickerRepo *repository.TickerRepository,
	syncStatusRepo *repository.SyncStatusRepository,
	pub *publisher.Publisher,
	logger *zap.Logger,
) *StreamService {
	return &StreamService{
		binanceClient:  binanceClient,
		klineRepo:      klineRepo,
		tickerRepo:     tickerRepo,
		syncStatusRepo: syncStatusRepo,
		publisher:      *pub,
		logger:         logger,
	}
}

// Start starts streaming live data for the given symbols
func (s *StreamService) Start(ctx context.Context, symbols []models.Symbol) error {
	if len(symbols) == 0 {
		return fmt.Errorf("no symbols provided for streaming")
	}

	// Build stream names
	symbolNames := make([]string, len(symbols))
	for i, sym := range symbols {
		symbolNames[i] = sym.Symbol
	}

	streams := binance.BuildStreamNames(symbolNames, s.binanceClient.Config.KlineIntervals)

	s.logger.Info("Starting WebSocket streams",
		zap.Int("symbol_count", len(symbolNames)),
		zap.Int("stream_count", len(streams)),
	)

	// Register handlers for each stream
	for _, stream := range streams {
		s.registerStreamHandler(stream)
	}

	// Start WebSocket client
	go func() {
		if err := s.binanceClient.WebSocket.Start(ctx, streams); err != nil {
			s.logger.Error("WebSocket client error", zap.Error(err))
		}
	}()

	s.logger.Info("WebSocket streams started successfully")
	return nil
}

// registerStreamHandler registers a handler for a specific stream
func (s *StreamService) registerStreamHandler(stream string) {
	symbol, streamType, interval := binance.GetStreamName(stream)

	switch streamType {
	case "kline":
		s.binanceClient.WebSocket.RegisterHandler(stream, func(message []byte) error {
			return s.handleKlineEvent(message, symbol, interval)
		})
	case "ticker":
		s.binanceClient.WebSocket.RegisterHandler(stream, func(message []byte) error {
			return s.handleTickerEvent(message, symbol)
		})
	case "depth":
		s.binanceClient.WebSocket.RegisterHandler(stream, func(message []byte) error {
			return s.handleDepthEvent(message, symbol)
		})
	case "aggTrade":
		s.binanceClient.WebSocket.RegisterHandler(stream, func(message []byte) error {
			return s.handleTradeEvent(message, symbol)
		})
	}
}

// handleKlineEvent handles kline WebSocket events
func (s *StreamService) handleKlineEvent(message []byte, symbol, interval string) error {
	var event binance.WSKlineEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal kline event: %w", err)
	}

	// Only process closed klines
	if !event.Kline.IsClosed {
		return nil
	}

	// Convert to model
	kline, err := s.convertWSKlineToModel(&event, symbol, interval)
	if err != nil {
		return fmt.Errorf("failed to convert kline: %w", err)
	}

	// Store in database
	ctx := context.Background()
	if err := s.klineRepo.Insert(ctx, kline); err != nil {
		s.logger.Error("Failed to insert kline", zap.Error(err))
	}

	// Publish to Redis
	if err := s.publisher.PublishKline(ctx, kline); err != nil {
		s.logger.Error("Failed to publish kline", zap.Error(err))
	}

	// Update sync status
	if err := s.syncStatusRepo.UpdateLastDataTime(ctx, symbol, "kline", &interval, kline.OpenTime); err != nil {
		s.logger.Warn("Failed to update sync status", zap.Error(err))
	}

	return nil
}

// handleTickerEvent handles ticker WebSocket events
func (s *StreamService) handleTickerEvent(message []byte, symbol string) error {
	var event binance.WSTickerEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal ticker event: %w", err)
	}

	// Convert to model
	ticker, err := s.convertWSTickerToModel(&event)
	if err != nil {
		return fmt.Errorf("failed to convert ticker: %w", err)
	}

	// Store in database
	ctx := context.Background()
	if err := s.tickerRepo.Insert(ctx, ticker); err != nil {
		s.logger.Error("Failed to insert ticker", zap.Error(err))
	}

	// Publish to Redis
	if err := s.publisher.PublishTicker(ctx, ticker); err != nil {
		s.logger.Error("Failed to publish ticker", zap.Error(err))
	}

	return nil
}

// handleDepthEvent handles depth WebSocket events
func (s *StreamService) handleDepthEvent(message []byte, symbol string) error {
	var event binance.WSDepthEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal depth event: %w", err)
	}

	// Convert to model
	depth, err := s.convertWSDepthToModel(&event)
	if err != nil {
		return fmt.Errorf("failed to convert depth: %w", err)
	}

	// Publish to Redis (depth is typically not stored in DB due to size)
	ctx := context.Background()
	if err := s.publisher.PublishDepth(ctx, depth); err != nil {
		s.logger.Error("Failed to publish depth", zap.Error(err))
	}

	return nil
}

// handleTradeEvent handles trade WebSocket events
func (s *StreamService) handleTradeEvent(message []byte, symbol string) error {
	var event binance.WSAggTradeEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal trade event: %w", err)
	}

	// Convert to model
	trade, err := s.convertWSTradeToModel(&event)
	if err != nil {
		return fmt.Errorf("failed to convert trade: %w", err)
	}

	// Publish to Redis
	ctx := context.Background()
	if err := s.publisher.PublishTrade(ctx, trade); err != nil {
		s.logger.Error("Failed to publish trade", zap.Error(err))
	}

	return nil
}

// Convert WebSocket events to models

func (s *StreamService) convertWSKlineToModel(event *binance.WSKlineEvent, symbol, interval string) (*models.Kline, error) {
	openPrice, _ := strconv.ParseFloat(event.Kline.Open, 64)
	highPrice, _ := strconv.ParseFloat(event.Kline.High, 64)
	lowPrice, _ := strconv.ParseFloat(event.Kline.Low, 64)
	closePrice, _ := strconv.ParseFloat(event.Kline.Close, 64)
	volume, _ := strconv.ParseFloat(event.Kline.Volume, 64)
	quoteVolume, _ := strconv.ParseFloat(event.Kline.QuoteVolume, 64)
	takerBuyVolume, _ := strconv.ParseFloat(event.Kline.TakerBuyBaseAssetVolume, 64)
	takerBuyQuoteVolume, _ := strconv.ParseFloat(event.Kline.TakerBuyQuoteAssetVolume, 64)

	return &models.Kline{
		Symbol:              symbol,
		Interval:            interval,
		OpenTime:            event.Kline.StartTime,
		CloseTime:           event.Kline.EndTime,
		OpenPrice:           openPrice,
		HighPrice:           highPrice,
		LowPrice:            lowPrice,
		ClosePrice:          closePrice,
		Volume:              volume,
		QuoteVolume:         quoteVolume,
		TradesCount:         event.Kline.NumberOfTrades,
		TakerBuyVolume:      takerBuyVolume,
		TakerBuyQuoteVolume: takerBuyQuoteVolume,
		CreatedAt:           time.Now().UnixMilli(),
	}, nil
}

func (s *StreamService) convertWSTickerToModel(event *binance.WSTickerEvent) (*models.Ticker, error) {
	price, _ := strconv.ParseFloat(event.LastPrice, 64)
	bidPrice, _ := strconv.ParseFloat(event.BidPrice, 64)
	bidQty, _ := strconv.ParseFloat(event.BidQty, 64)
	askPrice, _ := strconv.ParseFloat(event.AskPrice, 64)
	askQty, _ := strconv.ParseFloat(event.AskQty, 64)
	volume24h, _ := strconv.ParseFloat(event.Volume, 64)
	quoteVolume24h, _ := strconv.ParseFloat(event.QuoteVolume, 64)
	priceChange24h, _ := strconv.ParseFloat(event.PriceChange, 64)
	priceChangePercent24h, _ := strconv.ParseFloat(event.PriceChangePercent, 64)
	high24h, _ := strconv.ParseFloat(event.HighPrice, 64)
	low24h, _ := strconv.ParseFloat(event.LowPrice, 64)

	return &models.Ticker{
		Symbol:                event.Symbol,
		Timestamp:             event.EventTime,
		Price:                 price,
		BidPrice:              &bidPrice,
		BidQty:                &bidQty,
		AskPrice:              &askPrice,
		AskQty:                &askQty,
		Volume24h:             &volume24h,
		QuoteVolume24h:        &quoteVolume24h,
		PriceChange24h:        &priceChange24h,
		PriceChangePercent24h: &priceChangePercent24h,
		High24h:               &high24h,
		Low24h:                &low24h,
		TradesCount24h:        &event.Count,
		CreatedAt:             time.Now().UnixMilli(),
	}, nil
}

func (s *StreamService) convertWSDepthToModel(event *binance.WSDepthEvent) (*models.DepthSnapshot, error) {
	bidsJSON, _ := json.Marshal(event.Bids)
	asksJSON, _ := json.Marshal(event.Asks)

	return &models.DepthSnapshot{
		Symbol:       event.Symbol,
		Timestamp:    event.EventTime,
		LastUpdateID: event.FinalUpdateID,
		Bids:         string(bidsJSON),
		Asks:         string(asksJSON),
		CreatedAt:    time.Now().UnixMilli(),
	}, nil
}

func (s *StreamService) convertWSTradeToModel(event *binance.WSAggTradeEvent) (*models.Trade, error) {
	price, _ := strconv.ParseFloat(event.Price, 64)
	quantity, _ := strconv.ParseFloat(event.Quantity, 64)
	quoteQuantity := price * quantity

	return &models.Trade{
		Symbol:        event.Symbol,
		TradeID:       event.AggTradeID,
		Timestamp:     event.TradeTime,
		Price:         price,
		Quantity:      quantity,
		QuoteQuantity: quoteQuantity,
		IsBuyerMaker:  event.IsBuyerMaker,
		CreatedAt:     time.Now().UnixMilli(),
	}, nil
}

// Stop stops the stream service
func (s *StreamService) Stop() error {
	return s.binanceClient.WebSocket.Close()
}
