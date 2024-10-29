ME = $(lastword $(MAKEFILE_LIST))
.DEFAULT_GOAL := help
.PHONY: help
help:  # prints this help
	@bash -c "$$AUTOGEN_HELP_BASH" < $(ME)

BINARY_NAME=tinytune-linux

.PHONY: build
build: ## build
	mkdir -p out/
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME} cmd/tinytune/tinytune.go
	chmod +x 

.PHONY: run
run: ## run
	./out/${BINARY_NAME}

.PHONY: clean
clean: ## clean
	go clean
	rm -rf out/
	rm -f coverage*.out

.PHONY: test
test: ## test
	go test -v -timeout 5m ./... -coverprofile=coverage.out

.PHONY: ubuntu
ubuntu: ## Install deps for ubuntu (libvips, ffmpeg) 
	apt install libvips pkg-config

.PHONY: coverage
coverage: ## coverage
	make test
	go tool cover -html=coverage.out

.PHONY: lint
lint: ## lint
	golangci-lint run --enable-all

.PHONY: quality
quality: ## check-quality
	make lint
	make fmt

.PHONY: fmt
fmt: ## fmt
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
