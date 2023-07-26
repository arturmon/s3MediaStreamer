FROM golang:1.20-alpine

RUN adduser -D -g '' appuser
RUN mkdir -p /app
LABEL author="Artur Mudrykh"

WORKDIR /app
COPY albums .

RUN ls -la .
USER appuser
CMD [ "/app/albums" ]

EXPOSE 10000
