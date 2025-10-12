-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Table to store order book depth data
CREATE TABLE IF NOT EXISTS depth_snapshots (
    id BIGSERIAL,
    symbol VARCHAR(20) NOT NULL,
    timestamp BIGINT NOT NULL,
    last_update_id BIGINT NOT NULL,
    bids JSONB NOT NULL, -- Array of [price, quantity]
    asks JSONB NOT NULL, -- Array of [price, quantity]
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()) * 1000,
    PRIMARY KEY (id, timestamp)
);

-- Convert to hypertable
SELECT create_hypertable('depth_snapshots', 'timestamp', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_depth_symbol ON depth_snapshots(symbol, timestamp DESC);
