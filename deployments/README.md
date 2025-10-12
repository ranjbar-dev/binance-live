# Deployments

This directory contains deployment configurations for each service in the Binance Live Data Collector.

## Directory Structure

```
deployments/
├── app/                    # Main application container
├── timescaledb/           # TimescaleDB (PostgreSQL) container
├── dragonfly/             # DragonflyDB (Redis-compatible) container
├── redis-commander/       # Redis Commander (monitoring tool) container
└── README.md              # This file
```

## Setup Instructions

### 1. Environment Variables

Each service directory contains a `.env.example` file. Copy these to `.env` files and modify the values as needed:

```bash
# Copy environment files for each service
cp deployments/app/.env.example deployments/app/.env
cp deployments/timescaledb/.env.example deployments/timescaledb/.env
cp deployments/dragonfly/.env.example deployments/dragonfly/.env
cp deployments/redis-commander/.env.example deployments/redis-commander/.env
```

### 2. Build and Deploy

You can build and deploy each service individually using their respective Dockerfiles:

```bash
# Build the main application
docker build -f deployments/app/Dockerfile -t binance-live:latest .

# Build TimescaleDB (if you need custom configuration)
docker build -f deployments/timescaledb/Dockerfile -t binance-timescaledb:latest ./deployments/timescaledb/

# Build DragonflyDB (if you need custom configuration)
docker build -f deployments/dragonfly/Dockerfile -t binance-dragonfly:latest ./deployments/dragonfly/

# Build Redis Commander (if you need custom configuration)
docker build -f deployments/redis-commander/Dockerfile -t binance-redis-commander:latest ./deployments/redis-commander/
```

### 3. Docker Compose Integration

Update your `docker-compose.yml` to use the custom Dockerfiles:

```yaml
services:
  app:
    build:
      context: .
      dockerfile: deployments/app/Dockerfile
    env_file:
      - deployments/app/.env
    # ... other configuration

  timescaledb:
    build:
      context: ./deployments/timescaledb
      dockerfile: Dockerfile
    env_file:
      - deployments/timescaledb/.env
    # ... other configuration

  dragonfly:
    build:
      context: ./deployments/dragonfly
      dockerfile: Dockerfile
    env_file:
      - deployments/dragonfly/.env
    # ... other configuration

  redis-commander:
    build:
      context: ./deployments/redis-commander
      dockerfile: Dockerfile
    env_file:
      - deployments/redis-commander/.env
    # ... other configuration
```

## Service Configuration

### Application (`app/`)
- **Purpose**: Main Binance Live Data Collector application
- **Key Features**: 
  - Multi-stage build for optimal image size
  - Non-root user for security
  - Includes both server and CLI binaries
  - Health check placeholder

### TimescaleDB (`timescaledb/`)
- **Purpose**: Time-series database for storing kline, ticker, and trade data
- **Key Features**:
  - Based on official TimescaleDB image
  - Optimized PostgreSQL configuration for time-series workloads
  - Custom initialization scripts support
  - Performance tuning for high-volume data

### DragonflyDB (`dragonfly/`)
- **Purpose**: Redis-compatible in-memory database for live data publishing
- **Key Features**:
  - High-performance Redis alternative
  - Automatic snapshots for persistence
  - Memory optimization settings
  - Authentication and security

### Redis Commander (`redis-commander/`)
- **Purpose**: Web-based Redis/DragonflyDB management interface
- **Key Features**:
  - Web UI for database monitoring
  - Authentication protection
  - Dark theme configuration
  - Read-only mode option

## Security Considerations

1. **Change default passwords** in all `.env` files
2. **Use strong passwords** for production deployments
3. **Enable TLS/SSL** for production environments
4. **Restrict network access** using Docker networks
5. **Regular security updates** for base images

## Production Deployment

For production deployments, consider:

1. **Resource Limits**: Set appropriate CPU and memory limits
2. **Persistent Storage**: Use named volumes or bind mounts for data persistence
3. **Monitoring**: Add monitoring and alerting solutions
4. **Backup Strategy**: Implement automated backups for TimescaleDB
5. **Load Balancing**: Use reverse proxy for web interfaces
6. **Network Security**: Use custom Docker networks and firewall rules

## Troubleshooting

### Common Issues

1. **Permission Errors**: Ensure proper file permissions for data directories
2. **Connection Issues**: Check network connectivity between containers
3. **Memory Issues**: Adjust memory limits based on your system resources
4. **Port Conflicts**: Ensure ports are not already in use on your host system

### Useful Commands

```bash
# View logs for specific service
docker-compose logs -f <service_name>

# Connect to TimescaleDB
docker-compose exec timescaledb psql -U postgres -d postgres

# Connect to DragonflyDB
docker-compose exec dragonfly redis-cli -a password

# Check service health
docker-compose ps
```

## CLI Usage

The application container includes a CLI tool for utility operations:

```bash
# Run CLI commands
docker-compose exec app ./binance-cli --help

# Sync all klines
docker-compose exec app ./binance-cli sync all-klines --intervals 1m,15m,1h,4h,1d

# Sync specific symbol
docker-compose exec app ./binance-cli sync symbol-kline --symbol BTCUSDT --intervals 1m,15m

# Check sync status
docker-compose exec app ./binance-cli status sync

# Manage symbols
docker-compose exec app ./binance-cli symbols list --active-only
```
