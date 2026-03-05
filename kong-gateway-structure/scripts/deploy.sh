#!/bin/bash
# Deployment script for Kong Gateway and all services

set -e

echo "🚀 Deploying DrAndMe Microservices with Kong Gateway"
echo "=================================================="

# Check if .env exists
if [ ! -f .env ]; then
    echo "❌ Error: .env file not found!"
    echo "Please copy .env.example to .env and configure it."
    exit 1
fi

# Load environment variables
source .env

# Check required variables
if [ -z "$DB_PASSWORD" ] || [ -z "$JWT_ACCESS_SECRET" ] || [ -z "$JWT_REFRESH_SECRET" ]; then
    echo "❌ Error: Required environment variables not set!"
    echo "Please configure DB_PASSWORD, JWT_ACCESS_SECRET, and JWT_REFRESH_SECRET in .env"
    exit 1
fi

echo "✅ Environment variables loaded"

# Pull latest images (if using pre-built)
if [ ! -z "$AUTH_SERVICE_IMAGE" ]; then
    echo "📥 Pulling service images..."
    docker pull $AUTH_SERVICE_IMAGE || true
    docker pull $ORG_SERVICE_IMAGE || true
    docker pull $APPT_SERVICE_IMAGE || true
fi

# Start services
echo "🚀 Starting services..."
docker-compose up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be ready..."
sleep 10

# Health checks
echo "🏥 Checking service health..."
curl -f http://localhost:8000/api/auth/health && echo "✅ Auth service healthy" || echo "❌ Auth service not responding"
curl -f http://localhost:8000/api/organizations/health && echo "✅ Organization service healthy" || echo "❌ Organization service not responding"
curl -f http://localhost:8000/api/v1/health && echo "✅ Appointment service healthy" || echo "❌ Appointment service not responding"

echo ""
echo "🎉 Deployment complete!"
echo "Kong Gateway: http://localhost:8000"
echo "Kong Admin: http://localhost:8001"

