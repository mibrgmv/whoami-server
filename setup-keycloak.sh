#!/bin/bash

ADMIN_TOKEN=$(curl -s -X POST \
  http://localhost:8088/realms/master/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin" \
  -d "password=admin" \
  -d "grant_type=password" \
  -d "client_id=admin-cli" | jq -r '.access_token')

echo "Creating realm 'myrealm'..."
curl -s -X POST \
  http://localhost:8088/admin/realms \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "realm": "myrealm",
    "enabled": true,
    "displayName": "My Realm",
    "registrationAllowed": true,
    "loginWithEmailAllowed": true,
    "duplicateEmailsAllowed": false
  }'

echo "Creating public client 'whoami-public'..."
curl -s -X POST \
  http://localhost:8088/admin/realms/myrealm/clients \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "whoami-public",
    "enabled": true,
    "publicClient": true,
    "directAccessGrantsEnabled": true,
    "standardFlowEnabled": true,
    "implicitFlowEnabled": false,
    "serviceAccountsEnabled": false,
    "redirectUris": ["*"],
    "webOrigins": ["*"],
    "protocol": "openid-connect"
  }'

echo "Creating private client 'whoami-admin'..."
curl -s -X POST \
  http://localhost:8088/admin/realms/myrealm/clients \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "whoami-admin",
    "enabled": true,
    "publicClient": false,
    "serviceAccountsEnabled": true,
    "standardFlowEnabled": false,
    "directAccessGrantsEnabled": false,
    "protocol": "openid-connect"
  }'

ADMIN_CLIENT_UUID=$(curl -s -X GET \
  "http://localhost:8088/admin/realms/myrealm/clients?clientId=whoami-admin" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id')

ADMIN_CLIENT_SECRET=$(curl -s -X GET \
  "http://localhost:8088/admin/realms/myrealm/clients/$ADMIN_CLIENT_UUID/client-secret" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.value')

echo "Setup complete!"
echo "Admin client secret: $ADMIN_CLIENT_SECRET"
echo "Update your KEYCLOAK_ADMIN_CLIENT_SECRET with this value"