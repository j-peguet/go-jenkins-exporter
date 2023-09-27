APP_NAME = go-jenkins-exporter
DOCKER_IMAGE_TAG = latest

.PHONY: help
help:
	@echo "Simple golang devenv helper"
	@echo ""
	@echo "Usage:"
	@echo "  deps		to install and update dependencies"
	@echo "  install	to install the app in $$GOPATH/bin"
	@echo "  image		to build docker image for the exporter"
	@echo "  help		to show this help"
	@echo ""

.PHONY: deps
deps:
	export GO111MODULE=on
	go get -u -v

.PHONY: install
install:
	go build -o $$GOPATH/bin/$(APP_NAME) main.go

.PHONY: image
image:
	docker build -t $(APP_NAME):$(DOCKER_IMAGE_TAG) .

.PHONY: run
run:
	$$GOPATH/bin/$(APP_NAME) -h
