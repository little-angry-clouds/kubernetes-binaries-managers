# Create base for installing binaries
BIN = $(CURDIR)/bin
$(BIN):
	@mkdir -p $@
$(BIN)/golangci-lint: | $(BIN)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.24.0
$(BIN)/gopherbadger: | $(BIN)
	GOBIN=$(BIN) go get github.com/jpoles1/gopherbadger
$(BIN)/kind: | $(BIN)
	curl -Lo $(BIN)/kind https://kind.sigs.k8s.io/dl/v0.8.1/kind-$$(uname)-amd64; \
	chmod +x $(BIN)/kind

# Binaries to install
GOLANGCI-LINT = $(BIN)/golangci-lint
GOPHERBADGER = $(BIN)/gopherbadger
KIND = $(BIN)/kind

###############################################################################
###############################################################################
###############################################################################

all: clean static test build

# Build binaries
build:
	go build -a -o bin/helmenv cmd/helmenv/main.go; \
	go build -a -o bin/helm-wrapper cmd/helm-wrapper/main.go; \
	go build -a -o bin/kbenv cmd/kbenv/main.go; \
	go build -a -o bin/kubectl-wrapper cmd/kubectl-wrapper/main.go;

clean:
	-rm -r bin/
	-rm -r releases/

static: | $(GOLANGCI-LINT) $(GOPHERBADGER)
	$(GOLANGCI-LINT) run ./... --timeout 2m0s
	$(GOPHERBADGER) -md="README.md"

unit-test:
	go test ./...

int-test: | $(KIND)
	bats tests/wrapers.test
	bats tests/managers.test

PLATFORMS := linux-amd64 linux-386 darwin-amd64 darwin-386 windows-amd64 windows-386
temp = $(subst -, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))
releases: $(PLATFORMS)
$(PLATFORMS):
	@mkdir -p releases; \
	CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) go build -a -o bin/helmenv-$(os)-$(arch) cmd/helmenv/main.go; \
	CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) go build -a -o bin/helm-wrapper-$(os)-$(arch) cmd/helm-wrapper/main.go; \
	CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) go build -a -o bin/kbenv-$(os)-$(arch) cmd/kbenv/main.go; \
	CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) go build -a -o bin/kubectl-wrapper-$(os)-$(arch) cmd/kubectl-wrapper/main.go; \
	tar -C bin -cvzf releases/helmenv-$(os)-$(arch).tar.gz helmenv-$(os)-$(arch) helm-wrapper-$(os)-$(arch); \
	tar -C bin -cvzf releases/kbenv-$(os)-$(arch).tar.gz kbenv-$(os)-$(arch) kubectl-wrapper-$(os)-$(arch);
