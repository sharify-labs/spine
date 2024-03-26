.PHONY: clean build run
PROJECT='spine'

all: build

build: clean
	@echo "Building ${PROJECT}"
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=dev" -o bin/${PROJECT}-dev.bin .

clean:
	@rm -rf ./bin

keys:
	@echo "Generating keys..."
	@echo "SESSION_AUTH_KEY_64='$(shell openssl rand -base64 64 | tr -d '\n')'" >> .env
	@echo "SESSION_ENC_KEY_32='$(shell openssl rand -base64 32 | tr -d '\n')'" >> .env
	@echo "Keys generated and saved to .env"

run: build
	./bin/${PROJECT}-dev.bin
