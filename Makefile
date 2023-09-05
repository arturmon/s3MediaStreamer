.PHONY: build
build:
	@$(MAKE) swagger
	go build app

.PHONY: run
run:
	@$(MAKE) swagger
	go run app/main.go

swagger:
	cd app && swag init --parseDependency --parseDepth=1

lint:
	golangci-lint run --timeout 5m

test:
	go test -v ./...

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