#!/bin/bash

echo "Testing JWT Fix for KrakenD Authentication"
echo "=========================================="

# Wait for services to be ready
echo "Waiting for services to start..."
sleep 10

# Test 1: Login and get token
echo "1. Testing login to get JWT token..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8000/login \
  -H "Content-Type: application/json" \
  -d '{"email":"bria@sozuri.net","password":"password123"}')

echo "Login response: $LOGIN_RESPONSE"

if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    echo "✓ Login successful - token received"
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo "Token (first 50 chars): ${TOKEN:0:50}..."
    
    # Decode and display token header and payload for debugging
    echo ""
    echo "2. Decoding JWT token header..."
    HEADER=$(echo "$TOKEN" | cut -d'.' -f1)
    # Add padding if needed for base64 decoding
    HEADER_PADDED=$(printf '%s' "$HEADER" | sed 's/-/+/g; s/_/\//g')
    case $((${#HEADER_PADDED} % 4)) in
        2) HEADER_PADDED="${HEADER_PADDED}==" ;;
        3) HEADER_PADDED="${HEADER_PADDED}=" ;;
    esac
    echo "Token header:"
    echo "$HEADER_PADDED" | base64 -d 2>/dev/null | jq . 2>/dev/null || echo "Could not decode header"
    
    echo ""
    echo "3. Testing authenticated endpoint..."
    AUTH_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X GET http://localhost:8000/users \
      -H "Authorization: Bearer $TOKEN")
    
    HTTP_CODE=$(echo "$AUTH_RESPONSE" | grep "HTTP_CODE:" | cut -d':' -f2)
    RESPONSE_BODY=$(echo "$AUTH_RESPONSE" | sed '/HTTP_CODE:/d')
    
    echo "HTTP Status Code: $HTTP_CODE"
    echo "Response: $RESPONSE_BODY"
    
    if [ "$HTTP_CODE" = "200" ]; then
        echo "✓ SUCCESS: Authenticated request successful!"
        echo "✓ JWT validation is now working correctly"
    else
        echo "✗ FAILED: Authenticated request failed with code: $HTTP_CODE"
        echo "Response body: $RESPONSE_BODY"
    fi
else
    echo "✗ Login failed"
    echo "Response: $LOGIN_RESPONSE"
fi

echo ""
echo "Test completed."
