package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/capabilityparser"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// prepared statements are left open to be used throughout the execution of the application
// TODO: figure out if there's a better way to manage this for bulk calls
var addFHIREndpointInfoStatement *sql.Stmt
var updateFHIREndpointInfoStatement *sql.Stmt
var deleteFHIREndpointInfoStatement *sql.Stmt

// GetFHIREndpointInfo gets a FHIREndpointInfo from the database using the database id as a key.
// If the FHIREndpointInfo does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetFHIREndpointInfo(ctx context.Context, id int) (*endpointmanager.FHIREndpointInfo, error) {
	var endpointInfo endpointmanager.FHIREndpointInfo
	var capabilityStatementJSON []byte
	var validationJSON []byte

	sqlStatement := `
	SELECT
		id,
		fhir_endpoint_id,
		healthit_product_id,
		tls_version,
		mime_types,
		http_response,
		errors,
		organization_name,
		vendor,
		capability_statement,
		validation,
		created_at,
		updated_at
	FROM fhir_endpoints WHERE id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&endpointInfo.ID,
		&endpointInfo.FHIREndpointID,
		&endpointInfo.HealthITProductID,
		&endpointInfo.TLSVersion,
		pq.Array(&endpointInfo.MIMETypes),
		&endpointInfo.HTTPResponse,
		&endpointInfo.Errors,
		&endpointInfo.OrganizationName,
		&endpointInfo.Vendor,
		&capabilityStatementJSON,
		&validationJSON,
		&endpointInfo.CreatedAt,
		&endpointInfo.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if capabilityStatementJSON != nil {
		endpointInfo.CapabilityStatement, err = capabilityparser.NewCapabilityStatement(capabilityStatementJSON)
		if err != nil {
			return nil, err
		}
	}

	err = json.Unmarshal(validationJSON, &endpointInfo.Validation)

	return &endpointInfo, err
}

// GetFHIREndpointInfoUsingURL gets a FHIREndpointInfo from the database using the given url as a key.
// If the FHIREndpointInfo does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetFHIREndpointInfoUsingURL(ctx context.Context, url string) (*endpointmanager.FHIREndpointInfo, error) {
	var endpointInfo endpointmanager.FHIREndpointInfo
	var capabilityStatementJSON []byte
	var validationJSON []byte

	sqlStatement := `
	SELECT
		id,
		fhir_endpoint_id,
		healthit_product_id,
		tls_version,
		mime_types,
		http_response,
		errors,
		organization_name,
		vendor,
		list_source,
		capability_statement,
		validation,
		created_at,
		updated_at
	FROM fhir_endpoints WHERE url=$1`

	row := s.DB.QueryRowContext(ctx, sqlStatement, url)

	err := row.Scan(
		&endpointInfo.ID,
		&endpointInfo.FHIREndpointID,
		&endpointInfo.HealthITProductID,
		&endpointInfo.TLSVersion,
		pq.Array(&endpointInfo.MIMETypes),
		&endpointInfo.HTTPResponse,
		&endpointInfo.Errors,
		&endpointInfo.OrganizationName,
		&endpointInfo.Vendor,
		&capabilityStatementJSON,
		&validationJSON,
		&endpointInfo.CreatedAt,
		&endpointInfo.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if capabilityStatementJSON != nil {
		endpointInfo.CapabilityStatement, err = capabilityparser.NewCapabilityStatement(capabilityStatementJSON)
		if err != nil {
			return nil, err
		}
	}

	err = json.Unmarshal(validationJSON, &endpointInfo.Validation)

	return &endpointInfo, err
}

// AddFHIREndpointInfo adds the FHIREndpointInfo to the database.
func (s *Store) AddFHIREndpointInfo(ctx context.Context, e *endpointmanager.FHIREndpointInfo) error {
	var err error
	var capabilityStatementJSON []byte
	if e.CapabilityStatement != nil {
		capabilityStatementJSON, err = e.CapabilityStatement.GetJSON()
		if err != nil {
			return err
		}
	} else {
		capabilityStatementJSON = []byte("null")
	}
	validationJSON, err := json.Marshal(e.Validation)
	if err != nil {
		return err
	}

	row := addFHIREndpointInfoStatement.QueryRowContext(ctx,
		e.FHIREndpointID,
		e.HealthITProductID,
		e.TLSVersion,
		pq.Array(e.MIMETypes),
		e.HTTPResponse,
		e.Errors,
		e.OrganizationName,
		e.Vendor,
		capabilityStatementJSON,
		validationJSON)

	err = row.Scan(&e.ID)

	return err
}

// UpdateFHIREndpointInfo updates the FHIREndpointInfo in the database using the FHIREndpointInfo's database id as the key.
func (s *Store) UpdateFHIREndpointInfo(ctx context.Context, e *endpointmanager.FHIREndpointInfo) error {
	var err error
	var capabilityStatementJSON []byte
	if e.CapabilityStatement != nil {
		capabilityStatementJSON, err = e.CapabilityStatement.GetJSON()
		if err != nil {
			return err
		}
	} else {
		capabilityStatementJSON = []byte("null")
	}
	validationJSON, err := json.Marshal(e.Validation)
	if err != nil {
		return err
	}

	_, err = updateFHIREndpointInfoStatement.ExecContext(ctx,
		e.FHIREndpointID,
		e.HealthITProductID,
		e.TLSVersion,
		pq.Array(e.MIMETypes),
		e.HTTPResponse,
		e.Errors,
		e.OrganizationName,
		e.Vendor,
		capabilityStatementJSON,
		validationJSON,
		e.ID)

	return err
}

// DeleteFHIREndpointInfo deletes the FHIREndpointInfo from the database using the FHIREndpointInfo's database id  as the key.
func (s *Store) DeleteFHIREndpointInfo(ctx context.Context, e *endpointmanager.FHIREndpointInfo) error {
	_, err := deleteFHIREndpointInfoStatement.ExecContext(ctx, e.ID)

	return err
}

// GetAllFHIREndpointInfoOrgNames returns a sql.Rows of all of the orgNames
func (s *Store) GetAllFHIREndpointInfoOrgNames(ctx context.Context) ([]endpointmanager.FHIREndpointInfo, error) {
	sqlStatement := `
        SELECT id, organization_name FROM fhir_endpoints`
	rows, err := s.DB.QueryContext(ctx, sqlStatement)

	if err != nil {
		return nil, err
	}
	var endpoints []endpointmanager.FHIREndpointInfo
	defer rows.Close()
	for rows.Next() {
		var endpointInfo endpointmanager.FHIREndpointInfo
		err = rows.Scan(&endpointInfo.ID, &endpointInfo.OrganizationName)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, endpointInfo)
	}
	return endpoints, nil
}

func prepareFHIREndpointInfoStatements(s *Store) error {
	var err error
	addFHIREndpointInfoStatement, err = s.DB.Prepare(`
		INSERT INTO fhir_endpoints (
			fhir_endpoint_id,
			healthit_product_id,
			tls_version,
			mime_types,
			http_response,
			errors,
			organization_name,
			vendor,
			capability_statement,
			validation)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`)
	if err != nil {
		return err
	}
	updateFHIREndpointInfoStatement, err = s.DB.Prepare(`
		UPDATE fhir_endpoints
		SET 
		    fhir_endpoint_id = $1,
		    healthit_product_id = $2,
			tls_version = $3,
			mime_types = $4,
			http_response = $5,
			errors = $6,
			organization_name = $7,
			vendor = $8,
			capability_statement = $9,
			validation = $10
		WHERE id = $11`)
	if err != nil {
		return err
	}
	deleteFHIREndpointInfoStatement, err = s.DB.Prepare(`
        DELETE FROM fhir_endpoints
        WHERE id = $1`)
	if err != nil {
		return err
	}
	return nil
}
