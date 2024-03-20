PROJECT='spine'
build:
	@echo "Building ${PROJECT}"
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=dev" -o bin/${PROJECT}-dev.bin .

keys:
	@echo "Generating keys..."
	@echo "SESSION_AUTH_KEY_64='$(shell openssl rand -base64 64)'" >> .env
	@echo "SESSION_ENC_KEY_32='$(shell openssl rand -base64 32)'" >> .env
	@echo "Keys generated and saved to .env"
