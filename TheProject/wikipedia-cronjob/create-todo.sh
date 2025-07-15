#!/usr/bin/env bash
set -e

echo "Starting Wikipedia todo generator at $(date)"

# Fetch random Wikipedia URL by following the redirect
echo "Fetching random Wikipedia article..."
WIKIPEDIA_URL=$(curl -sI "https://en.wikipedia.org/wiki/Special:Random" | grep -i "^location:" | cut -d' ' -f2 | tr -d '\r')

if [ -z "$WIKIPEDIA_URL" ]; then
    echo "Error: Could not fetch random Wikipedia URL"
    exit 1
fi

echo "Got Wikipedia URL: $WIKIPEDIA_URL"

# Create todo text
TODO_TEXT="Read $WIKIPEDIA_URL"

echo "Creating todo: $TODO_TEXT"

# Create JSON payload
JSON_PAYLOAD=$(cat <<EOF
{
    "text": "$TODO_TEXT",
    "priority": "medium"
}
EOF
)

# Send POST request to todo-backend
RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$JSON_PAYLOAD" \
    "http://todo-backend-service.project.svc.cluster.local:3001/todos")

if [ $? -eq 0 ]; then
    echo "Successfully created todo!"
    echo "Response: $RESPONSE"
else
    echo "Error: Failed to create todo"
    exit 1
fi

echo "Wikipedia todo generator completed successfully at $(date)" 