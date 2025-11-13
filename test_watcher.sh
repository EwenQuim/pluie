#!/bin/bash

# Test script to verify file watcher is working
# Run this in one terminal, then modify a file in testdata/public_folder/ in another terminal

echo "Building pluie..."
go build || exit 1

echo ""
echo "Starting server with file watching enabled..."
echo "Path: ./testdata/public_folder"
echo ""
echo "Server logs:"
echo "============================================"

./pluie -path=./testdata/public_folder -watch=true
