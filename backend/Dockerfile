FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /mossbase-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /mossbase-mcp-server ./cmd/mcp-server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /mossbase-server /usr/local/bin/
COPY --from=builder /mossbase-mcp-server /usr/local/bin/
COPY configs/config.yaml /app/configs/config.yaml
WORKDIR /app
EXPOSE 8033
CMD ["mossbase-server"]
