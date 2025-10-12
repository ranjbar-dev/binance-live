# Use the official DragonflyDB image
FROM docker.dragonflydb.io/dragonflydb/dragonfly:latest

# Set environment variables
ENV DFLY_REQUIREPASS=password
ENV DFLY_SNAPSHOT_CRON="*/5 * * * *"

# Create data directory
RUN mkdir -p /data

# Set permissions for data directory
RUN chown -R dragonfly:dragonfly /data

# Expose DragonflyDB port
EXPOSE 6379

# Health check
HEALTHCHECK --interval=10s --timeout=5s --retries=5 \
    CMD dragonfly --logtostderr=true --alsologtostderr=false --check_config_only=true || exit 1

# Default command with configuration
CMD ["dragonfly", \
     "--requirepass=password", \
     "--snapshot_cron=*/5 * * * *", \
     "--dir=/data", \
     "--port=6379", \
     "--bind=0.0.0.0"]
