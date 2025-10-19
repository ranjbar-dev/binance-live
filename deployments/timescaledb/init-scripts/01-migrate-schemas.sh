#!/bin/bash
set -e

echo "========================================"
echo "Starting schema migrations..."
echo "========================================"

# Define the schema directory relative to where SQL files are mounted in the container
SCHEMA_DIR="/sql/schemas"

# Array of schema files in the order they should be executed
SCHEMA_FILES=(
    "symbols.sql"
    "sync_status.sql"
    "klines.sql"
    "tickers.sql"
)

# Execute each schema file
for file in "${SCHEMA_FILES[@]}"; do
    if [ -f "$SCHEMA_DIR/$file" ]; then
        echo "Executing schema: $file"
        psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$SCHEMA_DIR/$file"
        echo "✓ Successfully executed: $file"
    else
        echo "⚠ Warning: Schema file not found: $SCHEMA_DIR/$file"
    fi
done

echo "========================================"
echo "Schema migrations completed!"
echo "========================================"

