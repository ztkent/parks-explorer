.PHONY: build test clean app-up app-down app-fresh

# The name of the output binary
BINARY_NAME=parks

# The Go path
GOPATH=$(shell go env GOPATH)

# The build commands
GOBUILD=go build
GOTEST=go test
GOCLEAN=go clean
GOGET=go get
GOMODTIDY=go mod tidy
GOMODVENDOR=go mod vendor
GOINSTALL=go install

build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/parks

install:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/parks
	mv parks $(GOPATH)/bin

all: clean deps test build

.PHONY: app-up
app-up:
	docker compose -p parks --profile parks up

.PHONY: app-down
app-down:
	docker compose -p parks --profile parks down

.PHONY: app-fresh
app-fresh:
	docker compose -p parks --profile parks down
	docker compose -p parks --profile parks build --no-cache
	docker compose -p parks --profile parks up