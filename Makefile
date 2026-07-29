VERSION     ?= 0.1.0
UPDATED     ?= $(shell date +%Y-%m-%d)
TAG         ?= v$(VERSION)
LDFLAGS     := -X bingo/internal/version.Version=$(VERSION) -X bingo/internal/version.Updated=$(UPDATED)

BINARY      := bingo
RELEASE_DARWIN := bingo_darwin_arm64
RELEASE_LINUX  := bingo_linux_amd64
RELEASE_WIN    := bingo_windows_amd64.exe
RELEASES       := $(RELEASE_DARWIN) $(RELEASE_LINUX) $(RELEASE_WIN)

.PHONY: build release tag clean version ui

## Build the React UI into internal/server/static (commit the output)
ui:
	npm ci --prefix web
	npm run build --prefix web

## Build a local binary (bingo)
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

## Build release assets for darwin/arm64, linux/amd64, and windows/amd64
release:
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(RELEASE_DARWIN) .
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(RELEASE_LINUX) .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(RELEASE_WIN) .

## Create and push a version tag (GitHub Actions publishes the release)
tag:
	@test -z "$$(git status --porcelain)" || (echo "error: working tree not clean" >&2; exit 1)
	git tag -a $(TAG) -m "bingo $(VERSION)"
	git push origin $(TAG)
	@echo "Pushed $(TAG). GitHub Actions will attach darwin/linux/windows release assets."

## Show version that would be embedded
version:
	@echo "$(VERSION) (updated $(UPDATED))"

clean:
	rm -f $(BINARY) $(RELEASES)
