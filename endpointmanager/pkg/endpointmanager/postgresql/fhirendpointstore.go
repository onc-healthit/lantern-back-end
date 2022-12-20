package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/pkg/errors"
)

// prepared statements are left open to be used throughout the execution of the application
// TODO: figure out if there's a better way to manage this for bulk calls
var addFHIREndpointStatement *sql.Stmt
var updateFHIREndpointStatement *sql.Stmt
var deleteFHIREndpointStatement *sql.Stmt
var addFHIREndpointOrganizationStatement *sql.Stmt
var updateFHIREndpointOrganizationStatement *sql.Stmt
var deleteFHIREndpointOrganizationStatement *sql.Stmt
var addFHIREndpointOrganizationMapStatementNoId *sql.Stmt
var addFHIREndpointOrganizationMapStatement *sql.Stmt
var getFHIREndpointOrganizationsByMapID *sql.Stmt
var deleteFHIREndpointOrganizationMapStatement *sql.Stmt

// GetAllFHIREndpoints returns a list of all of the fhir endpoints
func (s *Store) GetAllFHIREndpoints(ctx context.Context) ([]*endpointmanager.FHIREndpoint, error) {
	var versionsResponseJSON []byte

	sqlStatement := `
	SELECT
		id,
		url,
		org_database_map_id,
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
			&endpoint.OrgDatabaseMapID,
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

		organizationsList, err := s.GetFHIREndpointOrganizations(ctx, endpoint.OrgDatabaseMapID)
		if err != nil {
			return nil, err
		}
		endpoint.OrganizationList = organizationsList

		endpoints = append(endpoints, &endpoint)
	}
	return endpoints, nil
}

// GetFHIREndpointOrganizations returns a list of all of the FHIR organizations for the FHIR endpoint
func (s *Store) GetFHIREndpointOrganizations(ctx context.Context, org_map_id int) ([]*endpointmanager.FHIREndpointOrganization, error) {
	var organizationsList []*endpointmanager.FHIREndpointOrganization

	orgRow, err := getFHIREndpointOrganizationsByMapID.QueryContext(ctx, org_map_id)
	if err != nil {
		return nil, err
	}
	defer orgRow.Close()
	for orgRow.Next() {
		var organization endpointmanager.FHIREndpointOrganization
		err = orgRow.Scan(
			&organization.ID,
			&organization.OrganizationName,
			&organization.OrganizationZipCode,
			&organization.OrganizationNPIID)
		if err != nil {
			return nil, err
		}
		organizationsList = append(organizationsList, &organization)
	}
	return organizationsList, nil
}

// GetAllDistinctFHIREndpoints returns a list of all of the fhir endpoints with distinct URLs
func (s *Store) GetAllDistinctFHIREndpoints(ctx context.Context) ([]*endpointmanager.FHIREndpoint, error) {
	sqlStatement := `
	SELECT
		DISTINCT url
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
			&endpoint.URL)
		if err != nil {
			return nil, err
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
		org_database_map_id,
		list_source,
		versions_response,
		created_at,
		updated_at
	FROM fhir_endpoints WHERE id=$1`

	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&endpoint.ID,
		&endpoint.URL,
		&endpoint.OrgDatabaseMapID,
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

	organizationsList, err := s.GetFHIREndpointOrganizations(ctx, endpoint.OrgDatabaseMapID)
	if err != nil {
		return nil, err
	}
	endpoint.OrganizationList = organizationsList

	return &endpoint, err
}

// GetFHIREndpointUsingURL returns all FHIREndpoint from the database using the given url as a key.
func (s *Store) GetFHIREndpointUsingURL(ctx context.Context, url string) ([]*endpointmanager.FHIREndpoint, error) {
	var versionsResponseJSON []byte

	sqlStatement := `
	SELECT
		id,
		url,
		org_database_map_id,
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
			&endpoint.OrgDatabaseMapID,
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
		organizationsList, err := s.GetFHIREndpointOrganizations(ctx, endpoint.OrgDatabaseMapID)
		if err != nil {
			return nil, err
		}
		endpoint.OrganizationList = organizationsList

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
		org_database_map_id,
		list_source,
		versions_response,
		created_at,
		updated_at
	FROM fhir_endpoints
	WHERE url=$1 AND list_source=$2`

	row := s.DB.QueryRowContext(ctx, sqlStatement, url, listSource)

	err := row.Scan(
		&endpoint.ID,
		&endpoint.URL,
		&endpoint.OrgDatabaseMapID,
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

	organizationsList, err := s.GetFHIREndpointOrganizations(ctx, endpoint.OrgDatabaseMapID)
	if err != nil {
		return nil, err
	}
	endpoint.OrganizationList = organizationsList

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
		org_database_map_id,
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
			&endpoint.OrgDatabaseMapID,
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

		organizationsList, err := s.GetFHIREndpointOrganizations(ctx, endpoint.OrgDatabaseMapID)
		if err != nil {
			return nil, err
		}
		endpoint.OrganizationList = organizationsList

		endpoints = append(endpoints, &endpoint)
	}
	return endpoints, nil
}

// UpdateFHIREndpointsNPIOrg updates each endpoint with new organization IDs and names
func (s *Store) UpdateFHIREndpointsNPIOrg(ctx context.Context, e *endpointmanager.FHIREndpoint, add bool) error {
	existingEndpts, err := s.GetFHIREndpointUsingURL(ctx, e.URL)
	if err != nil {
		return errors.Wrap(err, "getting fhir endpoints from store failed")
	} else {
		for _, existingEndpt := range existingEndpts {
			// Merge new data with old data
			// Org names NPI IDs
			if add {
				newOrgs := existingEndpt.OrganizationsToAdd(e.OrganizationList)

				for _, org := range newOrgs {
					databaseMapID, err := s.AddFHIREndpointOrganization(ctx, org, e.OrgDatabaseMapID)
					if err != nil {
						return errors.Wrap(err, "adding fhir endpoint organizations to store failed")
					}
					e.OrgDatabaseMapID = databaseMapID
				}
			} else {
				for _, org := range e.OrganizationList {
					for _, existingOrg := range existingEndpt.OrganizationList {
						if existingOrg.OrganizationNPIID == org.OrganizationNPIID {
							err = s.DeleteFHIREndpointOrganization(ctx, org)
							if err != nil {
								return err
							}
						}
					}
				}
			}
			err = s.UpdateFHIREndpoint(ctx, existingEndpt)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AddOrUpdateFHIREndpoint adds the endpoint if it doesn't already exist. If it does exist, it updates the endpoint.
func (s *Store) AddOrUpdateFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	existingEndpt, err := s.GetFHIREndpointUsingURLAndListSource(ctx, e.URL, e.ListSource)
	if err == sql.ErrNoRows {

		for _, org := range e.OrganizationList {
			databaseMapID, err := s.AddFHIREndpointOrganization(ctx, org, e.OrgDatabaseMapID)
			if err != nil {
				return errors.Wrap(err, "adding fhir endpoint to store failed")
			}
			e.OrgDatabaseMapID = databaseMapID
		}

		err = s.AddFHIREndpoint(ctx, e)
		if err != nil {
			return errors.Wrap(err, "adding fhir endpoint to store failed")
		}
	} else if err != nil {
		return errors.Wrap(err, "getting fhir endpoint from store failed")
	} else {
		// Merge new data with old data
		// Org names NPI IDs Org Zipcodes and VersionsResponse only possible new data
		newOrgs := existingEndpt.OrganizationsToAdd(e.OrganizationList)
		oldOrgs := existingEndpt.OrganizationsToRemove(e.OrganizationList)

		for _, org := range newOrgs {
			databaseMapID, err := s.AddFHIREndpointOrganization(ctx, org, e.OrgDatabaseMapID)
			if err != nil {
				return errors.Wrap(err, "adding fhir endpoint organizations to store failed")
			}
			e.OrgDatabaseMapID = databaseMapID
		}

		for _, org := range oldOrgs {
			err := s.DeleteFHIREndpointOrganization(ctx, org)
			if err != nil {
				return errors.Wrap(err, "removing fhir endpoint organizations from store failed")
			}
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

	for _, org := range e.OrganizationList {
		databaseMapID, err := s.AddFHIREndpointOrganization(ctx, org, e.OrgDatabaseMapID)
		if err != nil {
			return errors.Wrap(err, "adding fhir endpoint to store failed")
		}
		e.OrgDatabaseMapID = databaseMapID
	}

	row := addFHIREndpointStatement.QueryRowContext(ctx,
		e.URL,
		e.OrgDatabaseMapID,
		e.ListSource)

	err = row.Scan(&e.ID)

	return err
}

// AddFHIREndpointOrganization adds the FHIREndpoint Organization to the database.
func (s *Store) AddFHIREndpointOrganization(ctx context.Context, org *endpointmanager.FHIREndpointOrganization, databaseMapID int) (int, error) {
	var err error

	row := addFHIREndpointOrganizationStatement.QueryRowContext(ctx,
		org.OrganizationName,
		org.OrganizationZipCode)

	err = row.Scan(&org.ID)
	if err != nil {
		return 0, err
	}

	orgMapID, err := s.AddFHIREndpointOrganizationMap(ctx, org.ID, databaseMapID)
	if err != nil {
		return 0, err
	}

	return orgMapID, err
}

// AddFHIREndpointOrganizationMap creates a new ID for all the FHIR endpoint organizations for a particular endpoint and returns it
func (s *Store) AddFHIREndpointOrganizationMap(ctx context.Context, id int, OrgDatabaseMapID int) (int, error) {
	var err error
	var organizationMapRow *sql.Row
	if id == 0 {
		organizationMapRow = addFHIREndpointOrganizationMapStatementNoId.QueryRowContext(ctx, OrgDatabaseMapID)
	} else {
		organizationMapRow = addFHIREndpointOrganizationMapStatement.QueryRowContext(ctx, id, OrgDatabaseMapID)
	}
	orgMapID := 0
	err = organizationMapRow.Scan(&orgMapID)

	return orgMapID, err
}

// UpdateFHIREndpoint updates the FHIREndpoint in the database using the FHIREndpoint's database id as the key.
func (s *Store) UpdateFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	var err error
	var versionsResponseJSON []byte

	existingEndpt, err := s.GetFHIREndpointUsingURLAndListSource(ctx, e.URL, e.ListSource)

	if e.VersionsResponse.Response != nil {
		versionsResponseJSON, err = e.VersionsResponse.GetJSON()
		if err != nil {
			return err
		}
	} else {
		versionsResponseJSON = []byte("null")
	}

	newOrgs := existingEndpt.OrganizationsToAdd(e.OrganizationList)
	oldOrgs := existingEndpt.OrganizationsToRemove(e.OrganizationList)

	for _, org := range newOrgs {
		databaseMapID, err := s.AddFHIREndpointOrganization(ctx, org, e.OrgDatabaseMapID)
		if err != nil {
			return errors.Wrap(err, "adding fhir endpoint organizations to store failed")
		}
		e.OrgDatabaseMapID = databaseMapID
	}

	for _, org := range oldOrgs {
		err := s.DeleteFHIREndpointOrganization(ctx, org)
		if err != nil {
			return errors.Wrap(err, "removing fhir endpoint organizations from store failed")
		}
	}

	_, err = updateFHIREndpointStatement.ExecContext(ctx,
		e.URL,
		e.OrgDatabaseMapID,
		e.ListSource,
		versionsResponseJSON,
		e.ID)

	return err
}

// DeleteFHIREndpoint deletes the FHIREndpoint from the database using the FHIREndpoint's database id  as the key.
func (s *Store) DeleteFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {

	_, err := deleteFHIREndpointStatement.ExecContext(ctx, e.ID)

	err = s.DeleteFHIREndpointOrganizationMap(ctx, e)
	return err
}

// DeleteFHIREndpointOrganization deletes the FHIREndpoint Organization from the database using the Organization's database id  as the key.
func (s *Store) DeleteFHIREndpointOrganization(ctx context.Context, o *endpointmanager.FHIREndpointOrganization) error {
	_, err := deleteFHIREndpointOrganizationStatement.ExecContext(ctx, o.ID)

	return err
}

// DeleteFHIREndpointOrganization deletes the FHIREndpoint Organization from the database using the Organization's database id  as the key.
func (s *Store) DeleteFHIREndpointOrganizationMap(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	
	organizationsList, err := s.GetFHIREndpointOrganizations(ctx, e.OrgDatabaseMapID)
	if err != nil {
		return err
	}

	for _, org := range organizationsList {
		err := s.DeleteFHIREndpointOrganization(ctx, org)
		if err != nil {
			return errors.Wrap(err, "removing fhir endpoint organizations from store failed")
		}
	}

	_, err = deleteFHIREndpointOrganizationMapStatement.ExecContext(ctx, e.OrgDatabaseMapID)

	return err
}

func prepareFHIREndpointStatements(s *Store) error {
	var err error
	addFHIREndpointStatement, err = s.DB.Prepare(`
		INSERT INTO fhir_endpoints (url,
			org_database_map_id,
			list_source)
		VALUES ($1, $2, $3)
		RETURNING id`)
	if err != nil {
		return err
	}
	addFHIREndpointOrganizationStatement, err = s.DB.Prepare(`
		INSERT INTO fhir_endpoint_organizations (
			organization_name,
			organization_zipcode)
		VALUES ($1, $2)
		RETURNING id`)
	if err != nil {
		return err
	}
	updateFHIREndpointStatement, err = s.DB.Prepare(`
		UPDATE fhir_endpoints
		SET url = $1,
			org_database_map_id = $2,
			list_source = $3,
			versions_response = $4
		WHERE id = $5`)
	if err != nil {
		return err
	}
	updateFHIREndpointOrganizationStatement, err = s.DB.Prepare(`
	UPDATE fhir_endpoint_organizations
	SET organization_name = $1,
		organization_zipcode = $2
	WHERE id = $3`)
	if err != nil {
		return err
	}
	deleteFHIREndpointStatement, err = s.DB.Prepare(`
        DELETE FROM fhir_endpoints
        WHERE id = $1`)
	if err != nil {
		return err
	}
	deleteFHIREndpointOrganizationStatement, err = s.DB.Prepare(`
	DELETE FROM fhir_endpoint_organizations
	WHERE id = $1`)
	if err != nil {
		return err
	}
	deleteFHIREndpointOrganizationMapStatement, err = s.DB.Prepare(`
	DELETE FROM fhir_endpoint_organizations_map
	WHERE id = $1`)
	if err != nil {
		return err
	}
	addFHIREndpointOrganizationMapStatement, err = s.DB.Prepare(`
	INSERT INTO fhir_endpoint_organizations_map (id, org_database_id)
	VALUES ($1, $2)
	RETURNING id;`)
	if err != nil {
		return err
	}
	addFHIREndpointOrganizationMapStatementNoId, err = s.DB.Prepare(`
	INSERT INTO fhir_endpoint_organizations_map (org_database_id)
	VALUES ($1)
	RETURNING id;`)
	if err != nil {
		return err
	}
	getFHIREndpointOrganizationsByMapID, err = s.DB.Prepare(`
	SELECT org.id, org.organization_name, org.organization_zipcode, org.organization_npi_id
		FROM fhir_endpoint_organizations_map map, fhir_endpoint_organizations org
	WHERE map.id=$1 AND map.org_database_id = org.id;`)
	if err != nil {
		return err
	}
	return nil
}
