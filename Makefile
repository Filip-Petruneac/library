up:
	docker-compose up -d

down:
	docker-compose down -v

up-all:
	COMPOSE_PROFILES=all make down up