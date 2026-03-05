# System Recovery Instructions

## Issue Detected
The previous configuration for environment variables (`DB_HOST: ${DB_HOST:-postgres}`) caused a startup failure with the error: `DB ping error: dial tcp: lookup port=: no such host`.

This happened because the `DB_HOST` variable was resolving to an empty string in your specific Docker environment, causing the application to try and connect to a host literally named `port=` (due to malformed connection string construction from empty variables).

## Fix Applied
I have reverted the **non-sensitive** configuration values (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_NAME`) to hardcoded defaults in `docker-compose.yml` to guarantee stability.

- **DB_HOST**: `postgres` (Internal Docker DNS name)
- **DB_PORT**: `5432`
- **DB_USER**: `postgres`
- **DB_NAME**: `drandme`

**Sensitive variables** like `DB_PASSWORD` and `JWT_SECRETS` still use the secure variable substitution syntax (`${VAR:-default}`) so they can be overridden by your `.env` file or CI/CD system.

## Action Required
To restart the system with the fixed configuration:

1.  **Stop the crash-looping containers**:
    ```powershell
    docker-compose down
    ```

2.  **Rebuild and Start**:
    ```powershell
    docker-compose up -d --build
    ```

3.  **Verify**:
    The error `lookup port=: no such host` should now be gone, and services should connect to the `postgres` container successfully.
