# Stage 1: Build


FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# Install tools
RUN apk add --no-cache git build-base curl

# Install sqlc
RUN curl -sSL https://github.com/sqlc-dev/sqlc/releases/download/v1.25.0/sqlc_1.25.0_linux_amd64.tar.gz | tar -xz -C /usr/local/bin

# Copy dependencies first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the full project
COPY . .

# 🔄 Run sqlc generate in db folder
WORKDIR /app/db
RUN sqlc generate

# 📦 Build Go binaries
WORKDIR /app
RUN go build -o server .
RUN go build -o dbinit ./temp/dbinit.go

# Stage 2: Runtime container
FROM alpine:latest

WORKDIR /app

# Copy binaries and env file only
COPY --from=builder /app/server .
COPY --from=builder /app/dbinit .
COPY --from=builder /app/.env .
COPY --from=builder /app/entrypoint.sh ./entrypoint.sh


RUN chmod +x ./entrypoint.sh


EXPOSE 8000
CMD ["./entrypoint.sh"]
