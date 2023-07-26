FROM golang:1.20-alpine

RUN mkdir -p /app
COPY main /app/main

LABEL author="Artur Mudrykh"

WORKDIR /app
CMD [ "main" ]

EXPOSE 10000