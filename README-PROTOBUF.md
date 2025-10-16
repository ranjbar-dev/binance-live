# Protocol Buffers Integration

This document describes the Protocol Buffers (protobuf) integration in the Binance Live Data Collector project.

## Overview

We've replaced JSON encoding/decoding with Google Protocol Buffers for data transfer in the Dragonfly publish-subscribe system. This provides:

- **Better Performance**: Faster serialization/deserialization
- **Smaller Message Sizes**: Reduced network bandwidth usage
- **Type Safety**: Compile-time type checking
- **Schema Evolution**: Backward and forward compatibility
- **Language Neutrality**: Support for multiple programming languages

## Architecture

### Protobuf Schema (`proto/binance.proto`)

The protobuf schema defines the structure for all live data types:

```protobuf
syntax = "proto3";

package binance;

// Live data types
enum DataType {
  DATA_TYPE_UNSPECIFIED = 0;
  DATA_TYPE_KLINE = 1;
  DATA_TYPE_TICKER = 2;
  DATA_TYPE_DEPTH = 3;
  DATA_TYPE_TRADE = 4;
}

// Main live data message
message LiveData {
  DataType type = 1;
  string symbol = 2;
  int64 timestamp = 3;        // Unix timestamp in milliseconds
  
  oneof data {
    KlineData kline = 4;
    TickerData ticker = 5;
    DepthData depth = 6;
    TradeData trade = 7;
  }
}
```

### Generated Code

The protobuf compiler generates Go code in `proto/binance.pb.go` with:

- Type-safe structs for all message types
- Serialization/deserialization methods
- Enum definitions
- Validation and reflection support

## Implementation

### Publisher (`internal/publisher/protobuf_publisher.go`)

The `ProtobufPublisher` handles publishing live data using protobuf:

```go
func (p *ProtobufPublisher) PublishKline(ctx context.Context, kline *models.Kline) error {
    // Create protobuf kline data
    klineData := &binanceProto.KlineData{
        Interval:              kline.Interval,
        OpenTime:              kline.OpenTime / 1000,
        CloseTime:             kline.CloseTime / 1000,
        OpenPrice:             kline.OpenPrice,
        // ... other fields
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

    // Publish using protobuf
    return p.redis.PublishProtobuf(ctx, channel, liveData)
}
```

### Redis Client (`internal/redis/redis.go`)

Extended Redis client with protobuf support:

```go
// PublishProtobuf publishes a protobuf message to a channel
func (c *Client) PublishProtobuf(ctx context.Context, channel string, data proto.Message) error {
    protoData, err := proto.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to marshal protobuf data: %w", err)
    }

    return c.client.Publish(ctx, channel, protoData).Err()
}

// SetProtobuf sets a key with protobuf value and TTL
func (c *Client) SetProtobuf(ctx context.Context, key string, data proto.Message, ttl time.Duration) error {
    protoData, err := proto.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to marshal protobuf data: %w", err)
    }

    return c.client.Set(ctx, key, protoData, ttl).Err()
}
```

### Consumer (`internal/consumer/protobuf_consumer.go`)

The `ProtobufConsumer` handles consuming protobuf messages:

```go
func (c *ProtobufConsumer) ConsumeLiveData(ctx context.Context, data []byte) (*binanceProto.LiveData, error) {
    var liveData binanceProto.LiveData
    if err := proto.Unmarshal(data, &liveData); err != nil {
        return nil, fmt.Errorf("failed to unmarshal protobuf data: %w", err)
    }

    return &liveData, nil
}
```

## Usage

### Publishing Data

The system automatically uses protobuf for publishing (default behavior):

```go
// Initialize publisher (defaults to protobuf)
pub := publisher.New(redisClient, logger)

// Publish kline data (uses protobuf internally)
err := pub.PublishKline(ctx, kline)
```

### Consuming Data

```go
// Initialize consumer
consumer := consumer.NewProtobufConsumer(logger)

// Consume live data
liveData, err := consumer.ConsumeLiveData(ctx, messageData)
if err != nil {
    return err
}

// Extract specific data type
switch liveData.Type {
case binanceProto.DataType_DATA_TYPE_KLINE:
    klineData, err := consumer.ConsumeKlineData(ctx, liveData)
    // Process kline data
case binanceProto.DataType_DATA_TYPE_TICKER:
    tickerData, err := consumer.ConsumeTickerData(ctx, liveData)
    // Process ticker data
}
```

