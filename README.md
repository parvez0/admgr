# ADMGR Service

The ADMGR service provides these APIs to manage ad slots effectively. It allows creating new slots, updating existing slots, retrieving slot information, reserving open slots, and closing open slots as needed.

## API Endpoints
Please refer to the API documentation or specifications for detailed request and response formats, authentication requirements, and error handling.


## Prerequisites
- Go version 1.20 or greater
- Docker (to run MariaDB)

## Setup
- Install Go and Docker on your system if they are not already installed.
- Clone the admgr repository.
- Navigate to the root directory of the repository.

## Configuration
- Make sure the GO_VERSION_REQ variable in the Makefile is set to the minimum required Go version.
- Modify the MARIADB_IMAGE, MARIADB_PORT, MARIADB_PASSWORD, MARIADB_DB_NAME, MARIADB_PROD_DB_NAME, CONTAINER_NAME, SEED_FILE_PATH, COVERAGE_REPORT_DIR, and DOCKER_TAG variables in the Makefile according to your needs.

### Building the App
To build the Manager app, run the following command:

```shell
make build
```

This command installs the Go dependencies, pulls the MariaDB Docker image, and starts the MariaDB server.

### Running Tests

To run tests for the SlotManager app, execute the following command:
```shell
make clean test
```
This command runs the tests and displays the output.

Note. Before running the test make to sure to start the accounting dummy service
```shell
go run ./stubs/account/server.go
```

### Cleaning Up

To clean up all resources created by the SlotManager app, run the following command:

```shell
make clean
```
This command stops and removes the MariaDB Docker container and deletes the Docker network.

## Building Application Docker Image
To build a Docker image for the Manager app, use the following command:

```shell
make docker-build
```
The Docker image will be tagged with the current date.

This README provides an overview of the Manager app and the steps to build, test, and clean up the application.