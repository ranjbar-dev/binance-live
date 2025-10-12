# Use the official Redis Commander image
FROM rediscommander/redis-commander:latest

# Set environment variables for Redis connection
ENV REDIS_HOSTS=dragonfly:6379
ENV REDIS_PASSWORD=password
ENV HTTP_USER=admin
ENV HTTP_PASSWORD=admin
ENV PORT=8081

# Create non-root user
RUN addgroup -g 1001 -S rediscommander && \
    adduser -u 1001 -S rediscommander -G rediscommander

# Switch to non-root user
USER rediscommander

# Expose Redis Commander port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Use the default command from the base image
