export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
GOOS=$(shell go env GOOS)

dep:
	go mod tidy

generate:
	rm -rf ./internal/mocks/*
	go install go.uber.org/mock/mockgen/...@v0.1.0
	go install golang.org/x/tools/cmd/stringer@latest
	go generate -x ./internal/...

build:
	CGO_ENABLED=0 GOOS=${GOOS} go build -ldflags "-X=main.Revision=${REVISION} -X=main.Version=${VERSION}" -o bin/main ./cmd/main.go

test:
	go test -race -coverprofile=coverage.out $$(go list ./... | grep -v "/cmd\|/mocks\|generated")
	go tool cover -func=coverage.out

lint:
	@#curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.8.0
	golangci-lint run ./...

fmt:
	go run mvdan.cc/gofumpt@latest -l -w .
	go run golang.org/x/tools/cmd/goimports@latest -w -local github.com,golang.org,go.uber.org,gopkg.in .

audit:
	go mod verify
	go vet ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

updateDeps:
	go get -u ./...
	go mod tidy

all: dep generate fmt lint test build

clean: ## Remove build related file
	rm -fr ./bin
	rm -f ./coverage.out
	rm -f $(CUSTOM_GCL)
	@go clean -i -r -cache -testcache
