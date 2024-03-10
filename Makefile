VERSION=$(shell cat VERSION)
PROJECT='spine'
build:
	@echo "Building version ${VERSION}"
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/${PROJECT}-${VERSION} .

run:
	./bin/${PROJECT}-${VERSION}