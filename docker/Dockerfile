FROM golang:1.25.1-alpine

WORKDIR /grant

COPY go.mod go.sum ./
RUN go mod download


COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN go mod tidy 
RUN go build -o /grant/app ./cmd

EXPOSE 1337
ENTRYPOINT ["/grant/app"]