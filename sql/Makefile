ENV := $(PWD)/.env
include $(ENV)

mysql:
	@docker run -d --name mysql-container \
		--restart always \
		-e MYSQL_ROOT_PASSWORD=$(MYSQL_ROOT_PASSWORD) \
		-e MYSQL_DATABASE=$(MYSQL_DATABASE) \
		-e MYSQL_USER=$(MYSQL_USER) \
		-e MYSQL_PASSWORD=$(MYSQL_PASSWORD) \
		-p 4450:$(MYSQL_PORT) \
		-v $(PWD)/dump.sql:/docker-entrypoint-initdb.d/dump.sql \
		mysql:latest

.PHONY: 
