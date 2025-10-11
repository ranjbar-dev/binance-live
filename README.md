# Binance Live Data Collector

A high-performance, production-ready Go application that connects to Binance's WebSocket and REST APIs to collect real-time cryptocurrency market data. The system stores historical data in TimescaleDB (optimized PostgreSQL for time-series data) and publishes live data streams to Redis for real-time consumption.

## 🚀 Features

- **Real-time Data Collection**: WebSocket connections to Binance for live market data
- **Multiple Data Types**:
  - Klines/Candlesticks (multiple intervals)
  - 24hr Ticker Statistics
  - Order Book Depth
  - Aggregated Trades
- **Historical Data Storage**: TimescaleDB for efficient time-series data storage
- **Live Data Publishing**: Redis pub/sub for real-time data distribution
- **Smart Data Synchronization**: Automatically fetches missing data after downtime
- **Configurable Symbol Pairs**: Database-driven symbol management
- **Production-Ready**: 
  - Structured logging
  - Health checks
  - Graceful shutdown
  - Auto-reconnection
  - Rate limiting
- **Docker Support**: Complete Docker Compose setup with all dependencies
- **Well-Structured**: Clean architecture with separation of concerns

## 📋 Architecture

```
┌─────────────┐
│   Binance   │
│   REST API  │◄────┐
└─────────────┘     │
                    │
┌─────────────┐     │         ┌──────────────┐
│   Binance   │     │         │              │
│  WebSocket  │◄────┼─────────│  Go Backend  │
└─────────────┘     │         │              │
                    │         └───────┬──────┘
                    │                 │
                    └─────────────────┤
                                      │
                    ┌─────────────────┼──────────────┐
                    │                 │              │
                    ▼                 ▼              ▼
            ┌──────────────┐  ┌─────────────┐  ┌────────┐
            │ TimescaleDB  │  │    Redis    │  │  Logs  │
            │ (Historical) │  │   (Live)    │  │        │
            └──────────────┘  └─────────────┘  └────────┘
```

## 📦 Project Structure

```
binance-live/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── binance/
│   │   ├── client.go              # Main Binance client
│   │   ├── rest.go                # REST API client
│   │   ├── websocket.go           # WebSocket client
│   │   └── types.go               # API response types
│   ├── config/
│   │   └── config.go              # Configuration management (Viper)
│   ├── database/
│   │   └── postgres.go            # Database connection
│   ├── logger/
│   │   └── logger.go              # Structured logging (Zap)
│   ├── models/
│   │   └── models.go              # Data models
│   ├── publisher/
│   │   └── redis_publisher.go     # Redis publishing logic
│   ├── redis/
│   │   └── redis.go               # Redis client
│   ├── repository/
│   │   ├── symbol.go              # Symbol repository
│   │   ├── kline.go               # Kline repository
│   │   ├── ticker.go              # Ticker repository
│   │   └── sync_status.go         # Sync status repository
│   └── service/
│       ├── data_sync.go           # Historical data sync service
│       └── stream.go              # Live streaming service
├── config/
│   └── config.yaml                # Application configuration
├── migrations/
│   └── 001_init.sql               # Database schema
├── docker-compose.yml             # Docker Compose configuration
├── Dockerfile                     # Application Dockerfile
├── Makefile                       # Build automation
├── go.mod                         # Go dependencies
└── README.md                      # This file
```

## 🛠️ Prerequisites

- Docker and Docker Compose (recommended)
- Or: Go 1.21+, PostgreSQL 15+, Redis 7+

## 🚀 Quick Start

### Using Docker (Recommended)

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd binance-live
   ```

2. **Create environment file**:
   ```bash
   make init
   # Edit .env with your configuration
   ```

3. **Start all services**:
   ```bash
   make docker-up
   ```

4. **View logs**:
   ```bash
   make docker-logs-app
   ```

5. **Stop services**:
   ```bash
   make docker-down
   ```

### Using Docker with Monitoring Tools

Start with pgAdmin and Redis Commander for monitoring:

```bash
make dev
```

- pgAdmin: http://localhost:5050
- Redis Commander: http://localhost:8081

### Manual Setup (Without Docker)

1. **Install dependencies**:
   ```bash
   make deps
   ```

2. **Set up PostgreSQL with TimescaleDB**:
   ```bash
   # Install TimescaleDB extension
   psql -U postgres -c "CREATE DATABASE binance_data;"
   psql -U postgres -d binance_data -c "CREATE EXTENSION IF NOT EXISTS timescaledb;"
   
   # Run migrations
   psql -U postgres -d binance_data -f migrations/001_init.sql
   ```

3. **Configure environment**:
   ```bash
   export POSTGRES_HOST=localhost
   export POSTGRES_PASSWORD=your_password
   export REDIS_HOST=localhost
   ```

4. **Run the application**:
   ```bash
   make run
   ```

## ⚙️ Configuration

Configuration is managed through `config/config.yaml` and environment variables. Environment variables take precedence.

### Key Configuration Options

```yaml
app:
  name: "binance-live-collector"
  environment: "production"
  log_level: "info"

