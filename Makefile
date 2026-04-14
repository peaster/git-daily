BINARY  := git-daily
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

PLATFORMS := \
	darwin/arm64 \
	darwin/amd64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64

.PHONY: build install clean dist

build:
	go build -ldflags '$(LDFLAGS)' -o $(BINARY) .

install:
	go install -ldflags '$(LDFLAGS)' .

clean:
	rm -rf dist/ $(BINARY)

dist: clean
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		out=dist/$(BINARY)-$${os}-$${arch}; \
		[ "$$os" = "windows" ] && out="$${out}.exe"; \
		echo "Building $$out"; \
		GOOS=$$os GOARCH=$$arch go build -ldflags '$(LDFLAGS)' -o $$out .; \
	done
