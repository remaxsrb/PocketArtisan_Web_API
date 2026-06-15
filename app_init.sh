#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-remax}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Qcom_RK_t3st}"
ASSETS_DIR="${ASSETS_DIR:-$SCRIPT_DIR/assets}"
PRODUCT_DATA_FILE="${PRODUCT_DATA_FILE:-./internal/modules/product/create/new_product_test_data.json}"
CRAFTSMAN_PASSWORD="${CRAFTSMAN_PASSWORD:-Secure_pass1}"

echo "Running bulk test flow..."
echo

echo "0) Clear Go test cache"
go clean -testcache
echo

echo "Cleaning uploads directory"
find ./uploads -type f -delete 2>/dev/null || true

echo "1) User bulk creation test"
go test -v ./internal/modules/users/common/register -run '^TestBulkUserRegistration$'
echo

echo "2) Login as admin and fetch access token"
LOGIN_PAYLOAD=$(printf '{"username":"%s","password":"%s"}' "$ADMIN_USERNAME" "$ADMIN_PASSWORD")
LOGIN_RESPONSE=$(curl -sS -X POST "$API_BASE_URL/users/login" \
  -H "Content-Type: application/json" \
  -d "$LOGIN_PAYLOAD")
ADMIN_BEARER_TOKEN=$(printf '%s' "$LOGIN_RESPONSE" | tr -d '\n' | \
  sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')
if [[ -z "$ADMIN_BEARER_TOKEN" ]]; then
  echo "Failed to get admin access token from /users/login"
  echo "Login response: $LOGIN_RESPONSE"
  exit 1
fi
echo

echo "3) Craft bulk insertion test"
ADMIN_BEARER_TOKEN="$ADMIN_BEARER_TOKEN" \
  go test -v ./internal/modules/crafts/create -run '^TestBulkCraftCreate$'
echo

echo "4) Product category bulk insertion test"
ADMIN_BEARER_TOKEN="$ADMIN_BEARER_TOKEN" \
  go test -v ./internal/modules/product_categories/create -run '^TestBulkProductCategoryCreate$'
echo

echo "5) Craftsman application bulk insertion"
go test -v ./internal/modules/craftsman_application/create -run '^TestBulkCreateCraftsmanApplication$'
echo

echo "6) Craftsman application bulk approve test"
ADMIN_BEARER_TOKEN="$ADMIN_BEARER_TOKEN" \
  go test -v ./internal/modules/craftsman_application/approve -run '^TestBulkApproveCraftsmanApplication$'
echo

echo "7) Bulk elevate user to craftsman"
ADMIN_BEARER_TOKEN="$ADMIN_BEARER_TOKEN" \
  go test -v ./internal/modules/users/craftsman/create -run '^TestBulkApproveCraftsmanApplication$'
echo

#echo "8) Bulk file upload test"
#ASSETS_DIR="$ASSETS_DIR" \
#  go test -v ./internal/modules/product/create -run '^TestBulkFileUpload$'
#echo

echo "8) Bulk product creation — nested loop: per craftsman → per product"
USERNAMES=$(jq -r '.[].username' "$PRODUCT_DATA_FILE" | awk '!seen[$0]++')

while IFS= read -r USERNAME; do
  echo "  → Logging in as $USERNAME"
  CRAFTSMAN_LOGIN_PAYLOAD=$(printf '{"username":"%s","password":"%s"}' "$USERNAME" "$CRAFTSMAN_PASSWORD")
  CRAFTSMAN_LOGIN_RESPONSE=$(curl -sS -X POST "$API_BASE_URL/users/login" \
    -H "Content-Type: application/json" \
    -d "$CRAFTSMAN_LOGIN_PAYLOAD")
  BEARER_TOKEN=$(printf '%s' "$CRAFTSMAN_LOGIN_RESPONSE" | tr -d '\n' | \
    sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')

  if [[ -z "$BEARER_TOKEN" ]]; then
    echo "  ✗ Failed to get token for $USERNAME — skipping"
    echo "    Response: $CRAFTSMAN_LOGIN_RESPONSE"
    continue
  fi

  # Extract all products for this craftsman and iterate over them.
  PRODUCTS=$(jq -c --arg u "$USERNAME" '.[] | select(.username == $u)' "$PRODUCT_DATA_FILE")

  while IFS= read -r PRODUCT; do
    NAME=$(printf '%s' "$PRODUCT" | jq -r '.name')
    ASSETS_DIR="$ASSETS_DIR" \
    BEARER_TOKEN="$BEARER_TOKEN" \
    CRAFTSMAN_USERNAME="$USERNAME" \
    PRODUCT_JSON="$PRODUCT" \
      go test -v ./internal/modules/product/create -run '^TestSingleProductCreate$'
    echo "    ✓ $NAME"
  done <<< "$PRODUCTS"

  echo
done <<< "$USERNAMES"

echo "All requested tests finished."