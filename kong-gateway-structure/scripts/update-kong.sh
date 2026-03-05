#!/bin/bash
# Script to update Kong configuration without restarting all services

set -e

echo "🔄 Updating Kong Configuration"
echo "============================="

# Validate kong.yml
echo "✅ Validating kong.yml..."
docker exec drandme-kong kong config -c /kong/kong.yml validate || {
    echo "❌ kong.yml validation failed!"
    exit 1
}

# Reload Kong
echo "🔄 Reloading Kong configuration..."
docker exec drandme-kong kong reload || {
    echo "❌ Failed to reload Kong"
    exit 1
}

echo "✅ Kong configuration updated successfully!"

