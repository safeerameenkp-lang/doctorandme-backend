-- 003_auth_performance_indexing.sql
-- Optimizing Auth Service for high-scale low-latency architecture (>5000 req/sec)

-- 1. Optimize Login Lookups (email, username, phone)
-- We use a composite index targeting the WHERE clause in Login (matching identifier + is_active + is_blocked)
CREATE INDEX IF NOT EXISTS idx_users_login_perf ON users (username, email, phone) WHERE is_active = true AND is_blocked = false;

-- 2. Optimize Registration Existence Checks
-- High-throughput creation needs fast unique collision checks without full table scans
CREATE INDEX IF NOT EXISTS idx_users_email_exists ON users(email) WHERE email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_username_exists ON users(username);

-- 3. Optimize Refresh Token Verifications
-- Fast mapping for the FOR UPDATE SKIP LOCKED row checks (Token verification / invalidation)
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_verification ON refresh_tokens (user_id, token) WHERE expires_at > CURRENT_TIMESTAMP AND revoked_at IS NULL;

-- 4. Optimize Role Based Access Control (RBAC) - Extremely Critical
-- Every single API request passes through the middleware which rapidly queries user_roles and roles.
-- This needs a covering composite index so Postgres can read it straight from memory without touching disk pages.
CREATE INDEX IF NOT EXISTS idx_user_roles_active_lookup ON user_roles (user_id, role_id) INCLUDE (organization_id, clinic_id, service_id) WHERE is_active = true;

-- 5. Optimize Role Name Lookups
CREATE INDEX IF NOT EXISTS idx_roles_name ON roles (name);
