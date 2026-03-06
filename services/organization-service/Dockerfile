FROM golang:1.24-alpine AS builder

# Configure Go proxy settings for better connectivity
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GO111MODULE=on

# Install git and ca-certificates for dependency fetching with retry
# Try installing first (may use cached packages), then update if needed
RUN apk add --no-cache git ca-certificates || \
    (apk update --no-cache && apk add --no-cache git ca-certificates) || \
    (sleep 5 && apk update --no-cache && apk add --no-cache git ca-certificates) || \
    (sleep 10 && apk update --no-cache && apk add --no-cache git ca-certificates)

WORKDIR /app

# Copy go.mod and go.sum first (for better Docker layer caching)
COPY services/organization-service/go.mod services/organization-service/go.sum* ./

# Download dependencies with retry mechanism and fallback
RUN go mod download || (sleep 5 && go mod download) || (sleep 10 && go mod download) || go mod download -x

# Copy the rest of the service code
COPY services/organization-service/ .

# Generate go.sum and build the application (skip tidy if network fails, use existing go.sum)
RUN go mod tidy || echo "Warning: go mod tidy failed, using existing go.sum"
RUN go build -o main .

# Use a more recent Alpine version with better network handling
FROM alpine:3.19

# Add network timeout configurations and retry mechanisms
# Add network timeout configurations, retry mechanisms, and postgresql-client for migrations
RUN apk --no-cache add ca-certificates postgresql-client && \
    echo "network_timeout=300" >> /etc/apk/repositories && \
    echo "network_retries=3" >> /etc/apk/repositories
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8081
CMD ["./main"]
