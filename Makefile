.PHONY: help clean fmt vet test test-all check build build-RPi mod install

help:    ## Show this help.
	@sed -ne '/@sed/!s/## //p' $(MAKEFILE_LIST)

clean:   ## Cleanup build artifacts including UI assets
	go clean; \
	rm -rf .build

install: install-go-tools

install-go-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.0;
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0;
	go install golang.org/x/tools/cmd/goimports@v0.20.0;
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.1.0;
	go install gotest.tools/gotestsum@v1.11.0;

build-go:   ## go binary
	go build -o .build/shpankids -ldflags="-X 'main.Version=$(APP_VERSION)' -s -w";

build-ui:   ## Build UI app
	cd ui; \
	npm run build; \
	cd ..;\
	mkdir -p .build/ui; \
	cp -rf ./ui/dist .build/ui

build-ui:   ## Build UI app
	cd ui; \
	npm run build; \
	cd ..;\
	mkdir -p .build/ui; \
	cp -rf ./ui/dist .build/ui

build: build-go build-ui

gen-shpan: ## Shpan codegen
	cd tools/codegen/; \
 	python render.py; \
 	goimports -w ../../ermodel

deploy-gcp:
	 gcloud run deploy shpankids --source . --region=europe-west1 --update-secrets=SHPAN_SECRETS=SHPANSECRET:latest

gen-openapi:
	cd tools; \
	./generate-openapi-server.sh


gen-all: install-go-tools gen-shpan gen-openapi gen-go

fmt:     ## Run "go fmt" on the entire gortia
	go fmt $(shell go list ./...)

vet:     ## Run "go vet" on the entire gortia
	go vet $(shell go list ./...)

mod:
	go mod tidy

gen-go:
	go generate ./...

test:    ## Run all the tests in the project except for integration-tests
	gotestsum  --format-hide-empty-pkg -- -v -race -short ./...
test-apple:    ## apple silicon warning temporary fix:
	gotestsum  --format-hide-empty-pkg -- -v -race -short -ldflags=-extldflags=-Wl,-ld_classic ./...

test-all:    ## Run all the tests in the project including integration-tests
	gotestsum --format-hide-empty-pkg -- -tags=integration -v -race ./...

check: fmt vet test  ## Run fmt, vet and test
