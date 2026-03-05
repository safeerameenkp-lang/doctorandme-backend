#!/bin/bash
# Complete build and run script for microservices

set -e

echo "🚀 Building and Running DrAndMe Microservices"
echo "=============================================="

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Step 1: Stop existing containers
echo -e "${YELLOW}Step 1: Stopping existing containers...${NC}"
docker-compose down

# Step 2: Build all services
echo -e "${YELLOW}Step 2: Building all services...${NC}"
docker-compose build --no-cache

# Step 3: Start all services
echo -e "${YELLOW}Step 3: Starting all services...${NC}"
docker-compose up -d

# Step 4: Wait for services to be ready
echo -e "${YELLOW}Step 4: Waiting for services to be ready...${NC}"
sleep 15

# Step 5: Check service status
echo -e "${YELLOW}Step 5: Checking service status...${NC}"
docker-compose ps

# Step 6: Health checks
echo -e "${YELLOW}Step 6: Running health checks...${NC}"

# Check Kong
if curl -sf http://localhost:8001/status > /dev/null; then
    echo -e "${GREEN}✅ Kong Gateway: Healthy${NC}"
else
    echo -e "${RED}❌ Kong Gateway: Not responding${NC}"
fi

# Check Auth Service
if curl -sf http://localhost:8000/api/auth/health > /dev/null; then
    echo -e "${GREEN}✅ Auth Service: Healthy${NC}"
else
    echo -e "${RED}❌ Auth Service: Not responding${NC}"
fi

# Check Organization Service
if curl -sf http://localhost:8000/api/organizations/health > /dev/null; then
    echo -e "${GREEN}✅ Organization Service: Healthy${NC}"
else
    echo -e "${RED}❌ Organization Service: Not responding${NC}"
fi

# Check Appointment Service
if curl -sf http://localhost:8000/api/v1/health > /dev/null; then
    echo -e "${GREEN}✅ Appointment Service: Healthy${NC}"
else
    echo -e "${RED}❌ Appointment Service: Not responding${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Build and deployment complete!${NC}"
echo ""
echo "Access points:"
echo "  - Kong Gateway: http://localhost:8000"
echo "  - Kong Admin: http://localhost:8001"
echo "  - PgAdmin: http://localhost:5050"
echo ""
echo "View logs: docker-compose logs -f"
echo "Stop services: docker-compose down"

