FROM golang:1.21-alpine AS builder

WORKDIR /app

# Configure Go proxy settings for better connectivity
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GO111MODULE=on
ENV GONOPROXY=""
ENV GONOSUMDB=""

# Install git and ca-certificates for module downloads
RUN apk add --no-cache git ca-certificates

# Copy shared modules to the correct location
COPY shared/security ./shared-security

# Copy service files
COPY services/auth-service/go.mod ./

# Copy go.sum if it exists, otherwise it will be generated
COPY services/auth-service/go.sum* ./

# Download dependencies with retry mechanism and fallback
RUN go mod download || (sleep 5 && go mod download) || (sleep 10 && go mod download) || go mod download -x

# Copy the rest of the service code
COPY services/auth-service/ .

# Generate go.sum and build the application
RUN go mod tidy
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]