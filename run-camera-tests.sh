#!/bin/bash

# Test script for running ONVIF camera integration tests
# Usage: ./run-camera-tests.sh [test-name]

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== ONVIF Camera Integration Tests ===${NC}"
echo

# Check if environment variables are set
if [ -z "$ONVIF_TEST_ENDPOINT" ] || [ -z "$ONVIF_TEST_USERNAME" ] || [ -z "$ONVIF_TEST_PASSWORD" ]; then
    echo -e "${YELLOW}Warning: Camera credentials not set${NC}"
    echo "Set the following environment variables:"
    echo "  export ONVIF_TEST_ENDPOINT=\"http://192.168.1.201/onvif/device_service\""
    echo "  export ONVIF_TEST_USERNAME=\"service\""
    echo "  export ONVIF_TEST_PASSWORD=\"Service.1234\""
    echo
    echo -e "${YELLOW}Tests will be skipped.${NC}"
    echo
fi

# Determine which tests to run
TEST_PATTERN="${1:-TestBoschFLEXIDOMEIndoor5100iIR}"

echo -e "${GREEN}Running tests matching: ${TEST_PATTERN}${NC}"
echo

# Run tests with verbose output
go test -v -run "$TEST_PATTERN" -timeout 60s

# Check exit code
if [ $? -eq 0 ]; then
    echo
    echo -e "${GREEN}✓ All tests passed!${NC}"
else
    echo
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
