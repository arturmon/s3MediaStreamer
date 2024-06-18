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


deploy:
	helm version
	helm list --namespace media
	helm install s3stream ./chart --namespace media --create-namespace
.PHONY: deploy

deploy-update:
	helm version
	helm list --namespace media
	helm upgrade s3stream ./chart --namespace media
.PHONY: deploy

deploy-delete:
	helm list --namespace media
	helm uninstall s3stream --namespace media
	kubectl delete namespace media
.PHONY: deploy-delete

helm-dependency:
	helm dependency build ./chart
.PHONY: helm-dependency

infra-deploy:
	helm upgrade --install ingress-nginx ingress-nginx \
  	--repo https://kubernetes.github.io/ingress-nginx \
  	--set controller.opentelemetry.enabled=true \
  	--set controller.metrics.enabled=true \
  	--namespace ingress-nginx --create-namespace
	kubectl apply -f https://github.com/weaveworks/scope/releases/download/v1.13.2/k8s-scope.yaml
	kubectl apply -f infra-kube/weave-scope-ingress.yaml
	helm install greylog ./infra-kube/graylog  --namespace graylog --create-namespace
.PHONY: infra-deploy

infra-clean:
	helm uninstall greylog --namespace graylog
	kubectl delete namespace graylog
.PHONY: infra-clean
