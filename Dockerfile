FROM golang:1.20-alpine

RUN adduser -D -g '' appuser
RUN mkdir -p /app
LABEL author="Artur Mudrykh"
WORKDIR /app
COPY /migrations/psql/ ./migrations/psql/
COPY s3stream .
RUN chown -R appuser:appuser /app
RUN chmod +x /app/s3stream
USER appuser

# Add Health Check
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget -qO- http://localhost:10000/ || exit 1

ENTRYPOINT [ "/app/s3stream" ]

EXPOSE 10000