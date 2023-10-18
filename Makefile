APP_NAME := app.exe
BUILD_ROUTE := ./bin/${APP_NAME}
# docker pull, build, run = create new container + start the container

postgrescontainer:
	@docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=1234 -d postgres:15-alpine

echoinstall:
	@go get github.com/labstack/echo/v4

# Building docker image
dockerbuild:
	@docker build -t uninote .

# Creating container from image and starting it
dockerrun:
	@docker run --name uninote-container -p 5000:5000 -e PORT=5000 -d uninote

# dep install
install:
	@go mod download

build:
	@go build -o ${BUILD_ROUTE} ./src/main.go

tidy:
	@go mod tidy

vendor: tidy
	@go mod vendor

run: vendor build
	@${BUILD_ROUTE}

dev:
	@${BUILD_ROUTE}