package consumer

import (
	"context"
	"fmt"

	binanceProto "github.com/binance-live/proto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// ProtobufConsumer handles consuming protobuf messages from Redis
type ProtobufConsumer struct {
	logger *zap.Logger
}

// NewProtobufConsumer creates a new protobuf consumer
func NewProtobufConsumer(logger *zap.Logger) *ProtobufConsumer {
	return &ProtobufConsumer{
		logger: logger,
	}
}

// ConsumeLiveData consumes a protobuf live data message
func (c *ProtobufConsumer) ConsumeLiveData(ctx context.Context, data []byte) (*binanceProto.LiveData, error) {
	var liveData binanceProto.LiveData
	if err := proto.Unmarshal(data, &liveData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf data: %w", err)
	}

	return &liveData, nil
}

// ConsumeKlineData extracts kline data from a live data message
func (c *ProtobufConsumer) ConsumeKlineData(ctx context.Context, liveData *binanceProto.LiveData) (*binanceProto.KlineData, error) {
	if liveData.Type != binanceProto.DataType_DATA_TYPE_KLINE {
		return nil, fmt.Errorf("expected kline data type, got %v", liveData.Type)
	}

	klineData, ok := liveData.Data.(*binanceProto.LiveData_Kline)
	if !ok {
		return nil, fmt.Errorf("invalid kline data format")
	}

	return klineData.Kline, nil
}

// ConsumeTickerData extracts ticker data from a live data message
func (c *ProtobufConsumer) ConsumeTickerData(ctx context.Context, liveData *binanceProto.LiveData) (*binanceProto.TickerData, error) {
	if liveData.Type != binanceProto.DataType_DATA_TYPE_TICKER {
		return nil, fmt.Errorf("expected ticker data type, got %v", liveData.Type)
	}

	tickerData, ok := liveData.Data.(*binanceProto.LiveData_Ticker)
	if !ok {
		return nil, fmt.Errorf("invalid ticker data format")
	}

	return tickerData.Ticker, nil
}

// ConsumeDepthData extracts depth data from a live data message
func (c *ProtobufConsumer) ConsumeDepthData(ctx context.Context, liveData *binanceProto.LiveData) (*binanceProto.DepthData, error) {
	if liveData.Type != binanceProto.DataType_DATA_TYPE_DEPTH {
		return nil, fmt.Errorf("expected depth data type, got %v", liveData.Type)
	}

	depthData, ok := liveData.Data.(*binanceProto.LiveData_Depth)
	if !ok {
		return nil, fmt.Errorf("invalid depth data format")
	}

	return depthData.Depth, nil
}

// ConsumeTradeData extracts trade data from a live data message
func (c *ProtobufConsumer) ConsumeTradeData(ctx context.Context, liveData *binanceProto.LiveData) (*binanceProto.TradeData, error) {
	if liveData.Type != binanceProto.DataType_DATA_TYPE_TRADE {
		return nil, fmt.Errorf("expected trade data type, got %v", liveData.Type)
	}

	tradeData, ok := liveData.Data.(*binanceProto.LiveData_Trade)
	if !ok {
		return nil, fmt.Errorf("invalid trade data format")
	}

	return tradeData.Trade, nil
}

// ConsumeSymbolList consumes a protobuf symbol list message
func (c *ProtobufConsumer) ConsumeSymbolList(ctx context.Context, data []byte) (*binanceProto.SymbolList, error) {
	var symbolList binanceProto.SymbolList
	if err := proto.Unmarshal(data, &symbolList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal symbol list: %w", err)
	}

	return &symbolList, nil
}

// LogLiveData logs live data information for debugging
func (c *ProtobufConsumer) LogLiveData(ctx context.Context, liveData *binanceProto.LiveData) {
	c.logger.Debug("Received live data",
		zap.String("type", liveData.Type.String()),
		zap.String("symbol", liveData.Symbol),
		zap.Int64("timestamp", liveData.Timestamp),
	)
}
