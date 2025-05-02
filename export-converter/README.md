# FHIR Endpoint CSV to JSON Converter

This tool converts CSV data about FHIR endpoints into a structured JSON format that matches the specified target schema. The script processes CSV records representing FHIR endpoint operations and groups them by URL into consolidated endpoint objects.

## Features

- Converts CSV data about FHIR endpoints to a structured JSON format
- Groups multiple operations by endpoint URL
- Handles various date formats
- Parses MIME types from format strings
- Includes default supported resources for FHIR endpoints
- Provides error handling and validation

## Prerequisites

- Go 1.16 or later

## Usage

Place the CSV file in the same directory as the go script, then

1. Compile the Go script:

```bash
go build -o fhir-converter csv-to-json.go
```

2. Run the converter:

```bash
./fhir-converter <input-csv-file> <output-json-file>
```

For example:

```bash
./fhir-converter 05_01_2025endpointdata.csv endpoints.json
```

## Input CSV Format

The script expects a CSV file with the following columns:

- `url`: The endpoint URL (required)
- `api_information_source_name`: Name of the API information source (optional)
- `created_at`: Creation date of the endpoint (optional)
- `updated`: Last update date of the operation (optional)
- `list_source`: Source of the endpoint list (optional)
- `certified_api_developer_name`: Name of the certified API developer (optional)
- `capability_fhir_version`: FHIR version reported in capabilities (optional)
- `format`: Format(s) supported by the endpoint (e.g., "json,xml") (optional)
- `http_response`: HTTP response code (optional)
- `http_response_time_second`: Response time in seconds (optional)
- `smart_http_response`: SMART on FHIR HTTP response code (optional)
- `errors`: Any errors encountered (optional)

## Output JSON Format

The output is a JSON array of endpoint objects with the following structure:

```json
[
  {
    "url": "https://example.com/fhir/",
    "api_information_source_name": "Example Healthcare",
    "created_at": "2025-01-01T00:00:00Z",
    "list_source": ["https://example.com/list"],
    "certified_api_developer_name": "Example Developer",
    "operation": [
      {
        "http_response": 200,
        "http_response_time_second": 0.153,
        "errors": "",
        "fhir_version": "4.0.1",
        "tls_verison": "TLS 1.2",
        "mime_types": ["application/fhir+json"],
        "supported_resources": ["Patient", "Observation", ...],
        "smart_http_response": 404,
        "smart_response": null,
        "updated": "2025-05-01T12:34:56Z"
      },
      ...
    ]
  },
  ...
]
```

Note: 
 - The supported resoures key will be populated with new values only if the CSV file has the corresponding column.
 - In case the endpoint has multiple list-sources, one of them is selected at random to be displayed in the JSON export.

## Error Handling

The script includes error handling for:
- CSV parsing issues
- Missing required fields
- Date parsing errors
- JSON encoding issues

Errors are logged to the console, but the script attempts to continue processing when possible.

## Default Values

When certain values are missing from the CSV, the script uses these defaults:
- HTTP response: 0 if missing or invalid
- HTTP response time: -1 if missing or invalid
- SMART HTTP response: 404 if missing or invalid
- TLS version: "TLS 1.2"
- MIME types: ["application/fhir+json"] if none can be determined
- Dates: Current time if parsing fails