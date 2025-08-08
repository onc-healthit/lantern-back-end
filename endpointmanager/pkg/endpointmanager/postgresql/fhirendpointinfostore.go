package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/smartparser"
	log "github.com/sirupsen/logrus"
)

// prepared statements are left open to be used throughout the execution of the application
var addFHIREndpointInfoStatement *sql.Stmt
var updateFHIREndpointInfoStatement *sql.Stmt
var deleteFHIREndpointInfoStatement *sql.Stmt
var deleteFHIREndpointInfoOldEntriesStatement *sql.Stmt
var updateFHIREndpointInfoMetadataStatement *sql.Stmt
var getFHIREndpointsByURLAndDifferentRequestedVersion *sql.Stmt

// GetFHIREndpointInfo gets a FHIREndpointInfo from the database using the database id as a key.
// If the FHIREndpointInfo does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetFHIREndpointInfo(ctx context.Context, id int) (*endpointmanager.FHIREndpointInfo, error) {
	var endpointInfo endpointmanager.FHIREndpointInfo
	var capabilityStatementJSON []byte
	var includedFieldsJSON []byte
	var supportedProfilesJSON []byte
	var healthitProductIDNullable sql.NullInt64
	var validationResultIDNullable sql.NullInt64
	var vendorIDNullable sql.NullInt64
	var smartResponseJSON []byte
	var operResourceJSON []byte
	var metadataID int

	sqlStatementInfo := `
	SELECT
		id,
		url,
		healthit_mapping_id,
		vendor_id,
		tls_version,
		mime_types,
		capability_statement,
		created_at,
		updated_at,
		smart_response,
		included_fields,
		operation_resource,
		supported_profiles,
		validation_result_id,
		metadata_id,
		requested_fhir_version,
		capability_fhir_version
	FROM fhir_endpoints_info WHERE id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatementInfo, id)

	err := row.Scan(
		&endpointInfo.ID,
		&endpointInfo.URL,
		&healthitProductIDNullable,
		&vendorIDNullable,
		&endpointInfo.TLSVersion,
		pq.Array(&endpointInfo.MIMETypes),
		&capabilityStatementJSON,
		&endpointInfo.CreatedAt,
		&endpointInfo.UpdatedAt,
		&smartResponseJSON,
		&includedFieldsJSON,
		&operResourceJSON,
		&supportedProfilesJSON,
		&validationResultIDNullable,
		&metadataID,
		&endpointInfo.RequestedFhirVersion,
		&endpointInfo.CapabilityFhirVersion)
	if err != nil {
		return nil, err
	}

	if capabilityStatementJSON != nil {
		endpointInfo.CapabilityStatement, err = capabilityparser.NewCapabilityStatement(capabilityStatementJSON)
		if err != nil {
			return nil, err
		}
	}

	ints := getRegularInts([]sql.NullInt64{healthitProductIDNullable, vendorIDNullable, validationResultIDNullable})
	endpointInfo.HealthITProductID = ints[0]
	endpointInfo.VendorID = ints[1]
	endpointInfo.ValidationID = ints[2]

	if includedFieldsJSON != nil {
		err = json.Unmarshal(includedFieldsJSON, &endpointInfo.IncludedFields)
		if err != nil {
			return nil, err
		}
	}
	if operResourceJSON != nil {
		err = json.Unmarshal(operResourceJSON, &endpointInfo.OperationResource)
		if err != nil {
			return nil, err
		}
	}
	if supportedProfilesJSON != nil {
		err = json.Unmarshal(supportedProfilesJSON, &endpointInfo.SupportedProfiles)
		if err != nil {
			return nil, err
		}
	}

	if smartResponseJSON != nil {
		endpointInfo.SMARTResponse, err = smartparser.NewSMARTResp(smartResponseJSON)
		if err != nil {
			return nil, err
		}
	}

	endpointMetadata, err := s.GetFHIREndpointMetadata(ctx, metadataID)
	if err != nil {
		return nil, err
	}
	endpointInfo.Metadata = endpointMetadata

	return &endpointInfo, err
}

// GetFHIREndpointInfosUsingURL gets all the FHIREndpointInfo objects that correspond to the FHIREndpoints with the given URL.
func (s *Store) GetFHIREndpointInfosUsingURL(ctx context.Context, url string) ([]*endpointmanager.FHIREndpointInfo, error) {
	var endpointInfos []*endpointmanager.FHIREndpointInfo
	var operResourceJSON []byte
	sqlStatementInfo := `
	SELECT
		id,
		url,
		healthit_mapping_id,
		vendor_id,
		tls_version,
		mime_types,
		capability_statement,
		validation_result_id,
		created_at,
		updated_at,
		smart_response,
		included_fields,
		operation_resource,
		supported_profiles,
		metadata_id,
		requested_fhir_version,
		capability_fhir_version
	FROM fhir_endpoints_info WHERE fhir_endpoints_info.url = $1`

	rows, err := s.DB.QueryContext(ctx, sqlStatementInfo, url)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var endpointInfo endpointmanager.FHIREndpointInfo
		var capabilityStatementJSON []byte
		var includedFieldsJSON []byte
		var supportedProfilesJSON []byte
		var healthitProductIDNullable sql.NullInt64
		var validationResultIDNullable sql.NullInt64
		var vendorIDNullable sql.NullInt64
		var smartResponseJSON []byte
		var metadataID int

		err := rows.Scan(
			&endpointInfo.ID,
			&endpointInfo.URL,
			&healthitProductIDNullable,
			&vendorIDNullable,
			&endpointInfo.TLSVersion,
			pq.Array(&endpointInfo.MIMETypes),
			&capabilityStatementJSON,
			&validationResultIDNullable,
			&endpointInfo.CreatedAt,
			&endpointInfo.UpdatedAt,
			&smartResponseJSON,
			&includedFieldsJSON,
			&operResourceJSON,
			&supportedProfilesJSON,
			&metadataID,
			&endpointInfo.RequestedFhirVersion,
			&endpointInfo.CapabilityFhirVersion)
		if err != nil {
			return nil, err
		}

		if capabilityStatementJSON != nil {
			endpointInfo.CapabilityStatement, err = capabilityparser.NewCapabilityStatement(capabilityStatementJSON)
			if err != nil {
				return nil, err
			}
		}

		ints := getRegularInts([]sql.NullInt64{healthitProductIDNullable, vendorIDNullable, validationResultIDNullable})
		endpointInfo.HealthITProductID = ints[0]
		endpointInfo.VendorID = ints[1]
		endpointInfo.ValidationID = ints[2]

		if includedFieldsJSON != nil {
			err = json.Unmarshal(includedFieldsJSON, &endpointInfo.IncludedFields)
			if err != nil {
				return nil, err
			}
		}

		if supportedProfilesJSON != nil {
			err = json.Unmarshal(supportedProfilesJSON, &endpointInfo.SupportedProfiles)
			if err != nil {
				return nil, err
			}
		}

		if smartResponseJSON != nil {
			endpointInfo.SMARTResponse, err = smartparser.NewSMARTResp(smartResponseJSON)
			if err != nil {
				return nil, err
			}
		}

		endpointMetadata, err := s.GetFHIREndpointMetadata(ctx, metadataID)
		if err != nil {
			return nil, err
		}
		endpointInfo.Metadata = endpointMetadata

		endpointInfos = append(endpointInfos, &endpointInfo)

	}

	return endpointInfos, err
}

// GetFHIREndpointInfoUsingURLAndRequestedVersion gets the FHIREndpointInfo object that corresponds to the FHIREndpoint with the given URL and requestVersion
func (s *Store) GetFHIREndpointInfoUsingURLAndRequestedVersion(ctx context.Context, url string, requestedVersion string) (*endpointmanager.FHIREndpointInfo, error) {
	var endpointInfo endpointmanager.FHIREndpointInfo
	var capabilityStatementJSON []byte
	var includedFieldsJSON []byte
	var supportedProfilesJSON []byte
	var healthitProductIDNullable sql.NullInt64
	var validationResultIDNullable sql.NullInt64
	var vendorIDNullable sql.NullInt64
	var smartResponseJSON []byte
	var operResourceJSON []byte
	var metadataID int

	sqlStatementInfo := `
	SELECT
		id,
		url,
		healthit_mapping_id,
		vendor_id,
		tls_version,
		mime_types,
		capability_statement,
		created_at,
		updated_at,
		smart_response,
		included_fields,
		operation_resource,
		supported_profiles,
		validation_result_id,
		metadata_id,
		requested_fhir_version,
		capability_fhir_version
	FROM fhir_endpoints_info WHERE fhir_endpoints_info.url = $1 AND fhir_endpoints_info.requested_fhir_version = $2 LIMIT 1`

	row := s.DB.QueryRowContext(ctx, sqlStatementInfo, url, requestedVersion)

	err := row.Scan(
		&endpointInfo.ID,
		&endpointInfo.URL,
		&healthitProductIDNullable,
		&vendorIDNullable,
		&endpointInfo.TLSVersion,
		pq.Array(&endpointInfo.MIMETypes),
		&capabilityStatementJSON,
		&endpointInfo.CreatedAt,
		&endpointInfo.UpdatedAt,
		&smartResponseJSON,
		&includedFieldsJSON,
		&operResourceJSON,
		&supportedProfilesJSON,
		&validationResultIDNullable,
		&metadataID,
		&endpointInfo.RequestedFhirVersion,
		&endpointInfo.CapabilityFhirVersion)
	if err != nil {
		return nil, err
	}

	if capabilityStatementJSON != nil {
		endpointInfo.CapabilityStatement, err = capabilityparser.NewCapabilityStatement(capabilityStatementJSON)
		if err != nil {
			return nil, err
		}
	}

	ints := getRegularInts([]sql.NullInt64{healthitProductIDNullable, vendorIDNullable, validationResultIDNullable})
	endpointInfo.HealthITProductID = ints[0]
	endpointInfo.VendorID = ints[1]
	endpointInfo.ValidationID = ints[2]

	if includedFieldsJSON != nil {
		err = json.Unmarshal(includedFieldsJSON, &endpointInfo.IncludedFields)
		if err != nil {
			return nil, err
		}
	}

	if operResourceJSON != nil {
		err = json.Unmarshal(operResourceJSON, &endpointInfo.OperationResource)
		if err != nil {
			return nil, err
		}
	}

	if supportedProfilesJSON != nil {
		err = json.Unmarshal(supportedProfilesJSON, &endpointInfo.SupportedProfiles)
		if err != nil {
			return nil, err
		}
	}

	if smartResponseJSON != nil {
		endpointInfo.SMARTResponse, err = smartparser.NewSMARTResp(smartResponseJSON)
		if err != nil {
			return nil, err
		}
	}

	endpointMetadata, err := s.GetFHIREndpointMetadata(ctx, metadataID)
	if err != nil {
		return nil, err
	}
	endpointInfo.Metadata = endpointMetadata

	return &endpointInfo, err
}

// GetFHIREndpointInfoValidation gets the validation object for the given FhirEndpointInfo object
func (s *Store) GetFHIREndpointInfoValidation(ctx context.Context, e *endpointmanager.FHIREndpointInfo) (*endpointmanager.Validation, error) {
	validationRows, err := s.GetValidationByID(ctx, e.ValidationID)
	if err != nil {
		return nil, err
	}
	validationObj := endpointmanager.Validation{
		Results: *validationRows,
	}
	return &validationObj, nil
}

// AddFHIREndpointInfo adds the FHIREndpointInfo to the database.
func (s *Store) AddFHIREndpointInfo(ctx context.Context, e *endpointmanager.FHIREndpointInfo, metadataID int) error {
	var err error
	var capabilityStatementJSON []byte

	log.Infof("Adding FHIREndpointInfo with URL: %s", e.URL)

	if e.CapabilityStatementBytes != nil {
		capabilityStatementJSON = e.CapabilityStatementBytes
	} else {
		capabilityStatementJSON = []byte("null")
	}

	includedFieldsJSON, err := json.Marshal(e.IncludedFields)
	if err != nil {
		return err
	}

	operResourceJSON, err := json.Marshal(e.OperationResource)
	if err != nil {
		return err
	}

	supportedProfilesJSON, err := json.Marshal(e.SupportedProfiles)
	if err != nil {
		return err
	}

	var smartResponseJSON []byte
	if e.SMARTResponseBytes != nil {
		smartResponseJSON = e.SMARTResponseBytes
	} else {
		smartResponseJSON = []byte("null")
	}

	nullableInts := getNullableInts([]int{e.HealthITProductID, e.VendorID, e.ValidationID})

	row := addFHIREndpointInfoStatement.QueryRowContext(ctx,
		e.URL,
		nullableInts[0],
		nullableInts[1],
		e.TLSVersion,
		pq.Array(e.MIMETypes),
		capabilityStatementJSON,
		smartResponseJSON,
		includedFieldsJSON,
		operResourceJSON,
		supportedProfilesJSON,
		nullableInts[2],
		metadataID,
		e.RequestedFhirVersion,
		e.CapabilityFhirVersion)

	err = row.Scan(&e.ID)

	return err
}

// UpdateFHIREndpointInfo updates the FHIREndpointInfo in the database using the FHIREndpointInfo's database id as the key.
func (s *Store) UpdateFHIREndpointInfo(ctx context.Context, e *endpointmanager.FHIREndpointInfo, metadataID int) error {
	var err error
	var capabilityStatementJSON []byte

	if e.CapabilityStatementBytes != nil {
		capabilityStatementJSON = e.CapabilityStatementBytes
	} else if e.CapabilityStatement != nil {
		capabilityStatementJSON, err = e.CapabilityStatement.GetJSON()
		if err != nil {
			return err
		}
	} else {
		capabilityStatementJSON = []byte("null")
	}

	includedFieldsJSON, err := json.Marshal(e.IncludedFields)
	if err != nil {
		return err
	}

	operResourceJSON, err := json.Marshal(e.OperationResource)
	if err != nil {
		return err
	}

	supportedProfilesJSON, err := json.Marshal(e.SupportedProfiles)
	if err != nil {
		return err
	}

	var smartResponseJSON []byte
	if e.SMARTResponseBytes != nil {
		smartResponseJSON = e.SMARTResponseBytes
	} else if e.SMARTResponse != nil {
		smartResponseJSON, err = e.SMARTResponse.GetJSON()
		if err != nil {
			return err
		}
	} else {
		smartResponseJSON = []byte("null")
	}

	nullableInts := getNullableInts([]int{e.HealthITProductID, e.VendorID, e.ValidationID})

	_, err = updateFHIREndpointInfoStatement.ExecContext(ctx,
		e.URL,
		nullableInts[0],
		nullableInts[1],
		e.TLSVersion,
		pq.Array(e.MIMETypes),
		capabilityStatementJSON,
		smartResponseJSON,
		includedFieldsJSON,
		operResourceJSON,
		supportedProfilesJSON,
		nullableInts[2],
		metadataID,
		e.RequestedFhirVersion,
		e.CapabilityFhirVersion,
		e.ID)

	return err
}

// UpdateMetadataIDInfo only updates the metadata_id in the info table without affecting the info history table
func (s *Store) UpdateMetadataIDInfo(ctx context.Context, metadataID int, id int) error {
	_, err := s.DB.ExecContext(ctx, "SELECT set_config('metadata.setting', 'TRUE', 'FALSE');")
	if err != nil {
		return err
	}
	_, err = updateFHIREndpointInfoMetadataStatement.ExecContext(ctx, metadataID, id)
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, "SELECT set_config('metadata.setting', 'FALSE', 'FALSE');")
	if err != nil {
		return err
	}

	return err
}

// DeleteFHIREndpointInfo deletes the FHIREndpointInfo from the database using the FHIREndpointInfo's database id  as the key.
func (s *Store) DeleteFHIREndpointInfo(ctx context.Context, e *endpointmanager.FHIREndpointInfo) error {
	_, err := deleteFHIREndpointInfoStatement.ExecContext(ctx, e.URL, e.RequestedFhirVersion)
	return err
}

// deleteFHIREndpointInfoOldEntries deletes the FHIREndpointInfo from the database using the FHIREndpointInfo's database id  as the key.
func (s *Store) DeleteFHIREndpointInfoOldEntries(ctx context.Context) error {
	_, err := deleteFHIREndpointInfoOldEntriesStatement.ExecContext(ctx)
	return err
}

// GetFHIREndpointInfosByURLWithDifferentRequestedVersion gets all FHIREndpointInfo rows for the given url whose RequestedFhirVersion does not exist in the versions list
func (s *Store) GetFHIREndpointInfosByURLWithDifferentRequestedVersion(ctx context.Context, url string, versions []string) ([]*endpointmanager.FHIREndpointInfo, error) {
	var endpointInfos []*endpointmanager.FHIREndpointInfo
	var operResourceJSON []byte

	// Convert array of strings to a string that postgres can convert back to an sql ARRAY
	versionsString := strings.Join(versions, ",")

	rows, err := getFHIREndpointsByURLAndDifferentRequestedVersion.QueryContext(ctx, url, versionsString)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var endpointInfo endpointmanager.FHIREndpointInfo
		var capabilityStatementJSON []byte
		var includedFieldsJSON []byte
		var supportedProfilesJSON []byte
		var healthitProductIDNullable sql.NullInt64
		var validationResultIDNullable sql.NullInt64
		var vendorIDNullable sql.NullInt64
		var smartResponseJSON []byte
		var metadataID int

		err := rows.Scan(
			&endpointInfo.ID,
			&endpointInfo.URL,
			&healthitProductIDNullable,
			&vendorIDNullable,
			&endpointInfo.TLSVersion,
			pq.Array(&endpointInfo.MIMETypes),
			&capabilityStatementJSON,
			&validationResultIDNullable,
			&endpointInfo.CreatedAt,
			&endpointInfo.UpdatedAt,
			&smartResponseJSON,
			&includedFieldsJSON,
			&operResourceJSON,
			&supportedProfilesJSON,
			&metadataID,
			&endpointInfo.RequestedFhirVersion,
			&endpointInfo.CapabilityFhirVersion)
		if err != nil {
			return nil, err
		}

		if capabilityStatementJSON != nil {
			endpointInfo.CapabilityStatement, err = capabilityparser.NewCapabilityStatement(capabilityStatementJSON)
			if err != nil {
				return nil, err
			}
		}

		ints := getRegularInts([]sql.NullInt64{healthitProductIDNullable, vendorIDNullable, validationResultIDNullable})
		endpointInfo.HealthITProductID = ints[0]
		endpointInfo.VendorID = ints[1]
		endpointInfo.ValidationID = ints[2]

		if includedFieldsJSON != nil {
			err = json.Unmarshal(includedFieldsJSON, &endpointInfo.IncludedFields)
			if err != nil {
				return nil, err
			}
		}

		if supportedProfilesJSON != nil {
			err = json.Unmarshal(supportedProfilesJSON, &endpointInfo.SupportedProfiles)
			if err != nil {
				return nil, err
			}
		}

		if smartResponseJSON != nil {
			endpointInfo.SMARTResponse, err = smartparser.NewSMARTResp(smartResponseJSON)
			if err != nil {
				return nil, err
			}
		}

		endpointMetadata, err := s.GetFHIREndpointMetadata(ctx, metadataID)
		if err != nil {
			return nil, err
		}
		endpointInfo.Metadata = endpointMetadata

		endpointInfos = append(endpointInfos, &endpointInfo)

	}

	return endpointInfos, err
}

func prepareFHIREndpointInfoStatements(s *Store) error {
	var err error
	addFHIREndpointInfoStatement, err = s.DB.Prepare(`
		INSERT INTO fhir_endpoints_info (
			url,
			healthit_mapping_id,
			vendor_id,
			tls_version,
			mime_types,
			capability_statement,
			smart_response,
			included_fields,
			operation_resource,
			supported_profiles,
			validation_result_id,
			metadata_id,
			requested_fhir_version,
			capability_fhir_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id`)
	if err != nil {
		return err
	}
	updateFHIREndpointInfoStatement, err = s.DB.Prepare(`
		UPDATE fhir_endpoints_info
		SET 
		    url = $1,
		    healthit_mapping_id = $2,
			vendor_id = $3,
			tls_version = $4,
			mime_types = $5,
			capability_statement = $6,
			smart_response = $7,
			included_fields = $8,
			operation_resource = $9,
			supported_profiles = $10,
			validation_result_id = $11,
			metadata_id = $12,
			requested_fhir_version = $13,
			capability_fhir_version = $14		
		WHERE id = $15`)
	if err != nil {
		return err
	}
	updateFHIREndpointInfoMetadataStatement, err = s.DB.Prepare(`
		UPDATE fhir_endpoints_info
		SET 
			metadata_id = $1		
		WHERE id = $2`)
	if err != nil {
		return err
	}
	deleteFHIREndpointInfoStatement, err = s.DB.Prepare(`
        DELETE FROM fhir_endpoints_info
        WHERE url = $1 AND requested_fhir_version = $2`)
	if err != nil {
		return err
	}
	deleteFHIREndpointInfoOldEntriesStatement, err = s.DB.Prepare(`
		DELETE FROM fhir_endpoints_info 
		WHERE url NOT IN (SELECT url FROM fhir_endpoints)`)
	if err != nil {
		return err
	}
	getFHIREndpointsByURLAndDifferentRequestedVersion, err = s.DB.Prepare(`
		SELECT
		id,
		url,
		healthit_mapping_id,
		vendor_id,
		tls_version,
		mime_types,
		capability_statement,
		validation_result_id,
		created_at,
		updated_at,
		smart_response,
		included_fields,
		operation_resource,
		supported_profiles,
		metadata_id,
		requested_fhir_version,
		capability_fhir_version
		FROM fhir_endpoints_info WHERE fhir_endpoints_info.url = $1 AND NOT (fhir_endpoints_info.requested_fhir_version = ANY (string_to_array($2,',','')))`)
	if err != nil {
		return err
	}
	return nil
}
