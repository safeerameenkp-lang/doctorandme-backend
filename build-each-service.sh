#!/bin/bash
# Build Each Microservice Individually (Bash)

SERVICE=${1:-all}

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}🔨 Building Microservices Individually${NC}"
echo "======================================="
echo ""

build_service() {
    local service_name=$1
    local display_name=$2
    
    echo -e "${YELLOW}Building $display_name...${NC}"
    docker-compose build $service_name
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ $display_name built successfully!${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}❌ $display_name build failed!${NC}"
        echo ""
        return 1
    fi
}

# Build based on argument
case $SERVICE in
    auth)
        build_service "auth-service" "Auth Service"
        ;;
    organization)
        build_service "organization-service" "Organization Service"
        ;;
    appointment)
        build_service "appointment-service" "Appointment Service"
        ;;
    all)
        echo -e "${CYAN}Building all services...${NC}"
        echo ""
        
        build_service "auth-service" "Auth Service"
        build_service "organization-service" "Organization Service"
        build_service "appointment-service" "Appointment Service"
        
        echo "======================================="
        echo -e "${CYAN}Build Summary:${NC}"
        echo ""
        echo -e "${GREEN}✅ All services built!${NC}"
        echo ""
        echo -e "${CYAN}Built images:${NC}"
        docker images | grep drandme
        ;;
    *)
        echo -e "${RED}Invalid service: $SERVICE${NC}"
        echo "Usage: $0 [auth|organization|appointment|all]"
        exit 1
        ;;
esac

echo ""
echo -e "${CYAN}Next steps:${NC}"
echo "  - Start services: docker-compose up -d"
echo "  - View logs: docker-compose logs -f"
echo "  - Check status: docker-compose ps"

