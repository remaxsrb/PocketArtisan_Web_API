#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-remax}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Qcom_RK_t3st}"

echo "Running bulk test flow..."
echo

echo "0) Clear Go test cache"
go clean -testcache
echo

echo "1) User bulk creation test"
go test -v ./internal/modules/users/common/register -run '^TestBulkUserRegistration$'
echo

echo "2) Login as admin and fetch access token"
LOGIN_PAYLOAD=$(printf '{"username":"%s","password":"%s"}' "$ADMIN_USERNAME" "$ADMIN_PASSWORD")
LOGIN_RESPONSE=$(curl -sS -X POST "$API_BASE_URL/users/login" -H "Content-Type: application/json" -d "$LOGIN_PAYLOAD")

ADMIN_BEARER_TOKEN=$(printf '%s' "$LOGIN_RESPONSE" | tr -d '\n' | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')

if [[ -z "$ADMIN_BEARER_TOKEN" ]]; then
	echo "Failed to get admin access token from /users/login"
	echo "Login response: $LOGIN_RESPONSE"
	exit 1
fi
echo

echo "3) Craft bulk insertion test"
ADMIN_BEARER_TOKEN="$ADMIN_BEARER_TOKEN" go test -v ./internal/modules/crafts/create -run '^TestBulkCraftCreate$'
echo

echo "4) Product category bulk insertion test"
ADMIN_BEARER_TOKEN="$ADMIN_BEARER_TOKEN" go test -v ./internal/modules/product_categories/create -run '^TestBulkProductCategoryCreate$'
echo

echo "5. Craftsman Application bulk insertion"
go test -v ./internal/modules/craftsman_application/create -run '^TestBulkCreateCraftsmanApplication$'
echo

echo "6) Craftsman application bulk approve test"
ADMIN_BEARER_TOKEN="$ADMIN_BEARER_TOKEN" go test -v ./internal/modules/craftsman_application/approve -run '^TestBulkApproveCraftsmanApplication$'
echo

echo "7) Bulk elevate user to craftsman"
ADMIN_BEARER_TOKEN="$ADMIN_BEARER_TOKEN" go test -v ./internal/modules/users/craftsman/create -run '^TestBulkApproveCraftsmanApplication$'
echo

echo "All requested tests finished."
