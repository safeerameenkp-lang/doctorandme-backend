FROM golang:1.21-alpine AS builder

# Install git and ca-certificates for better network reliability
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy shared modules to the correct location
COPY shared/security ./shared-security

# Copy service go.mod first
COPY services/appointment-service/go.mod ./

# Copy go.sum if it exists, otherwise it will be generated
COPY services/appointment-service/go.sum* ./

# Set Go proxy and other environment variables for better reliability
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV CGO_ENABLED=0
ENV GOOS=linux

# Download dependencies with retry mechanism and timeout handling
RUN go mod download -x || (sleep 5 && go mod download -x) || (sleep 10 && go mod download -x) || (sleep 30 && go mod download -x)

# Copy the rest of the service code
COPY services/appointment-service/ .

# Generate go.sum and build the application with timeout handling
RUN timeout 300 go mod tidy || go mod tidy
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8082
CMD ["./main"]
