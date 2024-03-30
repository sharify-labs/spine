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
	@openssl ecparam -genkey -name prime256v1 -noout -out ec-private.pem
	@openssl ec -in ec-private.pem -pubout -out ec-public.pem
	@echo "JWT_PRIVATE_KEY='$$(cat ec-private.pem | base64 | tr -d '\n')'" >> .env
	@echo "JWT_PUBLIC_KEY='$$(cat ec-public.pem | base64 | tr -d '\n')'" >> .env
	@rm -f ec-private.pem ec-public.pem
	@echo "Keys generated and saved to .env"

run: build
	./bin/${PROJECT}-dev.bin
