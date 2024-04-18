BINARY_NAME=media-streamer
.DEFAULT_GOAL := run
build:
	@$(MAKE) swagger
	go build -o ./${BINARY_NAME} app/main.go


run: build
	./${BINARY_NAME}

swagger:
	cd app && swag init --parseDependency --parseDepth=1

clean:
	go clean
	rm ./${BINARY_NAME}

lint:
	golangci-lint run --sort-results

test:
	go test ./...

test_coverage:
	go test ./... -coverprofile=coverage.out

init: ##- Runs make modules, tools and tidy.
	@$(MAKE) modules
	@$(MAKE) install-tools
	@$(MAKE) tidy

modules:
	go mod download

install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest

tidy:
	go mod tidy

goimports:
	goimports -w file_name.go

compose-down: ### Down docker-compose
	docker-compose down --remove-orphans
.PHONY: compose-down