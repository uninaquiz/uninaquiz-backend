#!/bin/bash
set -euo pipefail

BINARY_NAME="${BINARY_NAME:-uninaquiz-backend}"
DEPLOY_PATH="${DEPLOY_PATH:-/opt/uninaquiz}"
TEMP_DIR="/tmp"

ENV_FILE="${DEPLOY_PATH}/.env"
APP_PORT=8080

echo ">>> Starting deployment of $BINARY_NAME <<<"

# -----------------------------
# 1. Validate binary exists
# -----------------------------
if [ ! -f "${TEMP_DIR}/${BINARY_NAME}" ]; then
    echo "ERROR: Binary not found in ${TEMP_DIR}"
    exit 1
fi

# -----------------------------
# 2. Read env (safe fallback)
# -----------------------------
if [ -f "$ENV_FILE" ]; then
    PORT_IN_ENV=$(grep -E '^APP_PORT=' "$ENV_FILE" | cut -d'=' -f2 || true)
    if [ -n "${PORT_IN_ENV:-}" ]; then
        APP_PORT="$PORT_IN_ENV"
    fi
fi

echo "Using APP_PORT=$APP_PORT"

# -----------------------------
# 3. Backup current binary
# -----------------------------
echo "Backing up current binary..."

if [ -f "${DEPLOY_PATH}/${BINARY_NAME}" ]; then
    cp "${DEPLOY_PATH}/${BINARY_NAME}" "${DEPLOY_PATH}/${BINARY_NAME}.old"
else
    echo "No previous binary found (first deploy)"
    touch "${DEPLOY_PATH}/${BINARY_NAME}.old"
fi

# -----------------------------
# 4. Deploy new binary
# -----------------------------
echo "Deploying new binary..."

mv "${TEMP_DIR}/${BINARY_NAME}" "${DEPLOY_PATH}/${BINARY_NAME}"
chmod 755 "${DEPLOY_PATH}/${BINARY_NAME}"

# -----------------------------
# 5. Cleanup temp script
# -----------------------------
rm -f "${TEMP_DIR}/deploy.sh" || true

# -----------------------------
# 6. Restart service
# -----------------------------
echo "Restarting service..."
sudo systemctl daemon-reload
sudo systemctl restart uninaquiz

# -----------------------------
# 7. Health check
# -----------------------------
echo "Waiting for startup..."
sleep 3

HEALTH_URL="http://localhost:${APP_PORT}/api/health"

for i in {1..10}; do
    if curl -sf "$HEALTH_URL" >/dev/null 2>&1; then
        echo "Health check passed!"
        exit 0
    fi
    echo "Attempt $i failed... retrying"
    sleep 2
done

# -----------------------------
# 8. Rollback
# -----------------------------
echo "HEALTH CHECK FAILED - ROLLING BACK"

if [ -f "${DEPLOY_PATH}/${BINARY_NAME}.old" ]; then
    mv "${DEPLOY_PATH}/${BINARY_NAME}.old" "${DEPLOY_PATH}/${BINARY_NAME}"
    chmod 755 "${DEPLOY_PATH}/${BINARY_NAME}"
    sudo systemctl restart uninaquiz
    echo "Rollback completed"
else
    echo "No backup found, cannot rollback"
fi

exit 1
