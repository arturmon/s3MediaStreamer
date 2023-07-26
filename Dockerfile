FROM golang:1.20-alpine

RUN mkdir -p /app
COPY app/cmd/main/main /app

LABEL author="Artur Mudrykh"

WORKDIR /app
CMD [ "main" ]

EXPOSE 10000