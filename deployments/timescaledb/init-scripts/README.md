# TimescaleDB Initialization Scripts

This directory contains scripts that automatically execute when the TimescaleDB container starts for the first time.

## How It Works

PostgreSQL (and TimescaleDB) automatically runs scripts found in `/docker-entrypoint-initdb.d/` during initial database creation. These scripts are executed in **alphabetical order**.

## Scripts

### 01-migrate-schemas.sh
- **Purpose**: Creates all database tables and schemas
- **Source**: Executes SQL files from `sql/schemas/` directory
- **Execution Order**:
  1. `symbols.sql` - Trading pair symbols table
  2. `sync_status.sql` - Synchronization tracking table
  3. `klines.sql` - Candlestick/kline data table
  4. `tickers.sql` - 24hr ticker statistics table

### 02-seed-data.sh
- **Purpose**: Populates initial data into the database
- **Source**: Executes SQL files from `sql/seeds/` directory
- **Execution Order**: Same as schema migration (follows dependencies)

## Configuration

The scripts use the following environment variables from docker-compose:
- `POSTGRES_USER` - Database user (default: `postgres`)
- `POSTGRES_DB` - Database name (default: `postgres`)
- `POSTGRES_PASSWORD` - Database password (set in docker-compose)

## Volume Mounts

The following volumes must be mounted in docker-compose.yml:
```yaml
volumes:
  - ./deployments/timescaledb/init-scripts:/docker-entrypoint-initdb.d:ro
  - ./sql/schemas:/sql/schemas:ro
  - ./sql/seeds:/sql/seeds:ro
```

## Error Handling

- Scripts use `set -e` to stop execution on any error
- `psql` is called with `-v ON_ERROR_STOP=1` to halt on SQL errors
- Missing files trigger warnings but don't stop execution

## Adding New Scripts

To add new initialization scripts:

1. Create a new `.sh` file with a numeric prefix (e.g., `03-custom-setup.sh`)
2. Make sure it's executable: `chmod +x 03-custom-setup.sh`
3. Add the shebang line: `#!/bin/bash`
4. Use `set -e` for error handling
5. Access environment variables: `$POSTGRES_USER`, `$POSTGRES_DB`

Example:
```bash
#!/bin/bash
set -e

echo "Running custom setup..."
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE EXTENSION IF NOT EXISTS timescaledb;
    -- Your custom SQL here
EOSQL
```

## Adding New Schema/Seed Files

1. Add SQL file to `sql/schemas/` or `sql/seeds/`
2. Update the respective script (`01-migrate-schemas.sh` or `02-seed-data.sh`)
3. Add the filename to the `SCHEMA_FILES` or `SEED_FILES` array
4. Place it in the correct execution order (consider dependencies)

Example:
```bash
SCHEMA_FILES=(
    "symbols.sql"
    "sync_status.sql"
    "your_new_table.sql"  # Add here
    "klines.sql"
    ...
)
```

## Testing

To test the initialization scripts:

1. **Stop and remove the TimescaleDB container**:
   ```bash
   docker-compose down
   docker volume rm binance-live_timescaledb_data
   ```

2. **Start fresh**:
   ```bash
   docker-compose up timescaledb
   ```

3. **Check logs**:
   ```bash
   docker-compose logs timescaledb
   ```

4. **Verify tables**:
   ```bash
   docker-compose exec timescaledb psql -U postgres -d postgres -c "\dt"
   ```

## Important Notes

⚠️ **These scripts only run once** during initial database creation. They will NOT run if the database already exists.

To re-run initialization scripts:
- Delete the database volume: `docker volume rm binance-live_timescaledb_data`
- Or use a fresh database directory

🔒 **Scripts are mounted as read-only** (`:ro`) for security.

📝 **Script execution is logged** to stdout and can be viewed with `docker-compose logs timescaledb`.

