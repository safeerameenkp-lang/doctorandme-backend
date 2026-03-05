#!/bin/bash

# Docker Build Retry Script for DrAndMe Backend
# This script handles Docker Hub connectivity issues and provides multiple fallback options

set -e

echo "🐳 Starting Docker build with retry mechanisms..."

# Function to retry Docker build with different strategies
retry_docker_build() {
    local service_name=$1
    local dockerfile_path=$2
    local max_attempts=3
    
    echo "📦 Building $service_name..."
    
    for attempt in $(seq 1 $max_attempts); do
        echo "🔄 Attempt $attempt/$max_attempts for $service_name"
        
        # Strategy 1: Standard build
        if docker build -f "$dockerfile_path" -t "drandme-$service_name" .; then
            echo "✅ Successfully built $service_name on attempt $attempt"
            return 0
        fi
        
        echo "❌ Attempt $attempt failed for $service_name"
        
        if [ $attempt -lt $max_attempts ]; then
            echo "⏳ Waiting 10 seconds before retry..."
            sleep 10
        fi
    done
    
    echo "🚨 All attempts failed for $service_name"
    return 1
}

# Function to build with alternative registry
build_with_alternative_registry() {
    local service_name=$1
    local dockerfile_path=$2
    
    echo "🌐 Trying alternative registry for $service_name..."
    
    # Create a temporary Dockerfile with alternative base image
    local temp_dockerfile="${dockerfile_path}.alt"
    cp "$dockerfile_path" "$temp_dockerfile"
    
    # Replace alpine:3.18 with alternative registry
    sed -i 's|FROM alpine:3.18|FROM registry.k8s.io/alpine:3.18|g' "$temp_dockerfile"
    
    if docker build -f "$temp_dockerfile" -t "drandme-$service_name" .; then
        echo "✅ Successfully built $service_name with alternative registry"
        rm "$temp_dockerfile"
        return 0
    fi
    
    # Try with another alternative
    sed -i 's|FROM registry.k8s.io/alpine:3.18|FROM quay.io/alpine/alpine:3.18|g' "$temp_dockerfile"
    
    if docker build -f "$temp_dockerfile" -t "drandme-$service_name" .; then
        echo "✅ Successfully built $service_name with Quay registry"
        rm "$temp_dockerfile"
        return 0
    fi
    
    rm "$temp_dockerfile"
    return 1
}

# Function to build with local Alpine image
build_with_local_alpine() {
    local service_name=$1
    local dockerfile_path=$2
    
    echo "🏠 Trying to pull Alpine image locally first..."
    
    # Try to pull Alpine image first
    if docker pull alpine:3.18; then
        echo "✅ Successfully pulled Alpine image"
        if docker build -f "$dockerfile_path" -t "drandme-$service_name" .; then
            echo "✅ Successfully built $service_name with local Alpine"
            return 0
        fi
    fi
    
    return 1
}

# Main build function
build_service() {
    local service_name=$1
    local dockerfile_path=$2
    
    echo "🚀 Building $service_name service..."
    
    # Try standard build first
    if retry_docker_build "$service_name" "$dockerfile_path"; then
        return 0
    fi
    
    # Try with local Alpine pull
    if build_with_local_alpine "$service_name" "$dockerfile_path"; then
        return 0
    fi
    
    # Try with alternative registry
    if build_with_alternative_registry "$service_name" "$dockerfile_path"; then
        return 0
    fi
    
    echo "💥 Failed to build $service_name with all strategies"
    return 1
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Build all services
services=(
    "auth-service:services/auth-service/Dockerfile"
    "organization-service:services/organization-service/Dockerfile"
    "appointment-service:services/appointment-service/Dockerfile"
)

failed_services=()

for service_info in "${services[@]}"; do
    IFS=':' read -r service_name dockerfile_path <<< "$service_info"
    
    if ! build_service "$service_name" "$dockerfile_path"; then
        failed_services+=("$service_name")
    fi
done

# Summary
echo ""
echo "📊 Build Summary:"
if [ ${#failed_services[@]} -eq 0 ]; then
    echo "✅ All services built successfully!"
    echo ""
    echo "🚀 You can now run: docker-compose up"
else
    echo "❌ Failed services: ${failed_services[*]}"
    echo ""
    echo "🔧 Troubleshooting suggestions:"
    echo "1. Check your internet connection"
    echo "2. Try running: docker system prune -a"
    echo "3. Restart Docker Desktop"
    echo "4. Try building individual services manually"
    echo ""
    echo "For individual service builds:"
    for service_info in "${services[@]}"; do
        IFS=':' read -r service_name dockerfile_path <<< "$service_info"
        echo "  docker build -f $dockerfile_path -t drandme-$service_name ."
    done
fi


