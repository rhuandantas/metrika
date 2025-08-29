FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o api main.go
RUN mkdir -p /app/data/db

# Etapa final
FROM scratch
COPY --from=builder /app/api /api
COPY --from=builder /app/data/db /data/db
ENTRYPOINT ["/api"]