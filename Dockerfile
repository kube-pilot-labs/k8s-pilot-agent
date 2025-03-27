FROM golang:1.24.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY pkg/ ./pkg/

RUN CGO_ENABLED=0 GOOS=linux go build -o k8s-pilot-agent ./cmd/pilot/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/k8s-pilot-agent /usr/local/bin/k8s-pilot-agent

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/k8s-pilot-agent"]