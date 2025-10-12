-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Table to store aggregated trade data (quotes)
CREATE TABLE IF NOT EXISTS trades (
    id BIGSERIAL,
    symbol VARCHAR(20) NOT NULL,
    trade_id BIGINT NOT NULL,
    timestamp BIGINT NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    quote_quantity DECIMAL(20, 8) NOT NULL,
    is_buyer_maker BOOLEAN NOT NULL,
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()) * 1000,
    PRIMARY KEY (id, timestamp),
    UNIQUE (symbol, trade_id, timestamp)
);

-- Convert to hypertable
SELECT create_hypertable('trades', 'timestamp', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol, timestamp DESC);
