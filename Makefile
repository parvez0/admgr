# Makefile
SHELL:=/usr/bin/env bash

# Minimum Go version required
GO_VERSION_REQ = 1.20
MARIADB_IMAGE = docker.io/library/mariadb:10.5
MARIADB_PORT = 3306
MARIADB_PASSWORD = password
MARIADB_DB_NAME = test_db
MARIADB_PROD_DB_NAME = admgr
CONTAINER_NAME = "test_mariadb"
SEED_FILE_PATH = ${HOME}/.admgr/mysql
COVERAGE_REPORT_DIR = ./coverage
DOCKER_TAG=$(shell date +'%Y-%m-%d')

default: build

build: pre-checks
	@echo "Installing go dependencies"
	@go get ./...
	@echo "Pulling Mariadb docker image and starting mariadb server on port ${MARIADB_PORT}"
	@docker pull ${MARIADB_IMAGE}
	@docker network create admgr 2> /dev/null || true
	@docker run --rm --network admgr \
		--name ${CONTAINER_NAME} \
		-e MYSQL_ROOT_PASSWORD=${MARIADB_PASSWORD} \
		-e MARIADB_DATABASE=${MARIADB_DB_NAME} \
		-p 3306:${MARIADB_PORT} -d ${MARIADB_IMAGE} 2> /dev/null || true
	docker exec -i \
		${CONTAINER_NAME} \
		mysql -uroot -p${MARIADB_PASSWORD} \
		-e "CREATE SCHEMA IF NOT EXISTS ${MARIADB_DB_NAME}; CREATE SCHEMA IF NOT EXISTS ${MARIADB_PROD_DB_NAME};"

docker-build:
	@docker build -t kiran/admgr:${DOCKER_TAG} -f ./build/package/Dockerfile .

pre-checks:
	@echo "Checking if Docker is installed..."
	@if ! [ -x $$(command -v docker) ]; then \
		echo "Docker is not installed. Please install Docker to continue." ; \
		exit 1 ; \
	fi

	@echo "Checking if Go is installed and is version $(GO_VERSION_REQ) or greater..."
	@if ! [ -x $$(command -v go) ]; then \
		echo "Go is not installed. Please install Go to continue." ; \
		exit 1 ; \
	fi

	@GO_VERSION=$$(go version | awk '{print $$3}' | sed 's/go//'); \
	if [ "$$(printf '%s\n' "$(GO_VERSION_REQ)" "$$GO_VERSION" | sort -V | head -n1)" != "$(GO_VERSION_REQ)" ]; then \
		echo "Go version $$GO_VERSION is not supported. Minimum Go version required is $(GO_VERSION_REQ)." ; \
		exit 1 ; \
	fi

	@echo "Pre-checks complete."

ensure-output-dir:
	@mkdir -p ${COVERAGE_REPORT_DIR}


test: build
	@go test -v ./...

coverage: build ensure-output-dir
	@go test -coverprofile=${COVERAGE_REPORT_DIR}/coverage.out ./...
	@go tool cover -html=${COVERAGE_REPORT_DIR}/coverage.out -o ${COVERAGE_REPORT_DIR}/coverage.html

clean:
	@echo "Cleaning all resources"
	@docker rm -f ${CONTAINER_NAME} 2> /dev/null || true
	@docker network rm admgr 2> /dev/null || true
