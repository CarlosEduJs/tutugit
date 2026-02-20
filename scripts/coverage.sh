#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status
set -e

# Define colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== tutugit Test Coverage Generator ===${NC}\n"

OUTPUT_DIR="coverage_report"
mkdir -p "$OUTPUT_DIR"

echo -e "${YELLOW}Running tests and generating profile...${NC}"
# run tests with coverage profile
go test -coverprofile="$OUTPUT_DIR/coverage.out" ./...

echo -e "\n${YELLOW}Generating terminal summary...${NC}"
# print a summary to the terminal
go tool cover -func="$OUTPUT_DIR/coverage.out" | grep -v "100.0%" | grep -v "0.0%" || true
echo -e "\n${GREEN}Total Coverage:${NC}"
go tool cover -func="$OUTPUT_DIR/coverage.out" | grep total | awk '{print $3}'

echo -e "\n${YELLOW}Generating HTML report...${NC}"
# generate the HTML report
go tool cover -html="$OUTPUT_DIR/coverage.out" -o "$OUTPUT_DIR/index.html"

echo -e "\n${GREEN}Success!${NC}"
echo -e "A detailed HTML coverage report has been generated at: ${BLUE}$OUTPUT_DIR/index.html${NC}"
echo -e "You can open it in your browser to view line-by-line coverage."
