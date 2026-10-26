package payervalidator

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	log "github.com/sirupsen/logrus"
)

// PayerRegistration represents the payer registration data structure
type PayerRegistration struct {
	ID                 int                    `db:"id"`
	PayerID            int                    `db:"payer_id"`
	URL                string                 `db:"url"`
	Name               string                 `db:"name"`
	EDIID              *int                   `db:"edi_id"`
	Address            map[string]interface{} `db:"address"`
	IsPersisted        bool                   `db:"is_persisted"`
	UserFacingURL      string                 `db:"user_facing_url"`
	ValidationResult   *bool                  `db:"validation_result"`
	ValidationComments *string                `db:"validation_comments"` // Changed to pointer to handle NULL
	ContactName        string                 `db:"contact_name"`
	ContactEmail       string                 `db:"contact_email"`
	EndpointType       string                 `db:"endpoint_type"`
	CreatedAt          time.Time              `db:"created_at"`
	UpdatedAt          time.Time              `db:"updated_at"`
}

// ValidationResult represents the result of endpoint validation
type ValidationResult struct {
	IsValid      bool
	HTTPStatus   int
	ErrorMessage string
}

// Validator handles the validation and processing of payer registrations
type Validator struct {
	store      *postgresql.Store
	httpClient *http.Client
	userAgent  string
}

// NewValidatorWithStore creates a new payer registration validator with an existing store
func NewValidatorWithStore(store *postgresql.Store) (*Validator, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// Read version file for user agent
	version, err := os.ReadFile("/etc/lantern/VERSION")
	userAgent := "LANTERN/payer-validator"
	if err == nil {
		versionString := string(version)
		versionNum := strings.Split(versionString, "=")
		if len(versionNum) > 1 {
			userAgent = "LANTERN/" + strings.TrimSuffix(versionNum[1], "\n") + "-payer-validator"
		}
	}
	log.Infof("User agent: %s", userAgent)

	return &Validator{
		store:      store,
		httpClient: client,
		userAgent:  userAgent,
	}, nil
}

// Close closes the database connection
func (v *Validator) Close() {
	// Don't close the store here as it's managed by main.go
}

// ValidateAndProcessRegistrations processes all unvalidated payer registrations
func (v *Validator) ValidateAndProcessRegistrations(ctx context.Context) error {
	log.Info("Starting payer registration validation process...")

	// Get unvalidated payer registrations
	registrations, err := v.getUnvalidatedRegistrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch unvalidated registrations: %w", err)
	}

	if len(registrations) == 0 {
		log.Info("No unvalidated payer registrations found")
		return nil
	}

	log.Infof("Found %d unvalidated payer registrations", len(registrations))

	// Process each registration
	for _, registration := range registrations {
		log.Infof("Processing registration ID %d for URL: %s", registration.ID, registration.URL)

		// Validate the FHIR endpoint
		validationResult := v.validateEndpoint(ctx, registration.URL)

		// Update the database with validation results
		err = v.updateValidationResult(ctx, registration.ID, validationResult)
		if err != nil {
			log.Errorf("Failed to update validation result for registration ID %d: %v", registration.ID, err)
			continue
		}

		// If validation passed, persist to payer_info table
		if validationResult.IsValid {
			err = v.persistToPayerInfo(ctx, registration)
			if err != nil {
				log.Errorf("Failed to persist registration ID %d to payer_info table: %v", registration.ID, err)
				continue
			}
			log.Infof("Successfully validated and persisted registration ID %d", registration.ID)
		} else {
			// Get the validation comments as a string for logging
			var validationComments string
			if registration.ValidationComments != nil {
				validationComments = *registration.ValidationComments
			} else {
				validationComments = "No validation comments"
			}
			log.Infof("Validation failed for registration ID %d: %s", registration.ID, validationComments)
		}

		// Add delay between requests to be respectful
		time.Sleep(2 * time.Second)
	}

	return nil
}

