# BUILDER
FROM golang:1.25.1-alpine AS builder
WORKDIR /grant

COPY go.mod go.sum ./ 
COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6 \
    && go mod download

RUN swag init -g cmd/main.go -o ./docs  \
    && go get github.com/swaggo/swag@v1.16.6 \
    && go mod tidy \
    && go build -o /grant/app ./cmd

# RUNNER
FROM alpine:3.22 AS runner
WORKDIR /root/
COPY --from=builder ./grant .
EXPOSE 8008
CMD ["./app"]