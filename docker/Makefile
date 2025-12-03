# --- Цвета ---
GREEN  := \033[1;32m
RESET  := \033[0m


SCRIPTS_PATH=scripts
UPDATE_SECRETS=$(SCRIPTS_PATH)/update_secrets.sh
COMPOSE=docker-compose.yaml

# 			APP			# 
app-up: # Start the app
	@./$(UPDATE_SECRETS)
	-docker network create reportify_network
	-docker compose -f $(COMPOSE) up -d

app-down: 
	-docker compose -f $(COMPOSE) down

app-update:
	-docker compose -f $(COMPOSE) down
	-docker compose -f $(COMPOSE) up -d

app-logs:
	-docker logs app


# 			DOCKER			# 
containers-status:
	-docker compose -f $(COMPOSE) ps -a

update-app-image:
	-docker compose -f $(COMPOSE) down
	-docker image rm docker-app
	-docker compose -f $(COMPOSE) up -d
	
clear-database-volumes:
	-docker compose -f (COMPOSE) down -v
	



# 			TESTING			# 
srv-test:
	@echo "$(GREEN)[TEST]$(RESET) Running tests..."
	@gotestsum --format=github-actions ./../internal/service

test-all: tools fmt vet lint race vuln
	@echo "$(GREEN)[OK]$(RESET) No problema :)"

tools:
	@echo "$(GREEN)[SETUP]$(RESET) Installing Go tools..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "$(GREEN)[OK]$(RESET) All tools installed into $(TOOLS_DIR)"

fmt:
	@echo "$(GREEN)[TEST]$(RESET) Formatting code..."
	@go fmt ./../...
	@goimports -w .

vet:
	@echo "$(GREEN)[TEST]$(RESET) Running go vet..."
	@go vet ./../...

lint:
	@echo "$(GREEN)[TEST]$(RESET) Running lint..."
	@golangci-lint run ./../...

race:
	@echo "$(GREEN)[TEST]$(RESET) Running race cheking..."
	@go test ./../... -race -count=1

vuln:
	@echo "$(GREEN)[TEST]$(RESET) Running vuln..."
	@govulncheck ./../...



# CI

code-review:
	@cd .. && ./docker/scripts/pre-commit.sh

code-auto-format:
	@cd .. && gofumpt -w $$(find . -name '*.go' -type f)