PROJECT='spine'
build:
	@echo "Building ${PROJECT}"
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=dev" -o bin/${PROJECT}-dev.bin .
