FROM redis:8.4.0-alpine

# Copy the Sentinel configuration
COPY services/redis-sentinel/sentinel.conf /usr/local/etc/redis/sentinel.conf

# Run Redis Sentinel
CMD ["redis-sentinel", "/usr/local/etc/redis/sentinel.conf"]