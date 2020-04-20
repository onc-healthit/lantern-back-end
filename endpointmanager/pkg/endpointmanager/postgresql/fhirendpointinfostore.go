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
	var fhirEndpointIDNullable sql.NullInt64
	var healthitProductIDNullable sql.NullInt64

	sqlStatement := `
	SELECT
		id,
		fhir_endpoint_id,
		healthit_product_id,
		tls_version,
		mime_types,
		http_response,
		errors,
		vendor,
		capability_statement,
		validation,
		created_at,
		updated_at
	FROM fhir_endpoints_info WHERE id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&endpointInfo.ID,
		&fhirEndpointIDNullable,
		&healthitProductIDNullable,
		&endpointInfo.TLSVersion,
		pq.Array(&endpointInfo.MIMETypes),
		&endpointInfo.HTTPResponse,
		&endpointInfo.Errors,
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

	ints := getRegularInts([]sql.NullInt64{fhirEndpointIDNullable, healthitProductIDNullable})
	endpointInfo.FHIREndpointID = ints[0]
	endpointInfo.HealthITProductID = ints[1]

	err = json.Unmarshal(validationJSON, &endpointInfo.Validation)

	return &endpointInfo, err
}

// GetFHIREndpointInfoUsingFHIREndpointID gets the FHIREndpointInfo object that corresponds to the FHIREndpoint with the given ID.
func (s *Store) GetFHIREndpointInfoUsingFHIREndpointID(ctx context.Context, id int) (*endpointmanager.FHIREndpointInfo, error) {
	var endpointInfo endpointmanager.FHIREndpointInfo
	var capabilityStatementJSON []byte
	var validationJSON []byte
	var fhirEndpointIDNullable sql.NullInt64
	var healthitProductIDNullable sql.NullInt64

	sqlStatement := `
	SELECT
		id,
		fhir_endpoint_id,
		healthit_product_id,
		tls_version,
		mime_types,
		http_response,
		errors,
		vendor,
		capability_statement,
		validation,
		created_at,
		updated_at
	FROM fhir_endpoints_info WHERE fhir_endpoints_info.fhir_endpoint_id = $1`

	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&endpointInfo.ID,
		&fhirEndpointIDNullable,
		&healthitProductIDNullable,
		&endpointInfo.TLSVersion,
		pq.Array(&endpointInfo.MIMETypes),
		&endpointInfo.HTTPResponse,
		&endpointInfo.Errors,
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

	ints := getRegularInts([]sql.NullInt64{fhirEndpointIDNullable, healthitProductIDNullable})
	endpointInfo.FHIREndpointID = ints[0]
	endpointInfo.HealthITProductID = ints[1]

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

	nullableInts := getNullableInts([]int{e.FHIREndpointID, e.HealthITProductID})

	row := addFHIREndpointInfoStatement.QueryRowContext(ctx,
		nullableInts[0],
		nullableInts[1],
		e.TLSVersion,
		pq.Array(e.MIMETypes),
		e.HTTPResponse,
		e.Errors,
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

	nullableInts := getNullableInts([]int{e.FHIREndpointID, e.HealthITProductID})

	_, err = updateFHIREndpointInfoStatement.ExecContext(ctx,
		nullableInts[0],
		nullableInts[1],
		e.TLSVersion,
		pq.Array(e.MIMETypes),
		e.HTTPResponse,
		e.Errors,
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

// converts foreign key ints to nullable ints so we don't have issues with non-existent foreign key references.
func getNullableInts(regularInts []int) []sql.NullInt64 {
	nullableInts := make([]sql.NullInt64, len(regularInts))

	for i, regInt := range regularInts {
		var nullInt sql.NullInt64
		if regInt < 1 {
			nullInt.Valid = false
		} else {
			nullInt.Valid = true
			nullInt.Int64 = int64(regInt)
		}
		nullableInts[i] = nullInt
	}
	return nullableInts
}

// converts nullable into to an integer. null values are made to be 0s. This should only be used for foreign key references. postgres does not use 0 as an index - starts at 1.
func getRegularInts(nullableInts []sql.NullInt64) []int {
	regularInts := make([]int, len(nullableInts))

	for i, nullInt := range nullableInts {
		var regInt int

		if nullInt.Valid == false {
			regInt = 0
		} else {
			regInt = int(nullInt.Int64)
		}
		regularInts[i] = regInt
	}
	return regularInts
}

func prepareFHIREndpointInfoStatements(s *Store) error {
	var err error
	addFHIREndpointInfoStatement, err = s.DB.Prepare(`
		INSERT INTO fhir_endpoints_info (
			fhir_endpoint_id,
			healthit_product_id,
			tls_version,
			mime_types,
			http_response,
			errors,
			vendor,
			capability_statement,
			validation)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`)
	if err != nil {
		return err
	}
	updateFHIREndpointInfoStatement, err = s.DB.Prepare(`
		UPDATE fhir_endpoints_info
		SET 
		    fhir_endpoint_id = $1,
		    healthit_product_id = $2,
			tls_version = $3,
			mime_types = $4,
			http_response = $5,
			errors = $6,
			vendor = $7,
			capability_statement = $8,
			validation = $9
		WHERE id = $10`)
	if err != nil {
		return err
	}
	deleteFHIREndpointInfoStatement, err = s.DB.Prepare(`
        DELETE FROM fhir_endpoints_info
        WHERE id = $1`)
	if err != nil {
		return err
	}
	return nil
}
