ENV := $(PWD)/.env
include $(ENV)

mysql:
	@sudo docker run -d --name mysql-container \
		-e MYSQL_ROOT_PASSWORD=$(MYSQL_ROOT_PASSWORD) \
		-e MYSQL_DATABASE=$(MYSQL_DATABASE) \
		-e MYSQL_USER=$(MYSQL_USER) \
		-e MYSQL_PASSWORD=$(MYSQL_PASSWORD) \
		-p 4450:$(MYSQL_PORT) \
		-v $(PWD)/library-project/schema.sql:/docker-entrypoint-initdb.d/schema.sql \
		mysql:latest

.PHONY: mysql