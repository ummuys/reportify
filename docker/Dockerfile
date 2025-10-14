FROM golang:1.25.1-alpine


RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6

WORKDIR /grant

COPY go.mod go.sum ./
RUN go mod download


COPY cmd/ ./cmd/
COPY internal/ ./internal/


RUN swag init -g cmd/main.go -o ./docs
RUN go get github.com/swaggo/swag@v1.16.6
RUN go mod tidy
RUN go build -o /grant/app ./cmd

EXPOSE 1337
ENTRYPOINT ["/grant/app"]