ME = $(lastword $(MAKEFILE_LIST))
.DEFAULT_GOAL := help
.PHONY: help
help:  # prints this help
	@bash -c "$$AUTOGEN_HELP_BASH" < $(ME)

BINARY_NAME=tinytune
VERSION=0.0.1
COMMIT_HASH=$(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP=$(shell date '+%Y-%m-%dT%H:%M:%S')
LDFLAGS=-ldflags "-X 'main.Version=${VERSION}' -X 'main.CommitHash=${COMMIT_HASH}' -X 'main.BuildTimestamp=${BUILD_TIMESTAMP}' -X 'main.Mode=Production'"

.PHONY: build
build: ## build executables
	echo "Building frontend assets"
	make -C ./web build
	mkdir -p out/
	echo "Building executable"
	GOARCH=amd64 GOOS=linux go build ${LDFLAGS} -o out/${BINARY_NAME}_linux_amd64 cmd/tinytune/tinytune.go
	chmod +x out/${BINARY_NAME}_linux_amd64
	echo "Done"

.PHONY: run
run: ## run tinytune server
	go run cmd/tinytune/tinytune.go ./test/

.PHONY: watch
watch: ## run tinytune server and frontend in hot-reload way
	reflex -r '\.(html|go)$\' -s make run & make -C ./web watch

.PHONY: clean
clean: ## clean
	go clean
	rm -rf out/
	rm -f coverage*.out

.PHONY: test
test: ## run server tests
	go test -timeout 2m -race -failfast ./...

.PHONY: ubuntu
ubuntu: ## Install deps for ubuntu (libvips, ffmpeg)
	sudo apt install build-essential libvips pkg-config libvips-dev ffmpeg -y
	npm i --prefix ./web

.PHONY: coverage
coverage: ## coverage
	make test
	go tool cover -html=coverage.out

.PHONY: lint
lint: ## run server linting
	golangci-lint run --fix

.PHONY: quality
quality: ## check-quality
	make fmt
	make lint

.PHONY: fmt
fmt: ## run server prettyfier
	go fmt ./...

$(VERBOSE).SILENT:


define AUTOGEN_HELP_BASH
    declare -A targets; declare -a torder
    targetre='^([A-Za-z]+):.* *# *(.*)'
    if [[ $$TERM && $$TERM != dumb && -t 1 ]]; then
        ul=$$'\e[0;4m'; bbold=$$'\e[34;1m'; reset=$$'\e[0m'
    fi
    if [[ -n "$(TITLE)" ]]; then
        printf "\n  %sMakefile targets - $(TITLE)%s\n\n" "$$ul" "$$reset"
    else
        printf "\n  %sMakefile targets%s\n\n" "$$ul" "$$reset"
    fi
    while read -r line; do
        if [[ $$line =~ $$targetre ]]; then
            target=$${BASH_REMATCH[1]}; help=$${BASH_REMATCH[2]}
            torder+=("$$target")
            targets[$$target]=$$help
            if (( $${#target} > max )); then max=$${#target}; fi
        fi
    done
    for t in "$${torder[@]}"; do
        printf "    %smake %-*s%s   %s\n" "$$bbold" $$max "$$t" "$$reset" \
                                          "$${targets[$$t]}"
    done
    if [[ -n "$(HOMEPAGE)" ]]; then
        printf "\n  Homepage:\n    $(HOMEPAGE)\n\n"
    else
        printf "\n"
    fi
endef
export AUTOGEN_HELP_BASH
