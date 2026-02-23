#!/bin/sh

# Payer Validation Script
# This script runs the payer validator service to validate payer-submitted FHIR endpoints
# and integrate valid endpoints into the main Lantern data pipeline.
#
# Usage:
#   ./payer_validation.sh           - Run validation and update database
#   ./payer_validation.sh --dry-run - Run validation without updating database
#
# What the validator does:
#   1. Fetches all unvalidated payer endpoint registrations from payer_endpoints table
#   2. For each endpoint:
#      - Requests /metadata endpoint
#      - Validates HTTP 200 response
#      - Validates response is valid JSON or XML
#      - Validates resourceType is CapabilityStatement (FHIR R4+) or Conformance (DSTU2)
#   3. Updates payer_endpoints with validation_result and validation_comments
#   4. For valid endpoints (when not in dry-run mode):
#      - Inserts endpoint into fhir_endpoints table
#      - Creates/updates list_source_info entry for "Payer Self-Registration"
#      - Inserts payer-specific metadata into payer_info table
#      - Marks endpoint as is_persisted = true

log_file="/etc/lantern/logs/payer_validation_logs.txt"
current_datetime=$(date +"%Y-%m-%d %H:%M:%S")

# Check for dry-run flag
DRY_RUN=""
if [ "$1" = "--dry-run" ] || [ "$1" = "-d" ]; then
    DRY_RUN="--dry-run"
    echo "$current_datetime - Running payer validation in DRY-RUN mode (no database updates)" >> $log_file
fi

# Navigate to project root and load environment variables
cd "$(dirname "$0")/.."
export $(cat .env 2>/dev/null | grep -v '^#' | xargs)

# Navigate to payer validator cmd directory
cd endpointmanager/cmd/payervalidator

echo "$current_datetime - Starting payer endpoint validation..." >> $log_file

# Run the payer validator
if [ -n "$DRY_RUN" ]; then
    go run main.go --dry-run 2>&1 | tee -a $log_file
    exit_code=$?
else
    go run main.go 2>&1 | tee -a $log_file
    exit_code=$?
fi

if [ $exit_code -eq 0 ]; then
    echo "$current_datetime - Payer endpoint validation completed successfully" >> $log_file
else
    echo "$current_datetime - Error: Payer endpoint validation failed with exit code $exit_code" >> $log_file
fi

exit $exit_code
