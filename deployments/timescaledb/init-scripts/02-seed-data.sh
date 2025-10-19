#!/bin/bash
set -e

echo "========================================"
echo "Starting data seeding..."
echo "========================================"

# Define the seed directory relative to where SQL files are mounted in the container
SEED_DIR="/sql/seeds"

# Array of seed files in the order they should be executed
SEED_FILES=(
    "symbols.sql"
    "sync_status.sql"
    "klines.sql"
    "tickers.sql"
)

# Execute each seed file
for file in "${SEED_FILES[@]}"; do
    if [ -f "$SEED_DIR/$file" ]; then
        echo "Seeding data from: $file"
        psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$SEED_DIR/$file"
        echo "✓ Successfully seeded: $file"
    else
        echo "⚠ Warning: Seed file not found: $SEED_DIR/$file"
    fi
done

echo "========================================"
echo "Data seeding completed!"
echo "========================================"

