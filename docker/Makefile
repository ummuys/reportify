# --- Цвета ---
GREEN  := \033[1;32m
RESET  := \033[0m


SCRIPTS_PATH=scripts
UPDATE_SECRETS=$(SCRIPTS_PATH)/update_secrets.sh
DEV_COMPOSE=dev/docker-compose-dev.yaml
PROD_COMPOSE=prod/docker-compose-prod.yaml

# 			DEV			# 
dev-start:
	@./$(UPDATE_SECRETS)
	-docker network create reportify_network
	-docker compose -f $(DEV_COMPOSE) up -d

dev-stop:
	-docker compose -f $(DEV_COMPOSE) down

dev-update:
	@./$(UPDATE_SECRETS)
	-docker network create reportify_network
	-docker compose -f $(DEV_COMPOSE) down
	-docker image rm dev-app
	-docker compose -f $(DEV_COMPOSE) up -d

dev-check:
	-docker compose -f $(DEV_COMPOSE) ps -a 

dev-update-app:
	-docker compose -f $(DEV_COMPOSE) down app
	-docker image rm dev-app
	-docker compose -f $(DEV_COMPOSE) up -d app

dev-update-front:
	-docker restart nginx

dev-psql:
	-docker exec -it db-report psql -U admin -d report

dev-update-volumes:
	-docker compose -f $(DEV_COMPOSE) down -v
	-docker compose -f $(DEV_COMPOSE) up -d




# 			PROD			# 
prod-start:
	-docker compose -f $(PROD_COMPOSE) up -d

prod-stop:
	-docker compose -f $(PROD_COMPOSE) down

prod-update:
	-docker compose -f $(PROD_COMPOSE) down
	-docker image rm prod-app
	-docker compose -f $(PROD_COMPOSE) up -d

prod-check:
	-docker compose -f $(PROD_COMPOSE) ps -a 

prod-update-app:
	-docker compose -f $(PROD_COMPOSE) down app
	-docker image rm prod-app
	-docker compose -f $(PROD_COMPOSE) up -d app

prod-update-front:
	-docker compose -f $(PROD_COMPOSE) down nginx
	-docker compose -f $(PROD_COMPOSE) up -d nginx

check_app_logs:
	-docker logs app

delete_all_volumes:
	-docker compose -f prod/docker-compose-prod.yaml down -v db-userdata
	-docker compose -f prod/docker-compose-prod.yaml down -v db-report
	-docker volume rm prod_pg_ud_data
	-docker volume rm dev_pg_ud_data
	-docker volume rm dev_pg_rp_data



# HELP
grafana-up:
	-docker compose -f grafana/docker-compose.yaml up -d
grafana-down:
	-docker compose -f grafana/docker-compose.yaml down



# TESTING
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