#!/bin/bash

echo "JWT Debug Script"
echo "================"

# Wait for services to be ready
echo "Waiting for services to start..."
sleep 15

# Check if services are running
echo "1. Checking service status..."
docker ps --format "table {{.Names}}\t{{.Status}}" | grep scalebit

echo ""
echo "2. Checking user-service JWT secret:"
docker logs scalebit-user-service-1 --tail 20 2>/dev/null | grep "JWT secret" || echo "No JWT secret logs found"

echo ""
echo "3. Testing login..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8000/login \
  -H "Content-Type: application/json" \
  -d '{"email":"bria@sozuri.net","password":"password123"}')

echo "Login response: $LOGIN_RESPONSE"

if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo ""
    echo "4. JWT Token (first 100 chars): ${TOKEN:0:100}..."
    
    # Decode header
    echo ""
    echo "5. Decoding JWT Header:"
    HEADER=$(echo "$TOKEN" | cut -d'.' -f1)
    # Convert base64url to base64
    HEADER_B64=$(echo "$HEADER" | sed 's/-/+/g; s/_/\//g')
    # Add padding
    case $((${#HEADER_B64} % 4)) in
        2) HEADER_B64="${HEADER_B64}==" ;;
        3) HEADER_B64="${HEADER_B64}=" ;;
    esac
    echo "$HEADER_B64" | base64 -d 2>/dev/null | python3 -m json.tool 2>/dev/null || echo "Could not decode header"
    
    # Decode payload
    echo ""
    echo "6. Decoding JWT Payload:"
    PAYLOAD=$(echo "$TOKEN" | cut -d'.' -f2)
    # Convert base64url to base64
    PAYLOAD_B64=$(echo "$PAYLOAD" | sed 's/-/+/g; s/_/\//g')
    # Add padding
    case $((${#PAYLOAD_B64} % 4)) in
        2) PAYLOAD_B64="${PAYLOAD_B64}==" ;;
        3) PAYLOAD_B64="${PAYLOAD_B64}=" ;;
    esac
    echo "$PAYLOAD_B64" | base64 -d 2>/dev/null | python3 -m json.tool 2>/dev/null || echo "Could not decode payload"
    
    echo ""
    echo "7. Testing authenticated request..."
    AUTH_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X GET http://localhost:8000/users \
      -H "Authorization: Bearer $TOKEN")
    
    HTTP_CODE=$(echo "$AUTH_RESPONSE" | grep "HTTP_CODE:" | cut -d':' -f2)
    RESPONSE_BODY=$(echo "$AUTH_RESPONSE" | sed '/HTTP_CODE:/d')
    
    echo "HTTP Status: $HTTP_CODE"
    echo "Response: $RESPONSE_BODY"
    
    if [ "$HTTP_CODE" = "200" ]; then
        echo "✓ SUCCESS!"
    else
        echo "✗ FAILED - Checking KrakenD logs for JWT validation errors..."
        docker logs scalebit-krakend-1 --tail 20 2>/dev/null | grep -i "jwt\|auth\|401" || echo "No JWT-related logs found"
    fi
else
    echo "✗ Login failed"
    echo "Checking user-service logs..."
    docker logs scalebit-user-service-1 --tail 10 2>/dev/null || echo "Could not get user-service logs"
fi

echo ""
echo "8. Checking JWK file content:"
docker exec scalebit-krakend-1 cat /etc/krakend/symmetric.jwk 2>/dev/null || echo "JWK file not accessible in container"

echo ""
echo "Debug completed."
