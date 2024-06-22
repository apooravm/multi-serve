FROM golang:1.21.3-alpine

WORKDIR /multi-mux

COPY . .

RUN go mod download && go mod tidy && go mod vendor && go build -o ./bin/app.exe ./src/main.go

ENV PORT=4000

# EXPOSE 4000

CMD ["./bin/app.exe"]
