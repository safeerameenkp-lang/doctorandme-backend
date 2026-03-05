FROM golang:1.21-alpine AS builder

# Install git and ca-certificates for better network reliability with retry
# Try installing first (may use cached packages), then update if needed
RUN apk add --no-cache git ca-certificates tzdata || \
    (apk update --no-cache && apk add --no-cache git ca-certificates tzdata) || \
    (sleep 5 && apk update --no-cache && apk add --no-cache git ca-certificates tzdata) || \
    (sleep 10 && apk update --no-cache && apk add --no-cache git ca-certificates tzdata)

WORKDIR /app

# Set Go proxy and other environment variables for better reliability
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV CGO_ENABLED=0
ENV GOOS=linux

# Copy go.mod and go.sum first (for better Docker layer caching)
COPY services/appointment-service/go.mod services/appointment-service/go.sum* ./

# Download dependencies with retry mechanism and timeout handling
RUN go mod download -x || (sleep 5 && go mod download -x) || (sleep 10 && go mod download -x) || (sleep 30 && go mod download -x)

# Copy the rest of the service code
COPY services/appointment-service/ .

# Generate go.sum and build the application with timeout handling
RUN timeout 300 go mod tidy || go mod tidy
RUN go build -o main .

# Use a more recent Alpine version with better network handling
FROM alpine:3.19

# Add network timeout configurations and retry mechanisms
RUN apk --no-cache add ca-certificates && \
    echo "network_timeout=300" >> /etc/apk/repositories && \
    echo "network_retries=3" >> /etc/apk/repositories
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8082
CMD ["./main"]