binance:
  # Kline intervals to collect
  kline_intervals:
    - "1m"
    - "5m"
    - "15m"
    - "1h"
    - "4h"
    - "1d"

database:
  host: "${POSTGRES_HOST:localhost}"
  port: 5432
  # ... more options

redis:
  host: "${REDIS_HOST:localhost}"
  port: 6379
  live_data_ttl: 60  # seconds

sync:
  enabled: true
  max_sync_hours: 24     # How far back to sync
  batch_size: 1000
  workers: 5             # Concurrent sync workers
```

## 📊 Database Schema

### Symbol Management

Symbols are stored in the `symbols` table. Only symbols marked as `is_active = true` will be monitored.

```sql
-- View active symbols
SELECT * FROM symbols WHERE is_active = true;

-- Add a new symbol
INSERT INTO symbols (symbol, base_asset, quote_asset, status, is_active)
VALUES ('ETHUSDT', 'ETH', 'USDT', 'TRADING', true);

-- Disable a symbol
UPDATE symbols SET is_active = false WHERE symbol = 'BTCUSDT';
```

### Data Tables

- **klines**: Candlestick/OHLCV data (hypertable)
- **tickers**: 24hr ticker statistics (hypertable)
- **depth_snapshots**: Order book depth snapshots (hypertable)
- **trades**: Aggregated trade data (hypertable)
- **sync_status**: Tracks synchronization status

## 📡 Redis Data Streams

### Published Channels

Live data is published to the following Redis channels:

- `binance:kline:{symbol}:{interval}` - Kline updates
- `binance:ticker:{symbol}` - Ticker updates
- `binance:depth:{symbol}` - Order book updates
- `binance:trade:{symbol}` - Trade updates

### Cached Data

Latest data is also cached in Redis:

- `binance:latest:kline:{symbol}:{interval}`
- `binance:latest:ticker:{symbol}`
- `binance:latest:depth:{symbol}`
- `binance:symbols:active` - List of active symbols

### Subscribing to Data

Example using Redis CLI:

```bash
# Subscribe to Bitcoin ticker updates
redis-cli SUBSCRIBE binance:ticker:BTCUSDT

# Subscribe to Ethereum 1-minute klines
redis-cli SUBSCRIBE binance:kline:ETHUSDT:1m

# Get latest cached ticker
redis-cli GET binance:latest:ticker:BTCUSDT
```

Example in your application:

```python
import redis
import json

r = redis.Redis(host='localhost', port=6379, decode_responses=True)
pubsub = r.pubsub()

# Subscribe to all Bitcoin streams
pubsub.psubscribe('binance:*:BTCUSDT*')

for message in pubsub.listen():
    if message['type'] == 'pmessage':
        data = json.loads(message['data'])
        print(f"Symbol: {data['symbol']}, Type: {data['type']}")
```

## 🔄 Data Synchronization

The application automatically handles downtime recovery:

1. **On Startup**: Checks `sync_status` table for last sync time
2. **Calculates Gap**: Determines missing data based on `max_sync_hours` configuration
3. **Fetches Historical Data**: Uses REST API to fetch missing data in batches
4. **Resumes Live Streaming**: Continues with WebSocket for real-time updates

### Sync Status Tracking

```sql
-- View sync status for all symbols
SELECT * FROM sync_status ORDER BY symbol, data_type;

-- Check last sync time for a specific symbol
SELECT * FROM sync_status 
WHERE symbol = 'BTCUSDT' AND data_type = 'kline' AND interval = '1m';
```

## 🔍 Monitoring & Logging

### Application Logs

```bash
# View live logs
docker-compose logs -f app

# View logs with timestamps
docker-compose logs -f --timestamps app

