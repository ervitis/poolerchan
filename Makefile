.PHONY: tests build format install

help: ## Show this help
	@echo "Help"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[36m%-20s\033[93m %s\n", $$1, $$2}'

tests: ## Execute tests
	go test -race -v ./...

build: ## Build the application
	mkdir -p tmp && \
	go build -ldflags "-s -w" -o ./tmp/poolerchan .

format: ## Format files
	go fmt ./...

install: ## Install dependencies
	go mod tidy