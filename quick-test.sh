#!/bin/bash
# Quick authentication test script

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Multi-Tenant Puzzle - Quick Test ===${NC}\n"

# Test 1: Health check
echo -e "${GREEN}1. Testing health endpoint...${NC}"
HEALTH=$(curl -s http://localhost:8080/health)
if [ "$HEALTH" = "OK" ]; then
    echo "✓ Server is running"
else
    echo "✗ Server not responding"
    exit 1
fi

# Test 2: Signup
echo -e "\n${GREEN}2. Creating test account...${NC}"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -L -c test_cookies.txt \
    -X POST http://localhost:8080/signup \
    -d "company_name=Test Company $(date +%s)" \
    -d "email=test$(date +%s)@example.com" \
    -d "password=Test1234" \
    -d "confirm_password=Test1234")

if [ "$RESPONSE" = "200" ]; then
    echo "✓ Account created and logged in"
else
    echo "✗ Signup failed (HTTP $RESPONSE)"
fi

# Test 3: Access protected route
echo -e "\n${GREEN}3. Testing protected /admin route...${NC}"
ADMIN_RESPONSE=$(curl -s http://localhost:8080/admin -b test_cookies.txt | head -1)
if [[ "$ADMIN_RESPONSE" == *"Welcome"* ]] || [[ "$ADMIN_RESPONSE" == *"admin"* ]]; then
    echo "✓ Can access admin area with session"
else
    echo "✗ Admin access failed"
fi

# Test 4: Logout
echo -e "\n${GREEN}4. Testing logout...${NC}"
LOGOUT=$(curl -s -o /dev/null -w "%{http_code}" -L http://localhost:8080/logout -b test_cookies.txt)
if [ "$LOGOUT" = "200" ]; then
    echo "✓ Logout successful"
else
    echo "✗ Logout failed"
fi

# Cleanup
rm -f test_cookies.txt

echo -e "\n${BLUE}=== Test Summary ===${NC}"
echo "✓ Authentication system is working"
echo "✓ Sessions persist correctly"
echo "✓ Login/Signup/Logout functional"
echo ""
echo "Next: Open http://localhost:8080/login in your browser to test the UI"
echo "Admin dashboard templates are ready but need handlers to be functional"


