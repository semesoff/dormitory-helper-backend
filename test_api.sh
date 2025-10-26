#!/bin/bash

# API Test Examples
# Make sure both backend and gateway are running before executing these commands

BASE_URL="http://localhost:8081"

echo "=== Testing Dormitory Helper API ==="
echo ""

# 1. Get authentication token
echo "1. Getting authentication token..."
RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/auth/check" \
  -H "Content-Type: application/json" \
  -d '{"token": ""}')

echo "Response: $RESPONSE"
echo ""

# Extract token from response (requires jq)
if command -v jq &> /dev/null; then
    TOKEN=$(echo $RESPONSE | jq -r '.token')
    USER_ID=$(echo $RESPONSE | jq -r '.user_id')
    USERNAME=$(echo $RESPONSE | jq -r '.username')
    
    echo "Extracted:"
    echo "  User ID: $USER_ID"
    echo "  Username: $USERNAME"
    echo "  Token: ${TOKEN:0:50}..."
    echo ""
else
    echo "jq not found. Please install jq to automatically extract token"
    echo "Or manually copy the token from the response above"
    echo ""
    read -p "Enter token: " TOKEN
fi

# Wait for user to review
sleep 2

# 2. Create laundry booking
echo "2. Creating laundry booking..."
START_TIME=$(date -u -d "+1 hour" +"%Y-%m-%dT%H:00:00Z")
END_TIME=$(date -u -d "+2 hours" +"%Y-%m-%dT%H:00:00Z")

LAUNDRY_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/laundry/bookings" \
  -H "Content-Type: application/json" \
  -d "{
    \"token\": \"$TOKEN\",
    \"start_time\": \"$START_TIME\",
    \"end_time\": \"$END_TIME\"
  }")

echo "Response: $LAUNDRY_RESPONSE"
echo ""

if command -v jq &> /dev/null; then
    LAUNDRY_BOOKING_ID=$(echo $LAUNDRY_RESPONSE | jq -r '.booking_id')
    echo "Created laundry booking ID: $LAUNDRY_BOOKING_ID"
    echo ""
fi

sleep 2

# 3. Create kitchen booking
echo "3. Creating kitchen booking..."
KITCHEN_START_TIME=$(date -u -d "+3 hours" +"%Y-%m-%dT%H:00:00Z")
KITCHEN_END_TIME=$(date -u -d "+5 hours" +"%Y-%m-%dT%H:00:00Z")

KITCHEN_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/kitchen/bookings" \
  -H "Content-Type: application/json" \
  -d "{
    \"token\": \"$TOKEN\",
    \"start_time\": \"$KITCHEN_START_TIME\",
    \"end_time\": \"$KITCHEN_END_TIME\"
  }")

echo "Response: $KITCHEN_RESPONSE"
echo ""

if command -v jq &> /dev/null; then
    KITCHEN_BOOKING_ID=$(echo $KITCHEN_RESPONSE | jq -r '.booking_id')
    echo "Created kitchen booking ID: $KITCHEN_BOOKING_ID"
    echo ""
fi

sleep 2

# 4. Get all laundry bookings
echo "4. Getting all laundry bookings..."
ALL_LAUNDRY=$(curl -s -X GET "${BASE_URL}/api/v1/laundry/bookings")
echo "Response: $ALL_LAUNDRY" | jq '.' 2>/dev/null || echo "$ALL_LAUNDRY"
echo ""

sleep 2

# 5. Get my laundry bookings
echo "5. Getting my laundry bookings..."
MY_LAUNDRY=$(curl -s -X GET "${BASE_URL}/api/v1/laundry/bookings/my?token=$TOKEN")
echo "Response: $MY_LAUNDRY" | jq '.' 2>/dev/null || echo "$MY_LAUNDRY"
echo ""

sleep 2

# 6. Get all kitchen bookings
echo "6. Getting all kitchen bookings..."
ALL_KITCHEN=$(curl -s -X GET "${BASE_URL}/api/v1/kitchen/bookings")
echo "Response: $ALL_KITCHEN" | jq '.' 2>/dev/null || echo "$ALL_KITCHEN"
echo ""

sleep 2

# 7. Get my kitchen bookings
echo "7. Getting my kitchen bookings..."
MY_KITCHEN=$(curl -s -X GET "${BASE_URL}/api/v1/kitchen/bookings/my?token=$TOKEN")
echo "Response: $MY_KITCHEN" | jq '.' 2>/dev/null || echo "$MY_KITCHEN"
echo ""

sleep 2

# 8. Delete laundry booking (if ID was extracted)
if [ ! -z "$LAUNDRY_BOOKING_ID" ] && [ "$LAUNDRY_BOOKING_ID" != "null" ]; then
    echo "8. Deleting laundry booking $LAUNDRY_BOOKING_ID..."
    DELETE_RESPONSE=$(curl -s -X DELETE "${BASE_URL}/api/v1/laundry/bookings/${LAUNDRY_BOOKING_ID}?token=$TOKEN")
    echo "Response: $DELETE_RESPONSE"
    echo ""
else
    echo "8. Skipping laundry booking deletion (no ID available)"
    echo ""
fi

sleep 2

# 9. Delete kitchen booking (if ID was extracted)
if [ ! -z "$KITCHEN_BOOKING_ID" ] && [ "$KITCHEN_BOOKING_ID" != "null" ]; then
    echo "9. Deleting kitchen booking $KITCHEN_BOOKING_ID..."
    DELETE_KITCHEN=$(curl -s -X DELETE "${BASE_URL}/api/v1/kitchen/bookings/${KITCHEN_BOOKING_ID}?token=$TOKEN")
    echo "Response: $DELETE_KITCHEN"
    echo ""
else
    echo "9. Skipping kitchen booking deletion (no ID available)"
    echo ""
fi

echo "=== Testing Complete ==="
echo ""
echo "Summary:"
echo "  Token: ${TOKEN:0:50}..."
echo "  Laundry Booking ID: $LAUNDRY_BOOKING_ID"
echo "  Kitchen Booking ID: $KITCHEN_BOOKING_ID"
