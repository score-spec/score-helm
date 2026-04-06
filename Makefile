# Disable all the default make stuff
MAKEFLAGS += --no-builtin-rules
.SUFFIXES:

## Display a list of the documented make targets
.PHONY: help
help:
	@echo Documented Make targets:
	@perl -e 'undef $$/; while (<>) { while ($$_ =~ /## (.*?)(?:\n# .*)*\n.PHONY:\s+(\S+).*/mg) { printf "\033[36m%-30s\033[0m %s\n", $$2, $$1 } }' $(MAKEFILE_LIST) | sort

.PHONY: .FORCE
.FORCE:

build:
	go build ./cmd/score-helm/

test:
	go vet ./...
	go env -w GOTOOLCHAIN="$(shell go env GOVERSION)+auto"
	go test ./... -cover -race

test-app: build
	./score-helm --version
	./score-helm init
	cat score.yaml
	./score-helm generate score.yaml
	cat values.yaml

build-container:
	docker build -t score-helm:local .

test-container: build-container
	docker run --rm score-helm:local --version
	docker run --rm -v .:/score-helm score-helm:local init
	cat score.yaml
	docker run --rm -v .:/score-helm score-helm:local generate score.yaml
	cat values.yaml