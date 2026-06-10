BINARY=bin/pay-log
CONFIG_FILE=config.yaml

.PHONY: help run build build-web

help:
	@echo "  make run        Run server"
	@echo "  make build      Build to $(BINARY)"
	@echo "  make build-web  Build frontend"

run:
	go run main.go

build:
	mkdir -p bin && go build -o $(BINARY) .

build-web:
	cd web && npm run build

add-user:
	go run cmd/adduser/main.go -u $(USER) -p $(PASSWORD) -f $(CONFIG_FILE)