// ValidateRegistrationsDryRun performs validation without updating database
func (v *Validator) ValidateRegistrationsDryRun(ctx context.Context) error {
	log.Info("Starting payer registration validation process (DRY RUN)...")

	// Get unvalidated payer registrations
	registrations, err := v.getUnvalidatedRegistrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch unvalidated registrations: %w", err)
	}

	if len(registrations) == 0 {
		log.Info("No unvalidated payer registrations found")
		return nil
	}

	log.Infof("Found %d unvalidated payer registrations", len(registrations))

	// Process each registration (validation only)
	successCount := 0
	failureCount := 0

	for _, registration := range registrations {
		log.Infof("[DRY RUN] Processing registration ID %d for URL: %s", registration.ID, registration.URL)

		// Validate the endpoint
		validationResult := v.validateEndpoint(ctx, registration.URL)

		if validationResult.IsValid {
			log.Infof("[DRY RUN] ✓ Validation PASSED for registration ID %d", registration.ID)
			successCount++
		} else {
			// Get the validation comments as a string for logging
			var validationComments string
			if registration.ValidationComments != nil {
				validationComments = *registration.ValidationComments
			} else {
				validationComments = validationResult.ErrorMessage
			}
			log.Warnf("[DRY RUN] ✗ Validation FAILED for registration ID %d: %s", registration.ID, validationComments)
			failureCount++
		}

		// Add delay between requests to be respectful
		time.Sleep(1 * time.Second)
	}

	log.Infof("[DRY RUN] Summary: %d successful validations, %d failed validations", successCount, failureCount)
	return nil
}

// validateEndpoint checks if the endpoint responds with HTTP 200 and returns a valid capability statement
func (v *Validator) validateEndpoint(ctx context.Context, endpointURL string) ValidationResult {
	result := ValidationResult{
		IsValid:      false,
		HTTPStatus:   0,
		ErrorMessage: "",
	}

	// Ensure URL ends with /metadata for capability statement
	metadataURL := strings.TrimSuffix(endpointURL, "/") + "/metadata"

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", metadataURL, nil)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to create HTTP request: %v", err)
		return result
	}

	// Set headers - request JSON but also accept XML as fallback
	req.Header.Set("User-Agent", v.userAgent)
	req.Header.Set("Accept", "application/fhir+json, application/json, application/fhir+xml, application/xml")

	// Make HTTP request
	resp, err := v.httpClient.Do(req)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("HTTP request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.HTTPStatus = resp.StatusCode

	// Check for 200 OK status
	if resp.StatusCode != http.StatusOK {
		result.ErrorMessage = fmt.Sprintf("HTTP status %d: %s", resp.StatusCode, resp.Status)
		return result
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to read response body: %v", err)
		return result
	}

	// Determine content type from response header
	contentType := resp.Header.Get("Content-Type")

	// Check if response is XML based on content type or body content
	isXML := strings.Contains(contentType, "xml") || strings.HasPrefix(strings.TrimSpace(string(body)), "<")

	if isXML {
		// Validate XML capability statement
		resourceType, err := extractXMLResourceType(body)
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("Failed to parse XML response: %v", err)
			return result
		}

		if resourceType != "CapabilityStatement" && resourceType != "Conformance" {
			result.ErrorMessage = fmt.Sprintf("Invalid XML root element: expected 'CapabilityStatement' or 'Conformance', got '%s'", resourceType)
			return result
		}
	} else {
		// Validate JSON capability statement
		var capabilityStatement map[string]interface{}
		err = json.Unmarshal(body, &capabilityStatement)
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("Response is not valid JSON: %v", err)
			return result
		}

		// Check for basic capability statement structure (resourceType should be CapabilityStatement or Conformance)
		resourceType, ok := capabilityStatement["resourceType"].(string)
		if !ok {
			result.ErrorMessage = "Response JSON missing 'resourceType' field"
			return result
		}

		if resourceType != "CapabilityStatement" && resourceType != "Conformance" {
			result.ErrorMessage = fmt.Sprintf("Invalid resourceType: expected 'CapabilityStatement' or 'Conformance', got '%s'", resourceType)
			return result
		}
	}

	// If we get here, validation passed
	result.IsValid = true
	result.ErrorMessage = "Validation passed: HTTP 200 with valid capability statement"

	return result
}

// extractXMLResourceType extracts the root element name from XML to determine the resource type
func extractXMLResourceType(body []byte) (string, error) {
	decoder := xml.NewDecoder(strings.NewReader(string(body)))
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", fmt.Errorf("failed to parse XML: %w", err)
		}
		// Find the first start element (root element)
		if startElement, ok := token.(xml.StartElement); ok {
			return startElement.Name.Local, nil
		}
	}
}