# View last 100 lines
docker-compose logs --tail=100 app
```

### Health Checks

The Docker Compose setup includes health checks for all services:

```bash
# Check service health
docker-compose ps
```

### Database Monitoring

Connect to pgAdmin (when running with `make dev`):
- URL: http://localhost:5050
- Email: admin@admin.com
- Password: admin123

### Redis Monitoring

Connect to Redis Commander (when running with `make dev`):
- URL: http://localhost:8081

## 📈 Performance Considerations

### TimescaleDB Optimizations

- **Hypertables**: Automatic partitioning by time
- **Compression**: Enable compression for older data
- **Retention Policies**: Automatically drop old data

```sql
-- Enable compression (after 7 days)
ALTER TABLE klines SET (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'symbol,interval'
);

SELECT add_compression_policy('klines', INTERVAL '7 days');

-- Add retention policy (keep 90 days)
SELECT add_retention_policy('klines', INTERVAL '90 days');
```

### Rate Limiting

The application implements rate limiting for Binance API:
- REST API: Configurable requests per minute
- WebSocket: Automatic reconnection with backoff

### Resource Usage

Typical resource usage:
- **Application**: 100-200 MB RAM
- **TimescaleDB**: 500 MB+ (depends on data retention)
- **Redis**: 50-100 MB

## 🛠️ Development

### Building

```bash
# Build binary
make build

# Run tests
make test

# Format code
make fmt

# Run linters
make lint
```

### Adding New Symbols

1. **Via SQL**:
   ```sql
   INSERT INTO symbols (symbol, base_asset, quote_asset, status, is_active)
   VALUES ('NEWUSDT', 'NEW', 'USDT', 'TRADING', true);
   ```

2. **Restart application** to pick up new symbols

### Modifying Kline Intervals

Edit `config/config.yaml`:

```yaml
binance:
  kline_intervals:
    - "1m"
    - "5m"
    - "15m"
    - "1h"
    # Add more intervals as needed
```

## 🐛 Troubleshooting

### WebSocket Connection Issues

```bash
# Check if WebSocket can connect
docker-compose logs app | grep "WebSocket"

# Verify network connectivity
docker-compose exec app ping stream.binance.com
```

### Database Connection Issues

```bash
# Check database logs
docker-compose logs timescaledb

# Test connection
docker-compose exec timescaledb psql -U postgres -d binance_data -c "SELECT 1;"
```

### Redis Connection Issues

```bash
# Check Redis logs
docker-compose logs redis

# Test connection
docker-compose exec redis redis-cli ping
```

### High Memory Usage

- Reduce `kline_intervals` in configuration
- Enable TimescaleDB compression
- Add retention policies
- Reduce number of active symbols

## 📝 API Documentation

### Binance API Documentation
- REST API: https://binance-docs.github.io/apidocs/spot/en/
- WebSocket API: https://binance-docs.github.io/apidocs/spot/en/#websocket-market-streams

### Rate Limits

Binance enforces rate limits:
- REST API: 1200 requests per minute
- WebSocket: 300 connections per IP

The application respects these limits through rate limiting and connection pooling.

## 🔒 Security Considerations

1. **Environment Variables**: Never commit sensitive data
2. **Database Credentials**: Use strong passwords
3. **Redis Password**: Set in production environments
4. **Network Security**: Use Docker networks for isolation
5. **API Keys**: Not required for public market data

## 📊 Example Queries

### Get Latest Klines

```sql
SELECT * FROM klines
WHERE symbol = 'BTCUSDT' AND interval = '1h'
ORDER BY open_time DESC
LIMIT 24;
```

### Calculate Average Price

```sql
SELECT 
    symbol,
    interval,
    time_bucket('1 day', open_time) AS day,
    AVG(close_price) as avg_price
FROM klines
WHERE symbol = 'BTCUSDT' AND interval = '1h'
  AND open_time >= NOW() - INTERVAL '7 days'
GROUP BY symbol, interval, day
ORDER BY day DESC;
```

### Get Price Changes

```sql
SELECT 
    symbol,
    timestamp,
    price,
    price_change_percent_24h
FROM tickers
WHERE symbol = 'BTCUSDT'
ORDER BY timestamp DESC
LIMIT 10;
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License.

## 🙏 Acknowledgments

- [Binance API](https://www.binance.com/en/binance-api)
- [TimescaleDB](https://www.timescale.com/)
- [Redis](https://redis.io/)
- [Viper](https://github.com/spf13/viper)
- [Zap](https://github.com/uber-go/zap)

## 📧 Support

For issues, questions, or contributions, please open an issue on GitHub.

---

**Note**: This application is for educational and personal use. Always comply with Binance's Terms of Service and API usage guidelines.

