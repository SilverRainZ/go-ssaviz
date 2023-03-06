GO		= go
TARGET	= ./bin/go-ssaviz 

.PHONY: build run test

build:
	mkdir bin || true
	go build -o $(TARGET) .

run: build
	$(TARGET) ./...

test:
	$(GO) test -v ./...
