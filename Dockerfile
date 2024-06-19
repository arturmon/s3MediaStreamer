FROM golang:1.20-alpine

RUN adduser -D -g '' appuser
RUN mkdir -p /app
LABEL author="Artur Mudrykh"
WORKDIR /app
RUN ls -la .
RUN ls -la /
RUN ls -la /workspace
COPY conf/ /app/conf/
COPY migrations/psql/ /app/migrations/psql/
COPY s3stream /app/
COPY acl/ /app/acl/
RUN chown -R appuser:appuser /app
RUN chmod +x /app/s3stream
USER appuser

# Add Health Check
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget -qO- http://$(hostname -i):10000/health/liveness || exit 1

ENTRYPOINT [ "/app/s3stream" ]
EXPOSE 10000
