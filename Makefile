up:
	docker-compose up -d

down:
	docker-compose down -v

up-all:
	COMPOSE_PROFILES=all make down up

# Start MySQL container for local development
mysql: 
	docker run -d --name mysql-container \
		--restart always \
		--env-file .env.sql.local \
		-p 4450:3306 \
		-v ./sql/dump.sql:/docker-entrypoint-initdb.d/dump.sql \
		mysql:latest

stop:
	docker stop mysql-container || true
	docker rm mysql-container || true

.PHONY: up down up-all mysql stop
