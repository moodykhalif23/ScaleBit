#!/bin/bash

echo "Testing ScaleBit Authentication Flow"
echo "===================================="

# Test 1: Check if KrakenD is running
echo "1. Testing KrakenD connectivity..."
curl -s -o /dev/null -w "%{http_code}" http://localhost:8000/login > /tmp/krakend_test
if [ $? -eq 0 ]; then
    echo "✓ KrakenD is accessible"
else
    echo "✗ KrakenD is not accessible"
    exit 1
fi

# Test 2: Test login endpoint
echo "2. Testing login endpoint..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8000/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}')

if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    echo "✓ Login endpoint is working"
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo "Token received: ${TOKEN:0:20}..."
else
    echo "✗ Login failed"
    echo "Response: $LOGIN_RESPONSE"
fi

# Test 3: Test authenticated endpoint
if [ ! -z "$TOKEN" ]; then
    echo "3. Testing authenticated endpoint..."
    AUTH_RESPONSE=$(curl -s -w "%{http_code}" -X GET http://localhost:8000/users \
      -H "Authorization: Bearer $TOKEN")
    
    HTTP_CODE=$(echo "$AUTH_RESPONSE" | tail -c 4)
    if [ "$HTTP_CODE" = "200" ]; then
        echo "✓ Authenticated request successful"
    else
        echo "✗ Authenticated request failed with code: $HTTP_CODE"
        echo "Response: $AUTH_RESPONSE"
    fi
else
    echo "3. Skipping authenticated endpoint test (no token)"
fi

echo "===================================="
echo "Test completed"
