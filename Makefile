COVERPROFILE=coverage.out
TAG_NAME=$(shell git describe --tags --long --dirty 2>/dev/null || echo "unknown")
GOX_LDFLAGS="-X main.version=$(TAG_NAME) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%S%z)"

.PHONY: clean install test

default: build

build: test
	gox -output "dist/{{.OS}}_{{.Arch}}_{{.Dir}}" -ldflags $(GOX_LDFLAGS)

install:
	go install -ldflags $(GOX_LDFLAGS)

clean:
	go clean

test:
	go test -v -coverprofile=$(COVERPROFILE)
	go tool cover -html=$(COVERPROFILE)
	rm $(COVERPROFILE)
