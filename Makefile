PROJECT='spine'
build:
	@echo "Building ${PROJECT}"
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/${PROJECT}.bin .

run:
	./bin/${PROJECT}.bin