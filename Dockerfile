# ─────────────────────────────────────────────────────────
# Dockerfile — Resource Standardization
#
# Build:  docker build -t auction-simulator .
# Run:    docker run --cpus=2 --memory=512m auction-simulator
#
# This ensures every execution uses exactly 2 vCPUs and 512 MB RAM,
# providing a standardized environment regardless of host machine specs.
# ─────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY . .
RUN go build -o auction-simulator .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/auction-simulator .
RUN mkdir -p output

CMD ["./auction-simulator"]
