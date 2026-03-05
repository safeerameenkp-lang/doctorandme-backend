#!/bin/bash
# Health check script for all services

echo "🏥 Health Check - DrAndMe Services"
echo "==================================="

# Kong Gateway
echo -n "Kong Gateway: "
if curl -sf http://localhost:8001/status > /dev/null; then
    echo "✅ Healthy"
else
    echo "❌ Not responding"
fi

# Auth Service
echo -n "Auth Service: "
if curl -sf http://localhost:8000/api/auth/health > /dev/null; then
    echo "✅ Healthy"
else
    echo "❌ Not responding"
fi

# Organization Service
echo -n "Organization Service: "
if curl -sf http://localhost:8000/api/organizations/health > /dev/null; then
    echo "✅ Healthy"
else
    echo "❌ Not responding"
fi

# Appointment Service
echo -n "Appointment Service: "
if curl -sf http://localhost:8000/api/v1/health > /dev/null; then
    echo "✅ Healthy"
else
    echo "❌ Not responding"
fi

# Database
echo -n "PostgreSQL: "
if docker exec drandme-postgres pg_isready -U postgres > /dev/null 2>&1; then
    echo "✅ Healthy"
else
    echo "❌ Not responding"
fi

echo ""
echo "✅ Health check complete!"

