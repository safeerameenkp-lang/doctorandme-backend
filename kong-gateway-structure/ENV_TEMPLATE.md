# Environment Variables Template

Create a `.env` file in the root directory with these variables:

```env
# Database Configuration
DB_NAME=drandme
DB_USER=postgres
DB_PASSWORD=your-strong-password-here
DB_PORT=5432

# JWT Secrets (MUST be same across all services)
JWT_ACCESS_SECRET=your-strong-access-secret-key-here-min-32-chars
JWT_REFRESH_SECRET=your-strong-refresh-secret-key-here-min-32-chars

# Kong Configuration
KONG_PROXY_PORT=8000
KONG_ADMIN_PORT=8001

# Docker Image Registry (if using pre-built images)
AUTH_SERVICE_IMAGE=your-registry/drandme-auth-service:latest
ORG_SERVICE_IMAGE=your-registry/drandme-organization-service:latest
APPT_SERVICE_IMAGE=your-registry/drandme-appointment-service:latest

# PgAdmin (Optional)
PGADMIN_EMAIL=admin@drandme.com
PGADMIN_PASSWORD=your-pgadmin-password-here
PGADMIN_PORT=5050
```

## Usage

1. Copy this template to `.env`
2. Replace placeholder values with your actual secrets
3. **DO NOT commit `.env` to Git!**

