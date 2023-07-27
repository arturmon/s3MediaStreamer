FROM golang:1.20-alpine

RUN adduser -D -g '' appuser
RUN mkdir -p /app
LABEL author="Artur Mudrykh"

WORKDIR /app
COPY albums .

USER appuser

# Add Health Check
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget -qO- http://localhost:10000/ || exit 1

CMD [ "/app/albums" ]

EXPOSE 10000
