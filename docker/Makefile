
DEV_COMPOSE=dev/docker-compose-dev.yaml
PROD_COMPOSE=prod/docker-compose-prod.yaml

# 			DEV			# 

dev-start:
	-docker compose -f $(DEV_COMPOSE) up -d

dev-stop:
	-docker compose -f $(DEV_COMPOSE) down

dev-update:
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
	-docker exec -it db psql -U admin -d base



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
	-docker restart nginx
	

check_app_logs:
	-docker logs app

delete_all_volumes:
	-docker volume rm prod_pg_ud_data
	-docker volume rm dev_pg_ud_data
	-docker volume rm dev_pg_rp_data