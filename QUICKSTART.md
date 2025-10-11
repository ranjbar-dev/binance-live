# Quick Start Guide

Get up and running with Binance Live Data Collector in 5 minutes!

## Prerequisites

- Docker and Docker Compose installed
- At least 2GB of free RAM
- Internet connection for Binance API

## Quick Start (Docker)

### 1. Clone and Setup

```bash
cd binance-live

# Create environment file from template
cp .env.template .env

# Edit .env and set a secure password
# vim .env  # or use your preferred editor
```

### 2. Start All Services

```bash
# Start with just the core services
docker-compose up -d

# OR start with monitoring tools (pgAdmin + Redis Commander)
docker-compose --profile tools up -d
```

### 3. View Logs

```bash
# View all logs
docker-compose logs -f

# View just the application logs
docker-compose logs -f app

# View last 100 lines
docker-compose logs --tail=100 app
```

### 4. Check Everything is Running

```bash
docker-compose ps
```

You should see:

- ✓ timescaledb (healthy)
- ✓ redis (healthy)
- ✓ app (running)

### 5. Verify Data Collection

Connect to Redis and subscribe to data:

```bash
# Subscribe to Bitcoin ticker updates
docker-compose exec redis redis-cli SUBSCRIBE binance:ticker:BTCUSDT

# In another terminal, get the latest ticker
docker-compose exec redis redis-cli GET binance:latest:ticker:BTCUSDT
```

### 6. Query Historical Data

Connect to the database:

```bash
# Using docker exec
docker-compose exec timescaledb psql -U postgres -d binance_data

# Run a query
SELECT symbol, interval, open_time, close_price
FROM klines
WHERE symbol = 'BTCUSDT' AND interval = '1h'
ORDER BY open_time DESC
LIMIT 10;
```

## Monitoring Tools (Optional)

If you started with `--profile tools`:

### pgAdmin (Database GUI)

- URL: http://localhost:5050
- Email: admin@admin.com
- Password: admin123

Add a new server connection:

- Host: timescaledb
- Port: 5432
- Username: postgres
- Password: (from your .env file)
- Database: binance_data

### Redis Commander (Redis GUI)

- URL: http://localhost:8081

## Managing Symbols

### View Active Symbols

```bash
docker-compose exec timescaledb psql -U postgres -d binance_data -c "SELECT * FROM symbols WHERE is_active = true;"
```

### Add a New Symbol

```bash
docker-compose exec timescaledb psql -U postgres -d binance_data -c "
INSERT INTO symbols (symbol, base_asset, quote_asset, status, is_active)
VALUES ('ETHUSDT', 'ETH', 'USDT', 'TRADING', true)
ON CONFLICT (symbol) DO UPDATE SET is_active = true;
"

# Restart the app to pick up the new symbol
docker-compose restart app
```

### Disable a Symbol

```bash
docker-compose exec timescaledb psql -U postgres -d binance_data -c "
UPDATE symbols SET is_active = false WHERE symbol = 'DOGEUSDT';
"

# Restart the app
docker-compose restart app
```

## Common Commands

### Stop Everything

```bash
docker-compose down
```

### Stop and Remove All Data

```bash
docker-compose down -v
```

### Restart Just the App

```bash
docker-compose restart app
```

### View Resource Usage

```bash
docker stats
```

### Update Configuration

1. Edit `config/config.yaml`
2. Restart the app:

```bash
docker-compose restart app
```

## Useful SQL Queries

### Get Last 24 Hours of BTC Prices (1h intervals)

```sql
SELECT
    open_time,
    open_price,
    high_price,
    low_price,
    close_price,
    volume
FROM klines
WHERE symbol = 'BTCUSDT'
  AND interval = '1h'
  AND open_time >= NOW() - INTERVAL '24 hours'
ORDER BY open_time DESC;
```

### Calculate Average Price by Day

```sql
SELECT
    time_bucket('1 day', open_time) AS day,
    symbol,
    AVG(close_price) as avg_price,
    MAX(high_price) as max_price,
    MIN(low_price) as min_price
FROM klines
WHERE symbol = 'BTCUSDT'
  AND interval = '1h'
  AND open_time >= NOW() - INTERVAL '7 days'
GROUP BY day, symbol
ORDER BY day DESC;
```

### Check Latest Ticker Data

```sql
SELECT
    symbol,
    timestamp,
    price,
    price_change_percent_24h,
    volume_24h
FROM tickers
WHERE timestamp >= NOW() - INTERVAL '1 hour'
ORDER BY timestamp DESC
LIMIT 20;
```

## Troubleshooting

### App Won't Start

1. Check logs:

   ```bash
   docker-compose logs app
   ```

2. Check database connectivity:

   ```bash
   docker-compose exec timescaledb pg_isready -U postgres
   ```

3. Check Redis connectivity:
   ```bash
   docker-compose exec redis redis-cli ping
   ```

### No Data Being Collected

1. Check if symbols are active:

   ```sql
   SELECT * FROM symbols WHERE is_active = true;
   ```

2. Check sync status:

   ```sql
   SELECT * FROM sync_status ORDER BY updated_at DESC;
   ```

3. Verify Binance API connectivity:
   ```bash
   curl https://api.binance.com/api/v3/ping
   ```

### High Memory Usage

1. Reduce number of active symbols
2. Reduce kline intervals in config.yaml
3. Add data retention policies (see README.md)

## Next Steps

1. Read the full [README.md](README.md) for detailed documentation
2. Customize `config/config.yaml` for your needs
3. Set up retention policies for old data
4. Build your own consumers to read from Redis streams
5. Create custom queries and analytics

## Support

For issues or questions:

1. Check the logs: `docker-compose logs app`
2. Review the full README.md
3. Check Binance API status: https://www.binance.com/en/support

## Important Notes

- The first sync may take several minutes depending on `max_sync_hours` config
- Data retention is important - TimescaleDB hypertables can grow large
- Always use strong passwords in production
- This is for educational/personal use - follow Binance's terms of service

---

**Happy collecting!** 🚀
