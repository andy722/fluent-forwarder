OS=linux
ARCH=amd64
BUILD_DIR=.build
TARGET="$(BUILD_DIR)/fluent-forwarder-${OS}-${ARCH}"

.PHONY: help
help: ## This help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
.DEFAULT_GOAL := help

build: ## Build application
	@echo "Building: $(TARGET)"
	docker run \
      --rm \
      -v $(CURDIR):/usr/src/app \
      -w /usr/src/app \
      fc-builder \
      sh -c "env GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -v -ldflags='-s -w' -i -o $(TARGET)"

builder: ## Create image used for application building
	cd .scripts && docker build -t fc-builder .

update: ## Roll out to production
	cd .scripts && ansible-playbook ansible.update-prod.yaml