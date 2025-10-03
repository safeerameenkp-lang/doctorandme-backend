FROM golang:1.21-alpine AS builder

# Install git and ca-certificates for dependency fetching
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy shared modules to the correct location first
COPY shared/security ./shared-security

# Copy service files
COPY services/organization-service/go.mod ./

# Download dependencies (now that shared module is available)
RUN go mod download

# Copy the rest of the service code
COPY services/organization-service/ .

# Generate go.sum and build the application
RUN go mod tidy
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8081
CMD ["./main"]
