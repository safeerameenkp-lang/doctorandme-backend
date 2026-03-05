#!/bin/bash
# Initial setup script for Kong Gateway repository

set -e

echo "🔧 Setting up Kong Gateway Repository"
echo "======================================"

# Create .env from example if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "✅ .env file created. Please edit it with your configuration."
else
    echo "✅ .env file already exists"
fi

# Create migrations directories if they don't exist
echo "📁 Creating migrations directories..."
mkdir -p migrations/auth-service
mkdir -p migrations/organization-service
mkdir -p migrations/appointment-service

echo "✅ Migrations directories created"
echo ""
echo "📋 Next steps:"
echo "1. Edit .env file with your configuration"
echo "2. Copy migration files to migrations/ directories (optional)"
echo "3. Run: docker-compose up -d"
echo ""
echo "✅ Setup complete!"

