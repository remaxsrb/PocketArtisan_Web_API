# ---------- BUILDER ----------
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/

# ---------- RUNTIME ----------
FROM --platform=linux/amd64 debian:bookworm-slim

# Build args for host UID/GID
ARG UID=1000
ARG GID=1000

# Create users matching host
RUN groupadd -g $GID appuser \
    && useradd -m -u $UID -g $GID appuser

WORKDIR /app

# Optional: install CA certificates for HTTPS
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Create uploads folder and assign ownership
RUN mkdir -p /app/uploads \
    && chown -R $UID:$GID /app/uploads

# Copy binary and make executable
COPY --from=builder /app/server .
RUN chmod +x /app/server

# Copy .env file
#COPY --from=builder /app/.env .

# Switch to non-root users
USER $UID:$GID

ENV GIN_MODE=release
EXPOSE 8080

ENTRYPOINT ["./server"]
