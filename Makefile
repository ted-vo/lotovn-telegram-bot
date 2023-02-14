SHELL=/bin/bash
BINARY_NAME=main.bin

prerequisites:
	go mod tidy

clean:
	go clean
	rm -f "$(BINARY_NAME)" __debug_bin

build: prerequisites
	go build -o ./build/$(BINARY_NAME) cmd/main.go

build-linux: prerequisites
	GOOS=linux GARCH=amd64 go build -o ./build/$(BINARY_NAME) cmd/main.go

run: build
	TOKEN=6002114207:AAFmDlfRPRnu-z5FuSaib_9YHGrObkrwsu8 bash -c './build/$(BINARY_NAME)'

