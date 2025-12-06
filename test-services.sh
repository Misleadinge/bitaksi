#!/bin/bash

echo "========================================="
echo "TaxiHub Service Health Check"
echo "========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check Docker containers
echo "1. Checking Docker containers..."
docker-compose ps
echo ""

# Check Gateway Health
echo "2. Checking Gateway Health..."
GATEWAY_HEALTH=$(curl -s http://localhost:8080/health)
if [ "$GATEWAY_HEALTH" == '{"status":"ok"}' ]; then
    echo -e "${GREEN}✓ Gateway is healthy${NC}"
else
    echo -e "${RED}✗ Gateway health check failed${NC}"
fi
echo ""

# Check Driver Service Health
echo "3. Checking Driver Service Health..."
DRIVER_HEALTH=$(curl -s http://localhost:8081/health)
if [ "$DRIVER_HEALTH" == '{"status":"ok"}' ]; then
    echo -e "${GREEN}✓ Driver Service is healthy${NC}"
else
    echo -e "${RED}✗ Driver Service health check failed${NC}"
fi
echo ""

# Test Login
echo "4. Testing Authentication..."
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"test"}' | jq -r '.token')

if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
    echo -e "${GREEN}✓ Login successful, token received${NC}"
    echo "Token: ${TOKEN:0:50}..."
else
    echo -e "${RED}✗ Login failed${NC}"
    TOKEN=""
fi
echo ""

# Test List Drivers (Public endpoint)
echo "5. Testing List Drivers (Public)..."
LIST_RESPONSE=$(curl -s "http://localhost:8080/drivers?page=1&pageSize=5")
if echo "$LIST_RESPONSE" | jq -e '.drivers' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ List drivers endpoint working${NC}"
    echo "Response: $LIST_RESPONSE"
else
    echo -e "${RED}✗ List drivers failed${NC}"
    echo "Response: $LIST_RESPONSE"
fi
echo ""

# Test Create Driver (Protected)
if [ -n "$TOKEN" ]; then
    echo "6. Testing Create Driver (Protected)..."
    CREATE_RESPONSE=$(curl -s -X POST http://localhost:8080/drivers \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d '{
        "firstName": "Ahmet",
        "lastName": "Demir",
        "plate": "34ABC123",
        "taksiType": "sari",
        "carBrand": "Toyota",
        "carModel": "Corolla",
        "lat": 41.0431,
        "lon": 29.0099
      }')
    
    if echo "$CREATE_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
        DRIVER_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id')
        echo -e "${GREEN}✓ Create driver successful${NC}"
        echo "Driver ID: $DRIVER_ID"
    else
        echo -e "${RED}✗ Create driver failed${NC}"
        echo "Response: $CREATE_RESPONSE"
    fi
    echo ""
    
    # Test Find Nearby
    if [ -n "$DRIVER_ID" ]; then
        echo "7. Testing Find Nearby Drivers..."
        NEARBY_RESPONSE=$(curl -s "http://localhost:8080/drivers/nearby?lat=41.0082&lon=28.9784&taksiType=sari")
        if echo "$NEARBY_RESPONSE" | jq -e '.[]' > /dev/null 2>&1 || echo "$NEARBY_RESPONSE" | jq -e '. == []' > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Find nearby drivers endpoint working${NC}"
            echo "Response: $NEARBY_RESPONSE"
        else
            echo -e "${RED}✗ Find nearby drivers failed${NC}"
            echo "Response: $NEARBY_RESPONSE"
        fi
        echo ""
    fi
else
    echo -e "${YELLOW}⚠ Skipping protected endpoints (no token)${NC}"
    echo ""
fi

# Check Swagger endpoints
echo "8. Checking Swagger Documentation..."
GATEWAY_SWAGGER=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/swagger/index.html)
DRIVER_SWAGGER=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/swagger/index.html)

if [ "$GATEWAY_SWAGGER" == "200" ]; then
    echo -e "${GREEN}✓ Gateway Swagger available at http://localhost:8080/swagger/index.html${NC}"
else
    echo -e "${YELLOW}⚠ Gateway Swagger not available${NC}"
fi

if [ "$DRIVER_SWAGGER" == "200" ]; then
    echo -e "${GREEN}✓ Driver Service Swagger available at http://localhost:8081/swagger/index.html${NC}"
else
    echo -e "${YELLOW}⚠ Driver Service Swagger not available${NC}"
fi
echo ""

echo "========================================="
echo "Health Check Complete!"
echo "========================================="

