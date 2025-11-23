#!/bin/bash
# Simple test script for API
# Run: chmod +x tests/test.sh && ./tests/test.sh

BASE_URL="http://localhost:8080"
REPORTS_DIR="reports"

# Create reports directory if it doesn't exist
mkdir -p "$REPORTS_DIR"

echo "=== Test 1: Check links ==="
RESPONSE1=$(curl -s -X POST "$BASE_URL/submit" \
  -H "Content-Type: application/json" \
  -d '{"links": ["google.com", "yandex.ru", "malformedlink.gg"]}')

if [ $? -eq 0 ]; then
    echo "Success! Response:"
    echo "$RESPONSE1" | jq .
    LINKS_NUM1=$(echo "$RESPONSE1" | jq -r '.links_num')
    echo -e "\nLinks number: $LINKS_NUM1"
else
    echo "Error executing request"
    exit 1
fi

echo -e "\n=== Test 2: Check another set of links ==="
RESPONSE2=$(curl -s -X POST "$BASE_URL/submit" \
  -H "Content-Type: application/json" \
  -d '{"links": ["gg.c", "yandex.ru"]}')

if [ $? -eq 0 ]; then
    echo "Success! Response:"
    echo "$RESPONSE2" | jq .
    LINKS_NUM2=$(echo "$RESPONSE2" | jq -r '.links_num')
    echo -e "\nLinks number: $LINKS_NUM2"
else
    echo "Error executing request"
    exit 1
fi

echo -e "\n=== Test 3: Get PDF report ==="
PDF_PATH="$REPORTS_DIR/report_${LINKS_NUM1}_${LINKS_NUM2}.pdf"
curl -s -X POST "$BASE_URL/report" \
  -H "Content-Type: application/json" \
  -d "{\"links_list\": [$LINKS_NUM1, $LINKS_NUM2]}" \
  --output "$PDF_PATH"

if [ $? -eq 0 ] && [ -f "$PDF_PATH" ]; then
    echo "PDF report saved to $PDF_PATH"
    echo "File size: $(stat -f%z "$PDF_PATH" 2>/dev/null || stat -c%s "$PDF_PATH" 2>/dev/null) bytes"
else
    echo "Error getting PDF report"
    exit 1
fi

echo -e "\n=== All tests passed! ==="

