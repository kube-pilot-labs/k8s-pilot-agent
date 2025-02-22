FROM golang:1.24.0-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o k8s-pilot-agent ./cmd/pilot/main.go

FROM ubuntu:24.04

RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/k8s-pilot-agent /usr/local/bin/k8s-pilot-agent

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/k8s-pilot-agent"]