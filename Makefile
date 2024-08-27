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

compose-down: ## Down docker-compose
	docker-compose down --remove-orphans
.PHONY: compose-down


debug-template:
	helm template s3media ./chart
.PHONY: debug-template

deploy-app:
	helm version
	helm list --namespace media
	helm install s3stream ./chart --namespace media --create-namespace
.PHONY: deploy

deploy-update-app:
	helm version
	helm list --namespace media
	helm upgrade s3stream ./chart --namespace media
.PHONY: deploy



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

deploy-delete-app:
	helm list --namespace media
	helm uninstall s3stream --namespace media
	kubectl delete namespace media
.PHONY: deploy-delete


infra-diff-deploy:
	helmfile diff -f ./infra-kube/helmfile.yaml
.PHONY: infra-diff-deploy

infra-deploy:
	helm upgrade --install ingress-nginx ingress-nginx \
  	--repo https://kubernetes.github.io/ingress-nginx \
  	--set controller.opentelemetry.enabled=true \
  	--set controller.metrics.enabled=true \
  	--namespace ingress-nginx --create-namespace
	kubectl apply -f infra-kube/weave-scope/k8s-scope.yaml
	kubectl apply -f infra-kube/weave-scope/weave-scope-ingress.yaml
	helm upgrade --install greylog ./infra-kube/graylog  --namespace logs --create-namespace
	helmfile apply -f ./infra-kube/helmfile.yaml
.PHONY: infra-deploy

infra-clean:
	helm uninstall greylog --namespace logs
	kubectl delete namespace logs
.PHONY: infra-clean
