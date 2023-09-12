all: build

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

build: go.sum
#	@swag init
	@go mod tidy
	@go fmt ./...
	@go build  -o build/dataserver .

build-linux: go.sum
	GOOS=linux GOARCH=amd64 $(MAKE) build

build-window: go.sum
	GOOS=windows GOARCH=amd64 $(MAKE) build

docker:
	docker build -t redesblock/dataserver .

.PHONY: all build build-linux build-window