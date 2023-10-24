APP_NAME := app.exe
BUILD_ROUTE := ./bin/${APP_NAME}
VERSION_TAG := latest
NEW_VERSION_TAG := 0.2.1
IMAGE_NAME := apooravm/multi_serve
IMAGE_NAME_WITH_TAG := ${IMAGE_NAME}:${VERSION_TAG}

# docker pull, build, run = create new container + start the container
postgrescontainer:
	@docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=1234 -d postgres:15-alpine

echoinstall:
	@go get github.com/labstack/echo/v4

# Building docker image
dockerbuild:
	@docker build -t ${IMAGE_NAME_WITH_TAG} .

# change the tag name??
# The docker tag command in Docker is used to create a new tag for an existing Docker image. 
# docker tag SOURCE_IMAGE[:TAG] TARGET_IMAGE[:TAG]
# docker tag myname/myimage:1.0 myname/myimage:latest
dockertagchange:
	@docker tag ${IMAGE_NAME_WITH_TAG} ${IMAGE_NAME}:${NEW_VERSION_TAG}

# push to dockerhub
dockerpush:
	@docker push ${IMAGE_NAME_WITH_TAG}

# Creating container from image and starting it
# multi-serve-container
dockerrun:
	@docker run --name ms-cont -p 5000:5000 -e PORT=5000 -d ${IMAGE_NAME_WITH_TAG}

awssdk:
	@go get github.com/aws/aws-sdk-go-v2
	@go get github.com/aws/aws-sdk-go-v2/config
	@go get github.com/aws/aws-sdk-go-v2/service/s3

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

dev: build
	@${BUILD_ROUTE}