package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/pkg/errors"
)

// prepared statements are left open to be used throughout the execution of the application
// TODO: figure out if there's a better way to manage this for bulk calls
var addFHIREndpointStatement *sql.Stmt
var updateFHIREndpointStatement *sql.Stmt
var deleteFHIREndpointStatement *sql.Stmt

// GetAllFHIREndpoints returns a list of all of the fhir endpoints
func (s *Store) GetAllFHIREndpoints(ctx context.Context) ([]*endpointmanager.FHIREndpoint, error) {
	var versionsResponseJSON []byte

	sqlStatement := `
	SELECT
		id,
		url,
		organization_names,
		npi_ids,
		versions_response
	FROM fhir_endpoints`
	rows, err := s.DB.QueryContext(ctx, sqlStatement)
	if err != nil {
		return nil, err
	}

	var endpoints []*endpointmanager.FHIREndpoint
	defer rows.Close()
	for rows.Next() {
		var endpoint endpointmanager.FHIREndpoint
		err = rows.Scan(
			&endpoint.ID,
			&endpoint.URL,
			pq.Array(&endpoint.OrganizationNames),
			pq.Array(&endpoint.NPIIDs),
			&versionsResponseJSON)
		if err != nil {
			return nil, err
		}
		if versionsResponseJSON != nil {
			err = json.Unmarshal(versionsResponseJSON, &endpoint.VersionsResponse)
			if err != nil {
				return nil, errors.Wrap(err, "error unmarshalling JSON versions response")
			}
		}
		endpoints = append(endpoints, &endpoint)
	}
	return endpoints, nil
}

// GetFHIREndpoint gets a FHIREndpoint from the database using the database id as a key.
// If the FHIREndpoint does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetFHIREndpoint(ctx context.Context, id int) (*endpointmanager.FHIREndpoint, error) {
	var endpoint endpointmanager.FHIREndpoint
	var versionsResponseJSON []byte

	sqlStatement := `
	SELECT
		id,
		url,
		organization_names,
		npi_ids,
		list_source,
		versions_response,
		created_at,
		updated_at
	FROM fhir_endpoints WHERE id=$1`
	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&endpoint.ID,
		&endpoint.URL,
		pq.Array(&endpoint.OrganizationNames),
		pq.Array(&endpoint.NPIIDs),
		&endpoint.ListSource,
		&versionsResponseJSON,
		&endpoint.CreatedAt,
		&endpoint.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if versionsResponseJSON != nil {
		err = json.Unmarshal(versionsResponseJSON, &endpoint.VersionsResponse)
		if err != nil {
			return nil, errors.Wrap(err, "error unmarshalling JSON versions response")
		}
	}

	return &endpoint, err
}

// GetFHIREndpointUsingURL returns all FHIREndpoint from the database using the given url as a key.
func (s *Store) GetFHIREndpointUsingURL(ctx context.Context, url string) ([]*endpointmanager.FHIREndpoint, error) {
	var versionsResponseJSON []byte

	sqlStatement := `
	SELECT
		id,
		url,
		organization_names,
		npi_ids,
		list_source,
		versions_response
	FROM fhir_endpoints WHERE url=$1`
	rows, err := s.DB.QueryContext(ctx, sqlStatement, url)
	if err != nil {
		return nil, err
	}

	var endpoints []*endpointmanager.FHIREndpoint
	defer rows.Close()
	for rows.Next() {
		var endpoint endpointmanager.FHIREndpoint
		err = rows.Scan(
			&endpoint.ID,
			&endpoint.URL,
			pq.Array(&endpoint.OrganizationNames),
			pq.Array(&endpoint.NPIIDs),
			&endpoint.ListSource,
			&versionsResponseJSON)
		if err != nil {
			return nil, err
		}
		if versionsResponseJSON != nil {
			err = json.Unmarshal(versionsResponseJSON, &endpoint.VersionsResponse)
			if err != nil {
				return nil, errors.Wrap(err, "error unmarshalling JSON versions response")
			}
		}
		endpoints = append(endpoints, &endpoint)
	}
	return endpoints, nil
}

