package chplendpointquerier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

type Entry struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	LogoURL string `json:"logoUrl,omitempty"`
	URL     string `json:"url"`
}

type Bundle struct {
	ResourceType string  `json:"resourceType"`
	Type         string  `json:"type"`
	Total        int     `json:"total"`
	Entry        []Entry `json:"entry"`
	ID           string  `json:"id"`
}

func BundleQuerierParser(CHPLURL string, fileToWriteTo string, errorFilePath string, listSource string) error {

	endpointEntryList := EndpointList{
		Endpoints: []LanternEntry{},
	}

	respBody, err := helpers.QueryEndpointList(CHPLURL)
	if err != nil {
		errMsg := classifyNetworkError(err)
		log.Errorf(errMsg)
		AppendQueryError(errorFilePath, listSource, errMsg)
	} else {
		// Check for invalid response format before attempting to parse
		trimmed := strings.TrimSpace(string(respBody))
		if len(trimmed) == 0 {
			parseErr := fmt.Errorf("FHIR bundle parsing failed: empty response body")
			log.Warn(parseErr)
			AppendQueryError(errorFilePath, listSource, parseErr.Error())
		} else if strings.HasPrefix(trimmed, "<") {
			parseErr := fmt.Errorf("FHIR bundle parsing failed: non-JSON (HTML/XML) response")
			log.Warn(parseErr)
			AppendQueryError(errorFilePath, listSource, parseErr.Error())
		} else if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
			parseErr := fmt.Errorf("FHIR bundle parsing failed: unexpected response format, first bytes: %.30s", trimmed)
			log.Warn(parseErr)
			AppendQueryError(errorFilePath, listSource, parseErr.Error())
		} else if !json.Valid(respBody) {
			parseErr := fmt.Errorf("FHIR bundle parsing failed: malformed JSON")
			log.Warn(parseErr)
			AppendQueryError(errorFilePath, listSource, parseErr.Error())
		} else {
			var resourceCheck struct {
				ResourceType string `json:"resourceType"`
			}
			if json.Unmarshal(respBody, &resourceCheck) == nil &&
				resourceCheck.ResourceType != "" &&
				resourceCheck.ResourceType != "Bundle" {
				parseErr := fmt.Errorf("FHIR bundle parsing failed: JSON resource is not of type Bundle (got %s)",
					resourceCheck.ResourceType)
				log.Warn(parseErr)
				AppendQueryError(errorFilePath, listSource, parseErr.Error())
			} else {
				// Convert bundle data to lantern format
				endpointEntryList.Endpoints = BundleToLanternFormat(respBody, CHPLURL)

				if len(endpointEntryList.Endpoints) == 0 {
					var emptyErr error
					if bytes.Contains(respBody, []byte(`"active":false`)) || bytes.Contains(respBody, []byte(`"active": false`)) {
						emptyErr = fmt.Errorf("FHIR bundle has entries but all organizations have active=false")
					} else {
						emptyErr = fmt.Errorf("FHIR bundle has 0 entries")
					}
					log.Warn(emptyErr)
					AppendQueryError(errorFilePath, listSource, emptyErr.Error())
				}
			}
		}
	}

	return WriteCHPLFile(endpointEntryList, fileToWriteTo)
}

func classifyNetworkError(err error) string {
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "certificate has expired") || strings.Contains(errStr, "certificate is not yet valid"):
		return fmt.Sprintf("SSL ERROR: certificate has expired or is not yet valid Error=%v", err)
	case strings.Contains(errStr, "x509:") || strings.Contains(errStr, "tls:"):
		return fmt.Sprintf("SSL ERROR: certificate validation failed Error=%v", err)
	case os.IsTimeout(err) || strings.Contains(errStr, "timeout") || strings.Contains(errStr, "context deadline exceeded"):
		return fmt.Sprintf("UNREACHABLE: request timed out Error=%v", err)
	case strings.Contains(errStr, "no such host") || strings.Contains(errStr, "dns"):
		return fmt.Sprintf("UNREACHABLE: DNS resolution failed Error=%v", err)
	case strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "connection reset") || strings.Contains(errStr, "connection timed out"):
		return fmt.Sprintf("UNREACHABLE: connection failed Error=%v", err)
	default:
		return fmt.Sprintf("NETWORK ERROR: failed to fetch Error=%v", err)
	}
}
