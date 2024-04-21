APP_NAME := app.exe
BUILD_ROUTE := ./bin/${APP_NAME}
VERSION_TAG := latest
OLD_VERSION_TAG := 0.2.9
IMAGE_NAME := apooravm/multi_serve
IMAGE_NAME_WITH_TAG := ${IMAGE_NAME}:${VERSION_TAG}
CURR_DATE_TIME := @powershell -Command "Get-Date -Format 'dd MMMM yyyy HH:mm'"

# Docker flow
# Change the current image tag from latest to vNum
# docker build => push

.PHONY: git

init:
	@mkdir data/S3

# docker pull, build, run = create new container + start the container
postgrescontainer:
	@docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=1234 -d postgres:15-alpine

depinstall:
	@go get github.com/labstack/echo/v4
	@go get github.com/gorilla/websocket

# Building docker image
dockerbuild:
	@docker build -t ${IMAGE_NAME_WITH_TAG} .

# change the tag name??
# The docker tag command in Docker is used to create a new tag for an existing Docker image. 
# docker tag SOURCE_IMAGE[:TAG] TARGET_IMAGE[:TAG]
# docker tag myname/myimage:1.0 myname/myimage:latest
dockertagchange:
	@docker tag ${IMAGE_NAME_WITH_TAG} ${IMAGE_NAME}:${OLD_VERSION_TAG}

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

git:
	@git add . && git commit -m "More S3 Routing and Log maintenance, ${CURR_DATE_TIME}" && git push origin main

date:
	${CURR_DATE_TIME}

# dep install
install:
	@go mod download

build:
	@echo "building..."
	@go build -o ${BUILD_ROUTE} ./src/main.go

tidy:
	@echo "tidying up..."
	@go mod tidy

vendor: tidy
	@go mod vendor

run: vendor build
	@${BUILD_ROUTE} dev

dev: build
	@echo "Running the binary..."
	@${BUILD_ROUTE} dev
