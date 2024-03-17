
CONFIG_PATH=./pkg/loadbalancer/bootstrap.yaml

CUR_DIR=$(shell pwd)
GO_BIN=$(CUR_DIR)/bin

# Go build command
GOBUILD=go build

# Makefile targets
.PHONY: all build run clean

all: build

clean_loadbalancer:
	@echo "Cleaning loadbalancer"
	rm -f $(GO_BIN)/loadbalancer

build_loadbalancer: clean_loadbalancer
	@echo "Building loadbalancer"
	mkdir -p $(GO_BIN)
	$(GOBUILD) -o $(GO_BIN)/loadbalancer ./cmd/lbctl

run_loadbalancer:
	@echo "Running loadbalancer"
	$(GO_BIN)/loadbalancer start --config $(CONFIG_PATH)

clean_client:
	@echo "Cleaning client"
	rm -rf tests/integration/client/client

build_client: clean_client
	@echo "Building client for testing"
	mkdir -p tests/integration/client
	go build -o tests/integration/client/client tests/integration/client/client.go

run_client:
	@echo "Running client"
	tests/integration/client/client

clean_server:
	@echo "Cleaning server"
	rm -f tests/integration/server/server

build_server: clean_server
	@echo "Building server for testing"
	mkdir -p tests/integration/server
	go build -o tests/integration/server/server tests/integration/server/server.go

run_server:
	@echo "Running server"
	tests/integration/server/server

build:
	$(MAKE) build_loadbalancer
	$(MAKE) build_client
	$(MAKE) build_server

run:
	$(MAKE) run_loadbalancer

clean:
	$(MAKE) clean_loadbalancer
	$(MAKE) clean_client
	$(MAKE) clean_server

lint:
	@echo "Running linters"
	golangci-lint run


test:
	@echo "Running tests"
	go test -race -timeout 10s ./...

gen_mocks:
	@echo "Generating mocks"
	mockgen -package mocks -destination=tests/mocks/mock_net_conn.go net Conn
	mockgen -package mocks -destination=tests/mocks/mock_loadbalance.go -source=lib/loadbalance/interface.go