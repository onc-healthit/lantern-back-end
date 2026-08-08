package chplendpointquerier

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

// maxEmptyBundleRetries is the number of extra attempts made when a query
// succeeds but returns a bundle with 0 entries, since some URLs
// intermittently return no data on the first hit.
const maxEmptyBundleRetries = 2

// emptyBundleRetryDelay is the pause between empty-bundle retry attempts.
const emptyBundleRetryDelay = 2 * time.Second

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

	for attempt := 0; attempt <= maxEmptyBundleRetries; attempt++ {
		if attempt > 0 {
			log.Warnf("Retrying BundleQuerierParser for %s due to empty bundle (attempt %d of %d)", CHPLURL, attempt, maxEmptyBundleRetries)
			time.Sleep(emptyBundleRetryDelay)
		}

		endpoints, emptyBundle := queryAndParseBundle(CHPLURL, errorFilePath, listSource)
		if !emptyBundle {
			endpointEntryList.Endpoints = endpoints
			break
		}

		if attempt == maxEmptyBundleRetries {
			emptyErr := fmt.Errorf("FHIR bundle has 0 entries")
			log.Warn(emptyErr)
			AppendQueryError(errorFilePath, listSource, emptyErr.Error())
		}
	}

	return WriteCHPLFile(endpointEntryList, fileToWriteTo)
}

// queryAndParseBundle performs a single query+parse attempt. It returns the
// parsed endpoints and, when the response was well-formed but contained no
// usable entries, emptyBundle=true so the caller can decide whether to retry.
// All other failures (network errors, malformed responses, inactive orgs)
// are terminal and are logged/appended here.
func queryAndParseBundle(CHPLURL string, errorFilePath string, listSource string) (endpoints []LanternEntry, emptyBundle bool) {
	respBody, err := helpers.QueryEndpointList(CHPLURL)
	if err != nil {
		errMsg := classifyNetworkError(err)
		log.Errorf(errMsg)
		AppendQueryError(errorFilePath, listSource, errMsg)
		return nil, false
	}

	if validationErr := validateBundleResponse(respBody); validationErr != nil {
		log.Warn(validationErr)
		AppendQueryError(errorFilePath, listSource, validationErr.Error())
		return nil, false
	}

	var unpopulatedOrgs []string
	var inactiveOrgCount int
	var noOrganizationsFound bool
	endpoints, unpopulatedOrgs, inactiveOrgCount, noOrganizationsFound = BundleToLanternFormat(respBody, CHPLURL)
	if noOrganizationsFound {
		msg := "No Organization resource(s) found in the FHIR bundle."
		log.Warn(msg)
		AppendQueryError(errorFilePath, listSource, msg)
	}

	if len(unpopulatedOrgs) > 0 {
		msg := fmt.Sprintf("%d organization(s) parsed from the FHIR bundle were not populated due to missing/invalid Endpoint resource reference.",
			len(unpopulatedOrgs))
		log.Warn(msg)
		AppendQueryError(errorFilePath, listSource, msg)
	}

	if inactiveOrgCount > 0 {
		msg := fmt.Sprintf("%d organization(s) parsed from the FHIR bundle were not populated due to active=false (i.e. inactive organizations).", inactiveOrgCount)
		log.Warn(msg)
		AppendQueryError(errorFilePath, listSource, msg)
	}

	if len(endpoints) > 0 {
		return endpoints, false
	}

	return nil, true
}

// validateBundleResponse checks that respBody looks like a parseable FHIR Bundle
// before it is handed to BundleToLanternFormat.
func validateBundleResponse(respBody []byte) error {
	trimmed := strings.TrimSpace(string(respBody))
	switch {
	case len(trimmed) == 0:
		return fmt.Errorf("FHIR bundle parsing failed: empty response body")
	case strings.HasPrefix(trimmed, "<"):
		return fmt.Errorf("FHIR bundle parsing failed: non-JSON (HTML/XML) response")
	case !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "["):
		return fmt.Errorf("FHIR bundle parsing failed: unexpected response format, first bytes: %.30s", trimmed)
	case !json.Valid(respBody):
		return fmt.Errorf("FHIR bundle parsing failed: malformed JSON")
	}

	var resourceCheck struct {
		ResourceType string `json:"resourceType"`
	}
	if json.Unmarshal(respBody, &resourceCheck) == nil &&
		resourceCheck.ResourceType != "" &&
		resourceCheck.ResourceType != "Bundle" {
		return fmt.Errorf("FHIR bundle parsing failed: JSON resource is not of type Bundle (got %s)",
			resourceCheck.ResourceType)
	}

	return nil
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
