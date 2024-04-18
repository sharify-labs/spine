.PHONY: all build clean keys lint run tidy
PROJECT='spine'

.PHONY: all build clean lint run tidy

all: tidy lint build

build: clean
	@echo "Building ${PROJECT}"
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=dev" -o bin/${PROJECT}-dev.bin .

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
	@# Spine -> Zephyr Admin Key
	@openssl rand -out admin-key.bin 64
	@echo "ZEPHYR_ADMIN_KEY='$$(openssl base64 -in admin-key.bin | tr -d '\n')'" >> .env
	@echo "ADMIN_KEY_HASH='$$(openssl dgst -sha512 admin-key.bin | cut -d ' ' -f 2)'" >> .env
	@rm -f admin-key.bin
	@echo "Keys generated and saved to .env"

lint: tidy
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.2 run ./...

run: build
	./bin/${PROJECT}-dev.bin

tidy:
	@go mod tidy -v
	@go run mvdan.cc/gofumpt@latest -w -l .
