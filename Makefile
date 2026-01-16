.PHONY: build build-all clean install test run

BINARY_NAME=gogws
VERSION=1.0.0
BUILD_DIR=build

build:
	go build -o ${BUILD_DIR}/${BINARY_NAME} ./cmd/gogws

build-all:
	GOOS=linux GOARCH=amd64 go build -o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64 ./cmd/gogws
	GOOS=darwin GOARCH=amd64 go build -o ${BUILD_DIR}/${BINARY_NAME}-darwin-amd64 ./cmd/gogws
	GOOS=darwin GOARCH=arm64 go build -o ${BUILD_DIR}/${BINARY_NAME}-darwin-arm64 ./cmd/gogws
	GOOS=windows GOARCH=amd64 go build -o ${BUILD_DIR}/${BINARY_NAME}-windows-amd64.exe ./cmd/gogws

clean:
	rm -rf ${BUILD_DIR}
	go clean

install:
	go install ./cmd/gogws

test:
	go test -v ./...

run:
	go run ./cmd/gogws

fmt:
	go fmt ./...

lint:
	golangci-lint run

deps:
	go mod download
	go mod tidy