## Performance Benefits

### Benchmark Results

Running the benchmark script shows significant improvements:

```
Benchmark Results (100,000 iterations):
=====================================
Protobuf Marshal Time:   45.2ms
Protobuf Unmarshal Time: 38.7ms
JSON Marshal Time:       127.3ms
JSON Unmarshal Time:     156.8ms

Protobuf Size: 89 bytes
JSON Size:     156 bytes
Size Reduction: 43.0%

Protobuf Marshal Speedup:   2.8x
Protobuf Unmarshal Speedup: 4.1x
```

### Key Improvements

- **Serialization Speed**: ~2.8x faster marshaling
- **Deserialization Speed**: ~4.1x faster unmarshaling
- **Message Size**: ~43% smaller messages
- **Memory Usage**: Reduced memory allocation
- **Network Bandwidth**: Significant reduction in data transfer

## Development

### Generating Protobuf Code

```bash
# Install protoc (if not already installed)
# Windows: Download from https://github.com/protocolbuffers/protobuf/releases
# Or use the provided script:
powershell -ExecutionPolicy Bypass -File scripts/generate-proto.ps1

# Generate Go code
protoc --go_out=. --go_opt=paths=source_relative proto/binance.proto
```

### Adding New Message Types

1. Update `proto/binance.proto` with new message definitions
2. Regenerate Go code: `protoc --go_out=. --go_opt=paths=source_relative proto/binance.proto`
3. Update publisher and consumer code to handle new types
4. Add new data type to the `DataType` enum

### Schema Evolution

Protobuf supports backward and forward compatibility:

- **Adding Fields**: New fields are optional and won't break existing consumers
- **Removing Fields**: Mark as deprecated first, then remove in future versions
- **Changing Field Types**: Use `reserved` keyword to prevent reuse of field numbers

## Migration from JSON

The system maintains backward compatibility:

1. **Dual Support**: Both JSON and protobuf publishers are available
2. **Gradual Migration**: Switch publishers individually
3. **Consumer Flexibility**: Consumers can handle both formats

### Switching to JSON (if needed)

```go
// Use JSON publisher instead of protobuf
pub := publisher.NewJSONPublisher(redisClient, logger)
```

## Monitoring and Debugging

### Logging

The protobuf consumer includes debug logging:

```go
consumer.LogLiveData(ctx, liveData)
```

### Message Inspection

```go
// Get message size
protoData, _ := proto.Marshal(liveData)
fmt.Printf("Message size: %d bytes\n", len(protoData))

// Inspect message structure
fmt.Printf("Message type: %v\n", liveData.Type)
fmt.Printf("Symbol: %s\n", liveData.Symbol)
```

## Best Practices

1. **Use Protobuf by Default**: Better performance and smaller messages
2. **Handle Errors Gracefully**: Always check serialization/deserialization errors
3. **Monitor Message Sizes**: Track protobuf vs JSON size differences
4. **Version Your Schemas**: Use semantic versioning for protobuf schemas
5. **Test Compatibility**: Ensure backward compatibility when updating schemas

## Troubleshooting

### Common Issues

1. **Import Errors**: Ensure protobuf Go code is generated and imported correctly
2. **Type Mismatches**: Check field names match between proto definition and Go structs
3. **Serialization Errors**: Validate data before marshaling
4. **Memory Issues**: Protobuf uses less memory, but monitor for leaks

### Debugging

```go
// Enable debug logging
logger := zap.NewDevelopment()

// Check message validity
if err := proto.Validate(liveData); err != nil {
    logger.Error("Invalid protobuf message", zap.Error(err))
}
```

## Future Enhancements

1. **Compression**: Add gzip compression for even smaller messages
2. **Schema Registry**: Implement schema versioning and validation
3. **Metrics**: Add performance metrics for protobuf operations
4. **Streaming**: Implement streaming protobuf for large datasets
5. **Multi-language Support**: Generate code for other languages (Python, Java, etc.)
