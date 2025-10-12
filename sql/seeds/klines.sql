-- Sample kline data (this would typically be populated by the application)
-- This file is intentionally left mostly empty as kline data is time-series data
-- that should be populated by the live data collection system

-- You can add sample historical data here if needed for testing:
-- INSERT INTO klines (
--     symbol, interval, open_time, close_time, open_price, high_price,
--     low_price, close_price, volume, quote_volume, trades_count,
--     taker_buy_volume, taker_buy_quote_volume
-- ) VALUES (
--     'BTCUSDT', '1d', 1697097600000, 1697183999999, 35000.00, 35500.00,
--     34500.00, 35200.00, 1000.50, 35100000.25, 2500,
--     600.25, 21060012.75
-- );

-- Note: Use EXTRACT(EPOCH FROM NOW()) * 1000 for current timestamp in milliseconds
