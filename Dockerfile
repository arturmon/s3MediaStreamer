FROM golang:1.23-alpine

RUN adduser -D -g '' appuser
RUN mkdir -p /app
LABEL author="Artur Mudrykh"
WORKDIR /app
COPY conf/ conf/
COPY /migrations/psql/ ./migrations/psql/
COPY s3stream .
COPY acl/ acl/
RUN chown -R appuser:appuser /app
RUN chmod +x /app/s3stream
USER appuser

# Add Health Check
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget -qO- http://$(hostname -i):10000/health/liveness || exit 1

ENTRYPOINT [ "/app/s3stream" ]

EXPOSE 10000
