# Create base for installin binaries
BIN = $(CURDIR)/bin
$(BIN):
	@mkdir -p $@
$(BIN)/golangci-lint: | $(BIN)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.24.0
$(BIN)/gopherbadger: | $(BIN)
	GOBIN=$(BIN) go get github.com/jpoles1/gopherbadger

# Binaries to install
GOLANGCI-LINT = $(BIN)/golangci-lint
GOPHERBADGER = $(BIN)/gopherbadger

###############################################################################
###############################################################################
###############################################################################

all: clean static test build

# Build binaries
build:
	go build -a -o bin/helmenv helmenv/main.go
	go build -a -o bin/kbenv kbenv/main.go

clean:
	-rm bin/*
	-rm releases/*

static: | $(GOLANGCI-LINT) $(GOPHERBADGER)
	$(GOLANGCI-LINT) run ./...
	$(GOPHERBADGER) -md="README.md"

test:
	go test -coverprofile cover.out ./...

PLATFORMS := linux-amd64 linux-386 darwin-amd64 darwin-386 windows-amd64 windows-386
temp = $(subst -, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))
releases: $(PLATFORMS)
$(PLATFORMS):
	@mkdir -p releases; \
	CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) go build -a -o bin/helmenv-$(os)-$(arch) helmenv/main.go; \
	CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) go build -a -o bin/kbenv-$(os)-$(arch) kbenv/main.go; \
	tar -C bin -cvzf releases/helmenv-$(os)-$(arch).tar.gz helmenv-$(os)-$(arch); \
	tar -C bin -cvzf releases/kbenv-$(os)-$(arch).tar.gz kbenv-$(os)-$(arch)
