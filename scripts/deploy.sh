#!/bin/bash
set -euo pipefail

# Parameters from workflow environment variables
BINARY_NAME="${BINARY_NAME:-uninaquiz-backend}"
DEPLOY_PATH="${DEPLOY_PATH:-/opt/uninaquiz}"
TEMP_DIR="/tmp"

# Load local environment variables from .env to discover the actual port
ENV_FILE="${DEPLOY_PATH}/.env"
APP_PORT=8080

if [ -f "$ENV_FILE" ]; then
    PORT_IN_ENV=$(grep -E '^APP_PORT=' "$ENV_FILE" | cut -d'=' -f2 || true)
    if [ -n "$PORT_IN_ENV" ]; then
        APP_PORT="$PORT_IN_ENV"
    fi
fi

echo ">>> Starting deployment of $BINARY_NAME on port $APP_PORT <<<"

# 1. Validate that the temporary binary exists
if [ ! -f "${TEMP_DIR}/${BINARY_NAME}" ]; then
    echo "Error: New binary not found at ${TEMP_DIR}/${BINARY_NAME}"
    exit 1
fi

# 2. Back up current binary
echo "Backing up current binary..."
if [ -f "${DEPLOY_PATH}/${BINARY_NAME}" ]; then
    cp "${DEPLOY_PATH}/${BINARY_NAME}" "${DEPLOY_PATH}/${BINARY_NAME}.old"
else
    echo "No previous version found. Initial deploy."
    touch "${DEPLOY_PATH}/${BINARY_NAME}.old"
fi

# 3. Swap binaries (atomic move to minimize offline time)
echo "Replacing old binary with new one..."
mv "${TEMP_DIR}/${BINARY_NAME}" "${DEPLOY_PATH}/${BINARY_NAME}"
chown appuser:appuser "${DEPLOY_PATH}/${BINARY_NAME}"
chmod +x "${DEPLOY_PATH}/${BINARY_NAME}"

# 4. Clean up temporary deployment script if present
rm -f "${TEMP_DIR}/deploy.sh"

# 5. Restart systemd service
echo "Restarting systemd service..."
sudo systemctl daemon-reload
sudo systemctl restart uninaquiz

# 6. Validate (local health check)
HEALTH_URL="http://localhost:${APP_PORT}/api/health"
echo "Waiting for application startup..."
sleep 3

echo "Running health check on ${HEALTH_URL}..."
HEALTH_CHECK_PASSED=false

for i in {1..10}; do
    if curl -sf "$HEALTH_URL" > /dev/null; then
        echo "Health check passed successfully!"
        HEALTH_CHECK_PASSED=true
        break
    fi
    echo "Attempt $i failed. Retrying in 2 seconds..."
    sleep 2
done

# 7. Rollback if validation fails
if [ "$HEALTH_CHECK_PASSED" = false ]; then
    echo "!!! CRITICAL: Health check failed after deployment !!!"
    echo "Service logs:"
    sudo journalctl -u uninaquiz -n 25 --no-pager

    if [ -f "${DEPLOY_PATH}/${BINARY_NAME}.old" ] && [ -s "${DEPLOY_PATH}/${BINARY_NAME}.old" ]; then
        echo ">>> Executing ROLLBACK to the previous working version <<<"
        mv "${DEPLOY_PATH}/${BINARY_NAME}.old" "${DEPLOY_PATH}/${BINARY_NAME}"
        chown appuser:appuser "${DEPLOY_PATH}/${BINARY_NAME}"
        chmod +x "${DEPLOY_PATH}/${BINARY_NAME}"
        
        sudo systemctl restart uninaquiz
        echo "Rollback complete. Service restored."
    else
        echo "Error: Previous working backup not found."
    fi
    exit 1
fi

# 8. Success: clean up temporary backup file
echo "Deployment successful."
rm -f "${DEPLOY_PATH}/${BINARY_NAME}.old"
