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

deploy-devspase-app:
	devspace use namespace media
	devspace deploy
	devspace dev
	devspace ui
	devspace purge
	devspace cleanup images
	devspace cleanup local-registry
.PHONY: deploy-devspase-app

deploy-devspace-app: deploy-devspase-app-select-namespace deploy-devspase-app-deploy


deploy-devspace-app-clean: deploy-devspase-app-purge deploy-devspase-app-clean-img

deploy-devspace-app-purge: deploy-devspase-app-delete-local-repo

deploy-devspase-app-select-namespace:
	devspace use namespace media
.PHONY: deploy-devspase-app-select-namespace

deploy-devspase-app-deploy:
	devspace deploy
.PHONY: deploy-devspase-app-deploy

deploy-devspase-app-dev:
	devspace dev
.PHONY: deploy-devspase-app-dev

deploy-devspase-app-ui:
	devspace ui
.PHONY: deploy-devspase-app-ui

deploy-devspase-app-purge:
	devspace purge
.PHONY: deploy-devspase-app-purge

deploy-devspase-app-clean-img:
	devspace cleanup images
.PHONY: deploy-devspase-app-clean-img

deploy-devspase-app-delete-local-repo:
	devspace cleanup local-registry
.PHONY: deploy-devspase-app-delete-local-repo

port-forwarding:
	kubectl port-forward service/postgresql 5432:5432 -n database
.PHONY: port-forwarding