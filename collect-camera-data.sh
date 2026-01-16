#!/bin/bash
# collect-camera-data.sh - Collect test data from all discovered cameras

set -e

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}ONVIF Camera Data Collection${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check if diagnostics tool exists
if [ ! -f "./bin/onvif-diagnostics" ]; then
    echo -e "${RED}Error: onvif-diagnostics not found. Building...${NC}"
    go build -o bin/onvif-diagnostics ./cmd/onvif-diagnostics
    echo -e "${GREEN}✓ Built onvif-diagnostics${NC}"
fi

# Prompt for credentials
echo -e "${YELLOW}Enter ONVIF credentials for your cameras:${NC}"
read -p "Username: " ONVIF_USER
read -sp "Password: " ONVIF_PASS
echo ""
echo ""

# Cameras discovered
declare -a CAMERAS=(
    "192.168.2.61:8000|Reolink_E1Zoom"
    "192.168.2.57:80|Bosch_AUTODOME_5000i"
    "192.168.2.82:80|AXIS_P3818"
    "192.168.2.236:8000|Reolink_TrackMixWiFi"
    "192.168.2.200:80|Bosch_FLEXIDOME_8000i"
    "192.168.2.24:80|Bosch_FLEXIDOME_5100i"
    "192.168.2.190:80|AXIS_Q3819"
    "192.168.2.30:80|AXIS_P5655"
)

SUCCESS=0
FAILED=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

# Create output directory for this batch
BATCH_DIR="camera-data-batch-${TIMESTAMP}"
mkdir -p "${BATCH_DIR}"

echo -e "${GREEN}Collecting data from ${#CAMERAS[@]} cameras...${NC}"
echo ""

# Loop through each camera
for camera_info in "${CAMERAS[@]}"; do
    IFS='|' read -r ip_port name <<< "$camera_info"
    
    # Check if port is specified
    if [[ $ip_port == *":"* ]]; then
        ENDPOINT="http://${ip_port}/onvif/device_service"
    else
        ENDPOINT="http://${ip_port}:80/onvif/device_service"
    fi
    
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}Camera: ${name}${NC}"
    echo -e "${YELLOW}Endpoint: ${ENDPOINT}${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    # Run COMPREHENSIVE diagnostics with XML capture (captures all operations)
    if ./bin/onvif-diagnostics \
        -endpoint "${ENDPOINT}" \
        -username "${ONVIF_USER}" \
        -password "${ONVIF_PASS}" \
        -capture-all \
        -verbose 2>&1 | tee "${BATCH_DIR}/${name}_log.txt"; then
        
        echo -e "${GREEN}✓ Successfully captured data from ${name}${NC}"
        SUCCESS=$((SUCCESS + 1))
    else
        echo -e "${RED}✗ Failed to capture data from ${name}${NC}"
        FAILED=$((FAILED + 1))
    fi
    
    echo ""
    sleep 2  # Brief delay between cameras to avoid network congestion
done

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Collection Complete${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "Success: ${GREEN}${SUCCESS}${NC} / ${#CAMERAS[@]}"
echo -e "Failed:  ${RED}${FAILED}${NC} / ${#CAMERAS[@]}"
echo ""
echo -e "${YELLOW}Results saved to: ${BATCH_DIR}/${NC}"
echo ""

# Move camera-logs to batch directory
if [ -d "camera-logs" ]; then
    echo -e "${YELLOW}Moving camera-logs to batch directory...${NC}"
    mv camera-logs/* "${BATCH_DIR}/" 2>/dev/null || true
    echo -e "${GREEN}✓ Logs organized${NC}"
fi

echo ""
echo -e "${GREEN}Next steps:${NC}"
echo "1. Review the capture files in ${BATCH_DIR}/"
echo "2. Copy .tar.gz files to testdata/captures/"
echo "3. Run: go build -o bin/generate-tests ./cmd/generate-tests"
echo "4. Generate tests for each camera capture"
echo ""
