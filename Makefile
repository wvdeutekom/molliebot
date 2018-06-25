SHELL := /bin/sh

.PHONY: dev

dev:
	@go build -o molliebot && source ./platforms/dev-config && ./molliebot