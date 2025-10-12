-- Insert default trading symbols (major pairs)
INSERT INTO symbols (symbol, base_asset, quote_asset, status, is_active) VALUES
    ('BTCUSDT', 'BTC', 'USDT', 'TRADING', true),
    ('ETHUSDT', 'ETH', 'USDT', 'TRADING', true),
    ('BNBUSDT', 'BNB', 'USDT', 'TRADING', true),
    ('ADAUSDT', 'ADA', 'USDT', 'TRADING', true),
    ('DOGEUSDT', 'DOGE', 'USDT', 'TRADING', true),
    ('XRPUSDT', 'XRP', 'USDT', 'TRADING', true),
    ('SOLUSDT', 'SOL', 'USDT', 'TRADING', true),
    ('DOTUSDT', 'DOT', 'USDT', 'TRADING', true),
    ('MATICUSDT', 'MATIC', 'USDT', 'TRADING', true),
    ('LINKUSDT', 'LINK', 'USDT', 'TRADING', true),
    ('AVAXUSDT', 'AVAX', 'USDT', 'TRADING', true),
    ('LTCUSDT', 'LTC', 'USDT', 'TRADING', true),
    ('ATOMUSDT', 'ATOM', 'USDT', 'TRADING', true),
    ('UNIUSDT', 'UNI', 'USDT', 'TRADING', true),
    ('FILUSDT', 'FIL', 'USDT', 'TRADING', true)
ON CONFLICT (symbol) DO NOTHING;
