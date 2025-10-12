# Use the official TimescaleDB image
FROM timescale/timescaledb:latest-pg17

# Set environment variables
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres
ENV POSTGRES_DB=postgres

# Copy custom configuration files if needed
# COPY postgresql.conf /etc/postgresql/postgresql.conf
# COPY pg_hba.conf /etc/postgresql/pg_hba.conf

# Copy initialization scripts
COPY init-scripts/ /docker-entrypoint-initdb.d/

# Set custom PostgreSQL configuration
RUN echo "max_connections = 300" >> /usr/share/postgresql/postgresql.conf.sample && \
    echo "shared_preload_libraries = 'timescaledb'" >> /usr/share/postgresql/postgresql.conf.sample && \
    echo "max_worker_processes = 16" >> /usr/share/postgresql/postgresql.conf.sample && \
    echo "max_parallel_workers_per_gather = 4" >> /usr/share/postgresql/postgresql.conf.sample && \
    echo "max_parallel_workers = 16" >> /usr/share/postgresql/postgresql.conf.sample

# Create data directory
RUN mkdir -p /var/lib/postgresql/data

# Set permissions
RUN chown -R postgres:postgres /var/lib/postgresql/data

# Expose PostgreSQL port
EXPOSE 5432

# Health check
HEALTHCHECK --interval=10s --timeout=5s --retries=5 \
    CMD pg_isready -U postgres || exit 1

# Use the default entrypoint from the base image
# CMD will be set via docker-compose or command line
