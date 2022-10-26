FROM golang:1.19-alpine

RUN apk add --no-cache git

RUN go install github.com/githubnemo/CompileDaemon@v1.4.0

COPY ./ /server
WORKDIR /server

RUN go build main.go
ENTRYPOINT CompileDaemon --build="go build main.go" --command="./main"
