#!/bin/bash

BIN_DIR="bin"
SERVICES=("api-gateway" "user-service" "video-service" "streaming-service")

# Create the bin directory if it doesn't exist
mkdir -p ./$BIN_DIR

# Build each service
for SERVICE in "${SERVICES[@]}"; do
  echo "Building $SERVICE..."
  cd $SERVICE
  CGO=0 GOOS=linux GOARCH=amd64 go build -o ../$BIN_DIR/$SERVICE
  if [ $? -ne 0 ]; then
    echo "Failed to build $SERVICE"
    exit 1
  fi
  cd ..
done

echo "All services built successfully!"