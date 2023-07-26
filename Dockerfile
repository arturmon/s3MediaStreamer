FROM golang:1.20-alpine

RUN adduser -D -g '' appuser
RUN mkdir -p /app
LABEL author="Artur Mudrykh"
RUN ls -la .
WORKDIR /app
RUN ls -la .
COPY . .

RUN ls -la .
USER appuser
CMD [ "/app/main" ]

EXPOSE 10000