// GetFHIREndpointUsingURLAndListSource gets a FHIREndpoint from the database using the given url as a key.
// If the FHIREndpoint does not exist in the database, sql.ErrNoRows will be returned.
func (s *Store) GetFHIREndpointUsingURLAndListSource(ctx context.Context, url string, listSource string) (*endpointmanager.FHIREndpoint, error) {
	var endpoint endpointmanager.FHIREndpoint
	var versionsResponseJSON []byte

	sqlStatement := `
	SELECT
		id,
		url,
		organization_names,
		npi_ids,
		list_source,
		versions_response,
		created_at,
		updated_at
	FROM fhir_endpoints WHERE url=$1 AND list_source=$2`

	row := s.DB.QueryRowContext(ctx, sqlStatement, url, listSource)

	err := row.Scan(
		&endpoint.ID,
		&endpoint.URL,
		pq.Array(&endpoint.OrganizationNames),
		pq.Array(&endpoint.NPIIDs),
		&endpoint.ListSource,
		&versionsResponseJSON,
		&endpoint.CreatedAt,
		&endpoint.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if versionsResponseJSON != nil {
		err = json.Unmarshal(versionsResponseJSON, &endpoint.VersionsResponse)
		if err != nil {
			return nil, errors.Wrap(err, "error unmarshalling JSON versions response")
		}
	}

	return &endpoint, err
}

// GetFHIREndpointsUsingListSourceAndUpdateTime retrieves all fhir endpoints from the database from the given
// listsource that update time is before the given update time.
func (s *Store) GetFHIREndpointsUsingListSourceAndUpdateTime(ctx context.Context, updateTime time.Time, listSource string) ([]*endpointmanager.FHIREndpoint, error) {
	var versionsResponseJSON []byte

	sqlStatement := `
	SELECT
		id,
		url,
		organization_names,
		npi_ids,
		versions_response
	FROM fhir_endpoints WHERE list_source=$1 AND updated_at<$2`

	rows, err := s.DB.QueryContext(ctx, sqlStatement, listSource, updateTime)
	if err != nil {
		return nil, err
	}

	var endpoints []*endpointmanager.FHIREndpoint
	defer rows.Close()
	for rows.Next() {
		var endpoint endpointmanager.FHIREndpoint
		err = rows.Scan(
			&endpoint.ID,
			&endpoint.URL,
			pq.Array(&endpoint.OrganizationNames),
			pq.Array(&endpoint.NPIIDs),
			&versionsResponseJSON)
		if err != nil {
			return nil, err
		}
		if versionsResponseJSON != nil {
			err = json.Unmarshal(versionsResponseJSON, &endpoint.VersionsResponse)
			if err != nil {
				return nil, errors.Wrap(err, "error unmarshalling JSON versions response")
			}
		}
		endpoints = append(endpoints, &endpoint)
	}
	return endpoints, nil
}

// AddOrUpdateFHIREndpoint adds the endpoint if it doesn't already exist. If it does exist, it updates the endpoint.
func (s *Store) AddOrUpdateFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	existingEndpt, err := s.GetFHIREndpointUsingURLAndListSource(ctx, e.URL, e.ListSource)
	if err == sql.ErrNoRows {
		err = s.AddFHIREndpoint(ctx, e)
		if err != nil {
			return errors.Wrap(err, "adding fhir endpoint to store failed")
		}
	} else if err != nil {
		return errors.Wrap(err, "getting fhir endpoint from store failed")
	} else {
		// Merge new data with old data
		// Org names and NPI IDs only possible new data
		for _, name := range e.OrganizationNames {
			existingEndpt.AddOrganizationName(name)
		}
		for _, npiID := range e.NPIIDs {
			existingEndpt.AddNPIID(npiID)
		}
		existingEndpt.VersionsResponse = e.VersionsResponse
		err = s.UpdateFHIREndpoint(ctx, existingEndpt)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddFHIREndpoint adds the FHIREndpoint to the database.
func (s *Store) AddFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	var err error

	row := addFHIREndpointStatement.QueryRowContext(ctx,
		e.URL,
		pq.Array(e.OrganizationNames),
		pq.Array(e.NPIIDs),
		e.ListSource)

	err = row.Scan(&e.ID)

	return err
}

// UpdateFHIREndpoint updates the FHIREndpoint in the database using the FHIREndpoint's database id as the key.
func (s *Store) UpdateFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	var err error
	var versionsResponseJSON []byte

	if e.VersionsResponse.Response != nil {
		versionsResponseJSON, err = e.VersionsResponse.GetJSON()
		if err != nil {
			return err
		}
	} else {
		versionsResponseJSON = []byte("null")
	}

	_, err = updateFHIREndpointStatement.ExecContext(ctx,
		e.URL,
		pq.Array(e.OrganizationNames),
		pq.Array(e.NPIIDs),
		e.ListSource,
		versionsResponseJSON,
		e.ID)

	return err
}

// DeleteFHIREndpoint deletes the FHIREndpoint from the database using the FHIREndpoint's database id  as the key.
func (s *Store) DeleteFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	_, err := deleteFHIREndpointStatement.ExecContext(ctx, e.ID)

	return err
}

func prepareFHIREndpointStatements(s *Store) error {
	var err error
	addFHIREndpointStatement, err = s.DB.Prepare(`
		INSERT INTO fhir_endpoints (url,
			organization_names,
			npi_ids,
			list_source)
		VALUES ($1, $2, $3, $4)
		RETURNING id`)
	if err != nil {
		return err
	}
	updateFHIREndpointStatement, err = s.DB.Prepare(`
		UPDATE fhir_endpoints
		SET url = $1,
			organization_names = $2,
			npi_ids = $3,
			list_source = $4,
			versions_response = $5
		WHERE id = $6`)
	if err != nil {
		return err
	}
	deleteFHIREndpointStatement, err = s.DB.Prepare(`
        DELETE FROM fhir_endpoints
        WHERE id = $1`)
	if err != nil {
		return err
	}
	return nil
}
