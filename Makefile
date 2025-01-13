.PHONY: run watch clean test ubuntu coverage lint quality fmt

BINARY_NAME=tinytune
VERSION=0.0.1

COMMIT_HASH=$(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP=$(shell date '+%Y-%m-%dT%H:%M:%S')

LDFLAGS=-ldflags "-X 'main.Version=${VERSION}' -X 'main.CommitHash=${COMMIT_HASH}' -X 'main.BuildTimestamp=${BUILD_TIMESTAMP}' -X 'main.Mode=Production'"

CGO_LDFLAGS=-ljemalloc
CGO_CFLAGS=-fno-builtin-malloc -fno-builtin-calloc -fno-builtin-realloc -fno-builtin-free

GO=CGO_CFLAGS="${CGO_CFLAGS}" CGO_LDFLAGS="${CGO_LDFLAGS}" go

LINUX_AMD64 = out/${BINARY_NAME}_linux_amd64
WEB_ASSETS = web/assets/index.min.js

RUN_FOLDER = ./test/

build: ${LINUX_AMD64}

${WEB_ASSETS}:
	make -C ./web build

${LINUX_AMD64}: ${WEB_ASSETS}
	GOARCH=amd64 GOOS=linux ${GO} build ${LDFLAGS} -o ${LINUX_AMD64} cmd/tinytune/tinytune.go

run: ## run tinytune server
	${GO} run cmd/tinytune/tinytune.go "${RUN_FOLDER}"

watch: ## run tinytune server and frontend in hot-reload way
	reflex -r '\.(html|go)$\' -s make run & make -C ./web watch

clear: ## clean
	go clean
	rm -rf out/
	make -C ./web clear

test: ## run server tests
	go test -timeout 2m -race -failfast ./...

ubuntu: ## Install deps for ubuntu (libvips, ffmpeg)
	sudo apt install build-essential libvips pkg-config libvips-dev libjemalloc-dev ffmpeg -y
	npm i --prefix ./web

lint: ## run server linting
	golangci-lint run --fix

fmt: ## run server prettyfier
	go fmt ./...

quality: fmt lint ## check-quality