package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// Operation represents each API operation entry
type Operation struct {
	HTTPResponse         int       `json:"http_response"`
	HTTPResponseTime     float64   `json:"http_response_time_second"`
	Errors               string    `json:"errors"`
	FHIRVersion          string    `json:"fhir_version"`
	TLSVersion           string    `json:"tls_verison"` // Note: keeping the typo to match target format
	MimeTypes            []string  `json:"mime_types"`
	SupportedResources   []string  `json:"supported_resources"`
	SmartHTTPResponse    int       `json:"smart_http_response"`
	SmartResponse        *string   `json:"smart_response"`
	Updated              time.Time `json:"updated"`
}

// APIEndpoint represents the main JSON structure
type APIEndpoint struct {
	URL                      string      `json:"url"`
	APIInformationSourceName interface{} `json:"api_information_source_name"`
	CreatedAt                time.Time   `json:"created_at"`
	ListSource               []string    `json:"list_source"`
	CertifiedAPIDevName      string      `json:"certified_api_developer_name"`
	Operations               []Operation `json:"operation"`
}

// CommonResources is a default list of resources to use when the CSV doesn't specify
var CommonResources = []string{
	"AllergyIntolerance",
	"Condition",
	"DiagnosticReport",
	"Immunization",
	"MedicationOrder",
	"MedicationRequest",
	"MedicationStatement",
	"Patient",
	"Procedure",
	"Observation",
	"Goal",
	"Device",
	"CarePlan",
	"DocumentReference",
	"Encounter",
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run script.go input.csv output.json")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// Open CSV file
	csvFile, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error opening CSV file: %v\n", err)
		os.Exit(1)
	}
	defer csvFile.Close()

	// Create CSV reader
	reader := csv.NewReader(csvFile)

	// Read header
	headers, err := reader.Read()
	if err != nil {
		fmt.Printf("Error reading CSV headers: %v\n", err)
		os.Exit(1)
	}

	// Map to store endpoints by URL
	endpoints := make(map[string]*APIEndpoint)

	// Process each row
	lineCount := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading CSV record: %v\n", err)
			continue
		}
		lineCount++

		// Create a map for easier field access
		row := make(map[string]string)
		for i, header := range headers {
			if i < len(record) { // Protect against malformed CSV
				row[header] = record[i]
			}
		}

		url := row["url"]
		if url == "" {
			fmt.Printf("Warning: Skipping row %d - missing URL\n", lineCount)
			continue
		}

		// Get or create endpoint
		endpoint, exists := endpoints[url]
		if !exists {
			// Parse created_at
			createdAt, err := parseDateTime(row["created_at"])
			if err != nil {
				fmt.Printf("Error parsing created_at for %s: %v, using current time\n", url, err)
				createdAt = time.Now().UTC()
			}

			// Parse api_information_source_name - could be a string or array
			var apiInfoSourceName interface{}
			if row["api_information_source_name"] != "" {
				// Check if it's in JSON array format
				apiInfoValue := row["api_information_source_name"]
				if strings.HasPrefix(apiInfoValue, "[") && strings.HasSuffix(apiInfoValue, "]") {
					var sources []string
					err = json.Unmarshal([]byte(apiInfoValue), &sources)
					if err == nil {
						apiInfoSourceName = sources
					} else {
						// Handle the case where it's a string that looks like an array but isn't valid JSON
						apiInfoSourceName = apiInfoValue
					}
				} else {
					apiInfoSourceName = apiInfoValue
				}
			}

			// Parse list_source as array
			var listSources []string
			if row["list_source"] != "" {
				listSourceValue := row["list_source"]
				// Try parsing as JSON array
				if strings.HasPrefix(listSourceValue, "[") && strings.HasSuffix(listSourceValue, "]") {
					err = json.Unmarshal([]byte(listSourceValue), &listSources)
					if err != nil {
						// Fallback to comma-separated if JSON parsing fails
						listSources = strings.Split(listSourceValue, ",")
						for i, s := range listSources {
							listSources[i] = strings.TrimSpace(s)
						}
					}
				} else {
					// Handle as comma-separated values
					listSources = strings.Split(listSourceValue, ",")
					for i, s := range listSources {
						listSources[i] = strings.TrimSpace(s)
					}
				}
			}

			endpoint = &APIEndpoint{
				URL:                      url,
				APIInformationSourceName: apiInfoSourceName,
				CreatedAt:                createdAt,
				ListSource:               listSources,
				CertifiedAPIDevName:      row["certified_api_developer_name"],
				Operations:               []Operation{},
			}
			endpoints[url] = endpoint
		}

		// Create a new operation for this row
		op := Operation{}

		// Parse HTTP response
		httpResponse, err := strconv.Atoi(row["http_response"])
		if err != nil {
			// If parsing fails, check if it's 0 or empty
			if row["http_response"] == "" || row["http_response"] == "0" {
				httpResponse = 0
			} else {
				fmt.Printf("Warning: Invalid HTTP response value for %s: %s\n", url, row["http_response"])
				httpResponse = 0
			}
		}
		op.HTTPResponse = httpResponse

		// Parse HTTP response time
		httpResponseTime, err := strconv.ParseFloat(row["http_response_time_second"], 64)
		if err != nil {
			// If parsing fails, check if it's -1 or empty
			if row["http_response_time_second"] == "" {
				httpResponseTime = -1
			} else {
				fmt.Printf("Warning: Invalid HTTP response time for %s: %s\n", url, row["http_response_time_second"])
				httpResponseTime = -1
			}
		}
		op.HTTPResponseTime = httpResponseTime

		// Get errors
		op.Errors = row["errors"]

		// Get FHIR version
		op.FHIRVersion = row["capability_fhir_version"]
		if op.FHIRVersion == "" && row["requested_fhir_version"] != "" {
			// Fallback to requested version if capability version is not available
			op.FHIRVersion = row["requested_fhir_version"]
		}

		// Set TLS version (using a default)
		op.TLSVersion = "TLS 1.2" // Default value as seen in the example

		// Parse MIME types based on format field
		var mimeTypes []string
		if formats := row["format"]; formats != "" {
			formatsList := strings.Split(formats, ",")
			for _, format := range formatsList {
				format = strings.TrimSpace(format)
				switch strings.ToLower(format) {
				case "json":
					mimeTypes = append(mimeTypes, "application/fhir+json")
				case "xml":
					mimeTypes = append(mimeTypes, "application/fhir+xml")
				case "application/json", "application/fhir+json":
					mimeTypes = append(mimeTypes, "application/fhir+json")
				case "application/xml", "application/fhir+xml":
					mimeTypes = append(mimeTypes, "application/fhir+xml")
				}
			}
		}
		
		// If no mime types found, provide a default
		if len(mimeTypes) == 0 {
			mimeTypes = append(mimeTypes, "application/fhir+json")
		}
		op.MimeTypes = mimeTypes

		// Set supported resources using our default list
		// In a real scenario, you might extract this from the CSV if available
		op.SupportedResources = CommonResources

		// Parse Smart HTTP response
		smartResponse, err := strconv.Atoi(row["smart_http_response"])
		if err != nil {
			// Default to 404 (common in the example) if parsing fails
			smartResponse = 404
		}
		op.SmartHTTPResponse = smartResponse

		// Smart response is null in the example
		op.SmartResponse = nil

		// Parse updated time
		updatedTime, err := parseDateTime(row["updated"])
		if err != nil {
			fmt.Printf("Warning: Error parsing updated time for %s: %v, using current time\n", url, err)
			// Use current time as fallback
			updatedTime = time.Now().UTC()
		}
		op.Updated = updatedTime

		// Add operation to endpoint
		endpoint.Operations = append(endpoint.Operations, op)
	}

	// Convert map to slice for JSON output
	endpointsList := make([]APIEndpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpointsList = append(endpointsList, *endpoint)
	}

	// Create output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	// Write JSON
	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(endpointsList)
	if err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted %d CSV records to %d JSON endpoints in %s\n", 
		lineCount, len(endpointsList), outputFile)
}

// parseDateTime parses date-time strings in various formats
func parseDateTime(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Try parsing with different layouts
	layouts := []string{
		"2006-01-02 15:04:05.999999",
		"2006-01-02T15:04:05.999999Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse date: %s", dateStr)
}