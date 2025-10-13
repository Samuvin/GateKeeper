#!/bin/bash

# Start script for the database service
# Runs the database HTTP service on port 8081

set -e

cd "$(dirname "$0")/.."

echo "🚀 Starting database service..."

cd database

# Download dependencies
echo "📦 Installing dependencies..."
go mod download
go mod tidy

# Run the server
echo "✨ Starting server on port 8081..."
go run cmd/server/main.go


