# environment variables
BINARY				:= timelock-worker
VERSION				:= $(shell git describe --tags --always --abbrev=0 --match='v[0-9]*.[0-9]*.[0-9]*' 2> /dev/null | sed 's/^.//')
COMMIT_HASH			:= $(shell git rev-parse HEAD)
LDFLAGS				:= "-X $(PACKAGE)/cmd.Version=$(VERSION) -X $(PACKAGE)/cmd.Commit=$(COMMIT_HASH) -w -s"

# Folders
BIN_FOLDER			:= bin
SCRIPTS_F			:= scripts

# assets
C_GREEN=\033[0;32m
C_RED=\033[0;31m
C_BLUE=\033[0;34m
C_END=\033[0m

##########
# builds #
##########

.PHONY: all
all: clean test build

.PHONY: clean
clean:
	@echo "\n\t$(C_GREEN)# Cleaning environment$(C_END)"
	go clean -x
	rm -rf $(BIN_FOLDER)
	rm -f $(SCRIPTS_F)/report.html

.PHONY: test
test:
	@echo "\n\t$(C_GREEN)# Run test and generate new coverage.out$(C_END)"
	go test -short -coverprofile=coverage.out -covermode=atomic -race ./...

.PHONY: build
build: clean
	@echo "\n\t$(C_GREEN)# Build binary $(BINARY)$(C_END)"
	go build -trimpath -ldflags $(LDFLAGS) -o $(BIN_FOLDER)/$(BINARY) main.go

.PHONY: lint
lint:
	$(SCRIPTS_F)/golangci_html_report.sh