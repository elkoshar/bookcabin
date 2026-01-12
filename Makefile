#!/bin/bash
export NOW=$(shell date +"%Y/%m/%d %T")
export REPO_NAME=bookcabing
# version based tags
IMG_TAG ?= ${shell git rev-parse --short HEAD}

swag:
	@swag init --parseDependency --parseInternal --parseDepth 2 -g cmd/http/main.go

build: swag
	@echo "${NOW} == Building HTTP Server"
	@go build -o ./bin/${REPO_NAME}-http cmd/http/main.go 

run-http: build 
	@./bin/${REPO_NAME}-http

build-image-http:
	@ echo "Building Dockerfile.http image for version ${IMG_TAG}"
	@ docker build -f Dockerfile.http -t ${REPO_NAME}-http:${IMG_TAG} .

docker-run-http:
	@ docker run --env GO_ENV=$(GO_ENV) -p 8812:8812 --name ${REPO_NAME}-http ${REPO_NAME}-http:${IMG_TAG} 

clean-mod-cache:
	@go clean -cache -modcache -i -r

test:
	@go test github.com/elkoshar/bookcabing/...

test-debug:
	@go test ./... -v | grep FAIL