// getUnvalidatedRegistrations fetches all payer registrations that haven't been validated yet
func (v *Validator) getUnvalidatedRegistrations(ctx context.Context) ([]PayerRegistration, error) {
	var registrations []PayerRegistration

	query := `
		SELECT
			pe.id, pe.payer_id, pe.url, pe.name, pe.edi_id, pe.address,
			pe.is_persisted, pe.user_facing_url, pe.validation_result,
			pe.validation_comments, pe.created_at, pe.updated_at,
			p.contact_name, p.contact_email,
			COALESCE(pe.endpoint_type, '') as endpoint_type
		FROM payer_endpoints pe
		JOIN payers p ON pe.payer_id = p.id
		WHERE pe.validation_result IS NULL
		ORDER BY pe.created_at ASC
	`

	rows, err := v.store.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unvalidated registrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var reg PayerRegistration
		var addressJSON []byte

		err := rows.Scan(
			&reg.ID, &reg.PayerID, &reg.URL, &reg.Name, &reg.EDIID, &addressJSON,
			&reg.IsPersisted, &reg.UserFacingURL, &reg.ValidationResult,
			&reg.ValidationComments, &reg.CreatedAt, &reg.UpdatedAt,
			&reg.ContactName, &reg.ContactEmail, &reg.EndpointType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registration row: %w", err)
		}

		// Parse address JSON
		if len(addressJSON) > 0 {
			err = json.Unmarshal(addressJSON, &reg.Address)
			if err != nil {
				log.Warnf("Failed to parse address JSON for registration ID %d: %v", reg.ID, err)
				reg.Address = make(map[string]interface{})
			}
		}

		registrations = append(registrations, reg)
	}

	return registrations, rows.Err()
}

// updateValidationResult updates the payer_endpoints table with validation results
func (v *Validator) updateValidationResult(ctx context.Context, registrationID int, result ValidationResult) error {
	query := `
		UPDATE payer_endpoints 
		SET validation_result = $1, validation_comments = $2, updated_at = NOW()
		WHERE id = $3
	`

	_, err := v.store.DB.ExecContext(ctx, query, result.IsValid, result.ErrorMessage, registrationID)
	if err != nil {
		return fmt.Errorf("failed to update validation result: %w", err)
	}

	return nil
}

// PayerListSource is the list_source value used for payer-submitted endpoints
const PayerListSource = "Payer Self-Registration"

// PayerSourceCategory is the is_chpl value used for payer-submitted endpoints
const PayerSourceCategory = "Payer"

// persistToLanternTables persists validated registration data to Lantern tables and payer_info
func (v *Validator) persistToPayerInfo(ctx context.Context, registration PayerRegistration) error {
	// Convert address to JSON
	addressJSON, err := json.Marshal(registration.Address)
	if err != nil {
		addressJSON = []byte("{}")
	}

	// Begin transaction
	tx, err := v.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Insert into fhir_endpoints table (main Lantern endpoint registry)
	// Use ON CONFLICT to handle re-submissions of the same URL
	fhirEndpointsQuery := `
		INSERT INTO fhir_endpoints (url, list_source)
		VALUES ($1, $2)
		ON CONFLICT (url, list_source) DO NOTHING
	`
	_, err = tx.ExecContext(ctx, fhirEndpointsQuery, registration.URL, PayerListSource)
	if err != nil {
		return fmt.Errorf("failed to insert into fhir_endpoints: %w", err)
	}
	log.Infof("Inserted endpoint into fhir_endpoints: %s", registration.URL)

	// 2. Insert/update list_source_info table (for source filtering)
	listSourceInfoQuery := `
		INSERT INTO list_source_info (list_source, is_chpl, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (list_source) DO UPDATE SET
			updated_at = NOW(),
			is_chpl = $2
	`
	_, err = tx.ExecContext(ctx, listSourceInfoQuery, PayerListSource, PayerSourceCategory)
	if err != nil {
		return fmt.Errorf("failed to insert into list_source_info: %w", err)
	}
	log.Infof("Updated list_source_info with source category: %s", PayerSourceCategory)

	// 3. Insert into payer_info table (payer-specific metadata)
	payerInfoQuery := `
		INSERT INTO payer_info (url, edi_id, endpoint_type, address, user_facing_url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING
	`
	_, err = tx.ExecContext(ctx, payerInfoQuery, registration.URL, registration.EDIID, registration.EndpointType, addressJSON, registration.UserFacingURL)
	if err != nil {
		return fmt.Errorf("failed to insert into payer_info: %w", err)
	}

	// 4. Mark as persisted in payer_endpoints
	markPersistedQuery := `
		UPDATE payer_endpoints
		SET is_persisted = true, updated_at = NOW()
		WHERE id = $1
	`
	_, err = tx.ExecContext(ctx, markPersistedQuery, registration.ID)
	if err != nil {
		return fmt.Errorf("failed to mark as persisted: %w", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Infof("Successfully persisted payer endpoint to Lantern tables: %s", registration.URL)
	return nil
}
