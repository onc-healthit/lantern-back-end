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
var addFHIREndpointStatement *sql.Stmt
var updateFHIREndpointStatement *sql.Stmt
var deleteFHIREndpointStatement *sql.Stmt
var addFHIREndpointOrganizationStatement *sql.Stmt
var deleteFHIREndpointOrganizationStatement *sql.Stmt
var addFHIREndpointOrganizationMapStatement *sql.Stmt
var getFHIREndpointOrganizationsByEndpointID *sql.Stmt
var getFHIREndpointOrganizationByInfoStatement *sql.Stmt
var deleteFHIREndpointOrganizationMapStatement *sql.Stmt
var deleteFHIREndpointOrganizationMapStatementConditional *sql.Stmt
var updateFHIREndpointOrganizationsUpdateTime *sql.Stmt
var updateProcessCompletionStatusStatement *sql.Stmt
var addFHIREndpointOrganizationIdentifierStatement *sql.Stmt
var addFHIREndpointOrganizationAddressStatement *sql.Stmt
var deleteFHIREndpointOrganizationIdentifierStatement *sql.Stmt
var deleteFHIREndpointOrganizationAddressStatement *sql.Stmt
var addFHIREndpointOrganizationActiveStatement *sql.Stmt
var deleteFHIREndpointOrganizationActiveStatement *sql.Stmt
var addFHIREndpointOrganizationURLStatement *sql.Stmt
var deleteFHIREndpointOrganizationURLStatement *sql.Stmt
var deleteFHIREndpointOrganizationMapByPairStatement *sql.Stmt

// GetAllFHIREndpoints returns a list of all of the fhir endpoints
func (s *Store) GetAllFHIREndpoints(ctx context.Context) ([]*endpointmanager.FHIREndpoint, error) {
	var versionsResponseJSON []byte

	sqlStatement := `
	SELECT
		id,
		url,
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

		organizationsList, err := s.GetFHIREndpointOrganizations(ctx, endpoint.ID)
		if err != nil {
			return nil, err
		}
		endpoint.OrganizationList = organizationsList

		endpoints = append(endpoints, &endpoint)
	}
	return endpoints, nil
}

// GetFHIREndpointOrganizations returns a list of all of the FHIR organizations for the FHIR endpoint
func (s *Store) GetFHIREndpointOrganizations(ctx context.Context, endpoint_id int) ([]*endpointmanager.FHIREndpointOrganization, error) {
	var organizationsList []*endpointmanager.FHIREndpointOrganization
	var organizationName sql.NullString
	var organizationNPIID sql.NullString
	var organizationZipCode sql.NullString

	orgRow, err := getFHIREndpointOrganizationsByEndpointID.QueryContext(ctx, endpoint_id)
	if err != nil {
		return nil, err
	}
	defer orgRow.Close()
	for orgRow.Next() {
		var organization endpointmanager.FHIREndpointOrganization
		err = orgRow.Scan(
			&organization.ID,
			&organizationName,
			&organizationZipCode,
			&organizationNPIID,
			&organization.UpdatedAt)
		if err != nil {
			return nil, err
		}

		orgName, orgZipCode, orgNPIID := organizationInformationValid(organizationName, organizationZipCode, organizationNPIID)
		organization.OrganizationName = orgName
		organization.OrganizationZipCode = orgZipCode
		organization.OrganizationNPIID = orgNPIID

		organizationsList = append(organizationsList, &organization)
	}
	return organizationsList, nil
}

// GetFHIREndpointOrganizationByInfo returns an organization for the FHIR endpoint matching the specific organization information
func (s *Store) GetFHIREndpointOrganizationByInfo(ctx context.Context, endpoint_id int, org *endpointmanager.FHIREndpointOrganization) (*endpointmanager.FHIREndpointOrganization, error) {
	var organizationName sql.NullString
	var organizationNPIID sql.NullString
	var organizationZipCode sql.NullString

	orgRow := getFHIREndpointOrganizationByInfoStatement.QueryRowContext(ctx, endpoint_id, org.OrganizationName, org.OrganizationZipCode, org.OrganizationNPIID)

	var organization endpointmanager.FHIREndpointOrganization
	err := orgRow.Scan(
		&organization.ID,
		&organizationName,
		&organizationZipCode,
		&organizationNPIID,
		&organization.UpdatedAt)
	if err != nil {
		return nil, err
	}

	orgName, orgZipCode, orgNPIID := organizationInformationValid(organizationName, organizationZipCode, organizationNPIID)
	organization.OrganizationName = orgName
	organization.OrganizationZipCode = orgZipCode
	organization.OrganizationNPIID = orgNPIID

	return &organization, nil
}

// GetFHIREndpointOrganizationByURLandListSource returns an organization for the FHIR endpoint with the given list source and url
func (s *Store) GetFHIREndpointOrganizationByURLandListSource(ctx context.Context, url string, listSource string) (*endpointmanager.FHIREndpointOrganization, error) {
	var organizationName sql.NullString
	var organizationNPIID sql.NullString
	var organizationZipCode sql.NullString

	sqlStatement := `
	SELECT o.id, o.organization_name, o.organization_zipcode, 
	o.organization_npi_id, o.updated_at
	FROM fhir_endpoints e, fhir_endpoint_organizations_map m, fhir_endpoint_organizations o
	WHERE e.id = m.id AND m.org_database_id = o.id 
	AND e.list_source=$1 AND e.url=$2 ORDER BY updated_at DESC;`

	orgRow := s.DB.QueryRowContext(ctx, sqlStatement, listSource, url)

	var organization endpointmanager.FHIREndpointOrganization
	err := orgRow.Scan(
		&organization.ID,
		&organizationName,
		&organizationZipCode,
		&organizationNPIID,
		&organization.UpdatedAt)
	if err != nil {
		return nil, err
	}

	orgName, orgZipCode, orgNPIID := organizationInformationValid(organizationName, organizationZipCode, organizationNPIID)
	organization.OrganizationName = orgName
	organization.OrganizationZipCode = orgZipCode
	organization.OrganizationNPIID = orgNPIID

	return &organization, nil
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
		list_source,
		versions_response,
		created_at,
		updated_at
	FROM fhir_endpoints WHERE id=$1`

	row := s.DB.QueryRowContext(ctx, sqlStatement, id)

	err := row.Scan(
		&endpoint.ID,
		&endpoint.URL,
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

	organizationsList, err := s.GetFHIREndpointOrganizations(ctx, endpoint.ID)
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
		organizationsList, err := s.GetFHIREndpointOrganizations(ctx, endpoint.ID)
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

	organizationsList, err := s.GetFHIREndpointOrganizations(ctx, endpoint.ID)
	if err != nil {
		return nil, err
	}
	endpoint.OrganizationList = organizationsList

	return &endpoint, err
}

// GetFHIREndpointsByListSourceAndOrganizationsUpdatedAtTime returns a list of all of the FHIR endpoints organizations for the given list source that have an update time before the given update time.
func (s *Store) GetFHIREndpointsByListSourceAndOrganizationsUpdatedAtTime(ctx context.Context, updateTime time.Time, listSource string) ([]*endpointmanager.FHIREndpoint, error) {
	var endpointList []*endpointmanager.FHIREndpoint
	var organizationName sql.NullString
	var organizationNPIID sql.NullString
	var organizationZipCode sql.NullString

	sqlStatement := `
	SELECT
		e.id,
		o.id,
		o.organization_name,
		o.organization_zipcode,
		o.organization_npi_id,
		o.updated_at
	FROM fhir_endpoints e, fhir_endpoint_organizations_map m, fhir_endpoint_organizations o 
	WHERE e.id = m.id AND m.org_database_id = o.id 
	AND e.list_source=$1 AND o.updated_at<$2`

	orgRow, err := s.DB.QueryContext(ctx, sqlStatement, listSource, updateTime)
	if err != nil {
		return nil, err
	}
	defer orgRow.Close()
	for orgRow.Next() {
		var organization endpointmanager.FHIREndpointOrganization
		var endpoint endpointmanager.FHIREndpoint

		err = orgRow.Scan(
			&endpoint.ID,
			&organization.ID,
			&organizationName,
			&organizationZipCode,
			&organizationNPIID,
			&organization.UpdatedAt)
		if err != nil {
			return nil, err
		}

		orgName, orgZipCode, orgNPIID := organizationInformationValid(organizationName, organizationZipCode, organizationNPIID)
		organization.OrganizationName = orgName
		organization.OrganizationZipCode = orgZipCode
		organization.OrganizationNPIID = orgNPIID

		endpoint.OrganizationList = []*endpointmanager.FHIREndpointOrganization{&organization}
		endpointList = append(endpointList, &endpoint)
	}
	return endpointList, nil
}

// GetFHIREndpointsUsingListSourceAndUpdateTime retrieves all fhir endpoints from the database from the given
// listsource that update time is before the given update time.
func (s *Store) GetFHIREndpointsUsingListSourceAndUpdateTime(ctx context.Context, updateTime time.Time, listSource string) ([]*endpointmanager.FHIREndpoint, error) {
	var versionsResponseJSON []byte

	sqlStatement := `
	SELECT
		id,
		url,
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

		organizationsList, err := s.GetFHIREndpointOrganizations(ctx, endpoint.ID)
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
				existingEndpt.OrganizationList = e.OrganizationList
				err = s.UpdateFHIREndpoint(ctx, existingEndpt)
				if err != nil {
					return err
				}
			} else {
				for _, org := range e.OrganizationList {
					for _, existingOrg := range existingEndpt.OrganizationList {
						if existingOrg.OrganizationNPIID == org.OrganizationNPIID {
							err = s.DeleteFHIREndpointOrganization(ctx, existingOrg, existingEndpt.ID)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}
	return nil
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
		// Org names NPI IDs Org Zipcodes and VersionsResponse only possible new data
		existingEndpt.VersionsResponse = e.VersionsResponse
		existingEndpt.OrganizationList = e.OrganizationList
		err = s.UpdateFHIREndpoint(ctx, existingEndpt)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateFHIREndpointOrganizations updates the FHIREndpoint's list of organizations
func (s *Store) UpdateFHIREndpointOrganizations(ctx context.Context, e *endpointmanager.FHIREndpoint) error {

	for _, org := range e.OrganizationList {
		organization, err := s.GetFHIREndpointOrganizationByInfo(ctx, e.ID, org)

		// If the organization does not exist, add it to the database, otherwise update the updated time
		if err == sql.ErrNoRows {
			err := s.AddFHIREndpointOrganization(ctx, org, e.ID)
			if err != nil {
				return errors.Wrap(err, "adding fhir endpoint to store failed")
			}
			e.OrganizationList = append(e.OrganizationList, org)
		} else if err != nil {
			return errors.Wrap(err, "getting fhir endpoint organization from store failed")
		} else {

			_, err := updateFHIREndpointOrganizationsUpdateTime.ExecContext(ctx, organization.ID)

			if err != nil {
				return errors.Wrap(err, "updating the fhir endpoint's organization update time failed")
			}

			err = s.AddFHIREndpointOrganizationIdentifiers(ctx, organization.ID, org.OrganizationIdentifiers)
			if err != nil {
				return err
			}

			err = s.AddFHIREndpointOrganizationAddresses(ctx, organization.ID, org.OrganizationAddresses)
			if err != nil {
				return err
			}

			err = s.AddFHIREndpointOrganizationActive(ctx, organization.ID, org.OrganizationActive)
			if err != nil {
				return err
			}

			err = s.AddFHIREndpointOrganizationURL(ctx, organization.ID, org.OrganizationURL)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AddFHIREndpoint adds the FHIREndpoint to the database.
func (s *Store) AddFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
	var err error

	row := addFHIREndpointStatement.QueryRowContext(ctx,
		e.URL,
		e.ListSource)

	err = row.Scan(&e.ID)

	for _, org := range e.OrganizationList {
		err := s.AddFHIREndpointOrganization(ctx, org, e.ID)
		if err != nil {
			return errors.Wrap(err, "adding fhir endpoint to store failed")
		}
	}

	return err
}

// AddFHIREndpointOrganization adds the FHIREndpoint Organization to the database.
func (s *Store) AddFHIREndpointOrganization(ctx context.Context, org *endpointmanager.FHIREndpointOrganization, endpointID int) error {
	var err error

	row := addFHIREndpointOrganizationStatement.QueryRowContext(ctx,
		org.OrganizationName,
		org.OrganizationNPIID,
		org.OrganizationZipCode)

	err = row.Scan(&org.ID)
	if err != nil {
		return err
	}

	err = s.AddFHIREndpointOrganizationMap(ctx, org.ID, endpointID)
	if err != nil {
		return err
	}

	err = s.AddFHIREndpointOrganizationIdentifiers(ctx, org.ID, org.OrganizationIdentifiers)
	if err != nil {
		return err
	}

	err = s.AddFHIREndpointOrganizationAddresses(ctx, org.ID, org.OrganizationAddresses)
	if err != nil {
		return err
	}

	err = s.AddFHIREndpointOrganizationURL(ctx, org.ID, org.OrganizationURL)
	if err != nil {
		return err
	}

	err = s.AddFHIREndpointOrganizationActive(ctx, org.ID, org.OrganizationActive)

	return err
}

// AddFHIREndpointOrganizationMap creates a new ID for all the FHIR endpoint organizations for a particular endpoint and returns it
func (s *Store) AddFHIREndpointOrganizationMap(ctx context.Context, orgID int, endpointID int) error {
	var err error

	_, err = addFHIREndpointOrganizationMapStatement.ExecContext(ctx, endpointID, orgID)

	return err
}

func (s *Store) AddFHIREndpointOrganizationIdentifiers(ctx context.Context, orgID int, orgIdentifiers []interface{}) error {

	var err error
	var count int

	row := s.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoint_organization_identifiers WHERE org_id=$1;", orgID)

	err = row.Scan(&count)
	if err != nil {
		return err
	}

	// If there are entries in the fhir_endpoint_organization_identifiers table that has this orgID, delete those first
	if count > 0 {
		_, err = deleteFHIREndpointOrganizationIdentifierStatement.ExecContext(ctx, orgID)
		if err != nil {
			return err
		}
	}

	for _, identifier := range orgIdentifiers {
		_, err = addFHIREndpointOrganizationIdentifierStatement.ExecContext(ctx, orgID, identifier.(string))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) AddFHIREndpointOrganizationAddresses(ctx context.Context, orgID int, orgAddresses []interface{}) error {
	var err error
	var count int

	row := s.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoint_organization_addresses WHERE org_id=$1;", orgID)

	err = row.Scan(&count)
	if err != nil {
		return err
	}

	// If there are entries in the fhir_endpoint_organization_addresses table that has this orgID, delete those first
	if count > 0 {
		_, err = deleteFHIREndpointOrganizationAddressStatement.ExecContext(ctx, orgID)
		if err != nil {
			return err
		}
	}

	for _, address := range orgAddresses {
		_, err = addFHIREndpointOrganizationAddressStatement.ExecContext(ctx, orgID, address.(string))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) AddFHIREndpointOrganizationActive(ctx context.Context, orgID int, orgActive string) error {
	var err error
	var count int

	row := s.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoint_organization_active WHERE org_id=$1;", orgID)

	err = row.Scan(&count)
	if err != nil {
		return err
	}

	// If there are entries in the fhir_endpoint_organization_active table that has this orgID, delete those first
	if count > 0 {
		_, err = deleteFHIREndpointOrganizationActiveStatement.ExecContext(ctx, orgID)
		if err != nil {
			return err
		}
	}

	// Only insert organization active data if it was provided in the FHIR bundle
	if orgActive != "" {
		_, err = addFHIREndpointOrganizationActiveStatement.ExecContext(ctx, orgID, orgActive)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) AddFHIREndpointOrganizationURL(ctx context.Context, orgID int, orgURL string) error {
	var err error
	var count int

	row := s.DB.QueryRow("SELECT COUNT(*) FROM fhir_endpoint_organization_url WHERE org_id=$1;", orgID)

	err = row.Scan(&count)
	if err != nil {
		return err
	}

	// If there are entries in the fhir_endpoint_organization_url table that has this orgID, delete those first
	if count > 0 {
		_, err = deleteFHIREndpointOrganizationURLStatement.ExecContext(ctx, orgID)
		if err != nil {
			return err
		}
	}

	// Only insert organization URL data if it was provided in the FHIR bundle
	if orgURL != "" {
		_, err = addFHIREndpointOrganizationURLStatement.ExecContext(ctx, orgID, orgURL)
		if err != nil {
			return err
		}
	}

	return nil
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
		e.ListSource,
		versionsResponseJSON,
		e.ID)

	if err != nil {
		return err
	}

	err = s.UpdateFHIREndpointOrganizations(ctx, e)
	if err != nil {
		return err
	}

	return err
}

// DeleteFHIREndpoint deletes the FHIREndpoint from the database using the FHIREndpoint's database id  as the key.
func (s *Store) DeleteFHIREndpoint(ctx context.Context, e *endpointmanager.FHIREndpoint) error {

	_, err := deleteFHIREndpointStatement.ExecContext(ctx, e.ID)
	if err != nil {
		return err
	}

	return err
}

// DeleteFHIREndpointOrganization deletes one organization and all related rows for a given endpoint.
func (s *Store) DeleteFHIREndpointOrganization(ctx context.Context, o *endpointmanager.FHIREndpointOrganization, endpointID int) error {
	// 1) Delete Children First
	if _, err := deleteFHIREndpointOrganizationIdentifierStatement.ExecContext(ctx, o.ID); err != nil {
		return err
	}
	if _, err := deleteFHIREndpointOrganizationAddressStatement.ExecContext(ctx, o.ID); err != nil {
		return err
	}
	if _, err := deleteFHIREndpointOrganizationActiveStatement.ExecContext(ctx, o.ID); err != nil {
		return err
	}
	if _, err := deleteFHIREndpointOrganizationURLStatement.ExecContext(ctx, o.ID); err != nil {
		return err
	}

	// 2) Delete Mapping Row for (endpointID, orgID)
	if _, err := deleteFHIREndpointOrganizationMapByPairStatement.ExecContext(ctx, endpointID, o.ID); err != nil {
		return err
	}

	// 3) Delete Parent Org
	if _, err := deleteFHIREndpointOrganizationStatement.ExecContext(ctx, o.ID); err != nil {
		return err
	}

	return nil
}

// DeleteFHIREndpointOrganizationMap removes the given endpoint's organizations in e.OrganizationList.
func (s *Store) DeleteFHIREndpointOrganizationMap(ctx context.Context, e *endpointmanager.FHIREndpoint) error {

	for _, org := range e.OrganizationList {
		if err := s.DeleteFHIREndpointOrganization(ctx, org, e.ID); err != nil {
			return errors.Wrap(err, "removing fhir endpoint organization failed")
		}
	}

	return nil
}

func organizationInformationValid(organizationName sql.NullString, organizationZipCode sql.NullString, organizationNPIID sql.NullString) (string, string, string) {
	var organizationNameString string
	var organizationZipCodeString string
	var organizationNPIIDString string

	if !organizationName.Valid {
		organizationNameString = ""
	} else {
		organizationNameString = organizationName.String
	}

	if !organizationZipCode.Valid {
		organizationZipCodeString = ""
	} else {
		organizationZipCodeString = organizationZipCode.String
	}

	if !organizationNPIID.Valid {
		organizationNPIIDString = ""
	} else {
		organizationNPIIDString = organizationNPIID.String
	}

	return organizationNameString, organizationZipCodeString, organizationNPIIDString
}

// UpdateProcessCompletionStatus updates the completion status of the daily querying process.
func (s *Store) UpdateProcessCompletionStatus(ctx context.Context, status string) error {
	var err error

	_, err = updateProcessCompletionStatusStatement.ExecContext(ctx, status)

	if err != nil {
		return err
	}

	return nil
}

func prepareFHIREndpointStatements(s *Store) error {
	var err error
	addFHIREndpointStatement, err = s.DB.Prepare(`
		INSERT INTO fhir_endpoints (url,
			list_source)
		VALUES ($1, $2)
		RETURNING id`)
	if err != nil {
		return err
	}
	addFHIREndpointOrganizationStatement, err = s.DB.Prepare(`
		INSERT INTO fhir_endpoint_organizations (
			organization_name,
			organization_npi_id,
			organization_zipcode)
		VALUES ($1, $2, $3)
		RETURNING id`)
	if err != nil {
		return err
	}
	updateFHIREndpointStatement, err = s.DB.Prepare(`
		UPDATE fhir_endpoints
		SET url = $1,
			list_source = $2,
			versions_response = $3
		WHERE id = $4`)
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
	deleteFHIREndpointOrganizationMapStatementConditional, err = s.DB.Prepare(`
	DELETE FROM fhir_endpoint_organizations_map
	WHERE id = $1 AND org_database_id IS NULL`)
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
	getFHIREndpointOrganizationsByEndpointID, err = s.DB.Prepare(`
	SELECT org.id, org.organization_name, org.organization_zipcode, org.organization_npi_id, org.updated_at
		FROM fhir_endpoint_organizations_map map, fhir_endpoint_organizations org
	WHERE map.id=$1 AND map.org_database_id = org.id;`)
	if err != nil {
		return err
	}
	getFHIREndpointOrganizationByInfoStatement, err = s.DB.Prepare(`
	SELECT org.id, org.organization_name, org.organization_zipcode, org.organization_npi_id, org.updated_at
		FROM fhir_endpoint_organizations_map map, fhir_endpoint_organizations org
	WHERE map.id=$1 AND map.org_database_id = org.id 
	AND organization_name = $2 AND organization_zipcode = $3 AND organization_npi_id = $4;`)
	if err != nil {
		return err
	}
	updateFHIREndpointOrganizationsUpdateTime, err = s.DB.Prepare(`
	UPDATE fhir_endpoint_organizations SET updated_at = now() WHERE id = $1;`)
	if err != nil {
		return err
	}
	addFHIREndpointOrganizationIdentifierStatement, err = s.DB.Prepare(`
	INSERT INTO fhir_endpoint_organization_identifiers (org_id, identifier)
	VALUES ($1, $2)
	RETURNING org_id;`)
	if err != nil {
		return err
	}
	addFHIREndpointOrganizationAddressStatement, err = s.DB.Prepare(`
	INSERT INTO fhir_endpoint_organization_addresses (org_id, address)
	VALUES ($1, $2)
	RETURNING org_id;`)
	if err != nil {
		return err
	}
	addFHIREndpointOrganizationActiveStatement, err = s.DB.Prepare(`
	INSERT INTO fhir_endpoint_organization_active (org_id, active)
	VALUES ($1, $2)
	RETURNING org_id;`)
	if err != nil {
		return err
	}
	addFHIREndpointOrganizationURLStatement, err = s.DB.Prepare(`
	INSERT INTO fhir_endpoint_organization_url (org_id, org_url)
	VALUES ($1, $2)
	RETURNING org_id;`)
	if err != nil {
		return err
	}
	deleteFHIREndpointOrganizationIdentifierStatement, err = s.DB.Prepare(`
	DELETE FROM fhir_endpoint_organization_identifiers
	WHERE org_id = $1`)
	if err != nil {
		return err
	}
	deleteFHIREndpointOrganizationAddressStatement, err = s.DB.Prepare(`
	DELETE FROM fhir_endpoint_organization_addresses
	WHERE org_id = $1`)
	if err != nil {
		return err
	}
	deleteFHIREndpointOrganizationActiveStatement, err = s.DB.Prepare(`
	DELETE FROM fhir_endpoint_organization_active
	WHERE org_id = $1`)
	if err != nil {
		return err
	}
	deleteFHIREndpointOrganizationURLStatement, err = s.DB.Prepare(`
	DELETE FROM fhir_endpoint_organization_url
	WHERE org_id = $1`)
	if err != nil {
		return err
	}
	deleteFHIREndpointOrganizationMapByPairStatement, err = s.DB.Prepare(`
    DELETE FROM fhir_endpoint_organizations_map
    WHERE id = $1 AND org_database_id = $2`)
	if err != nil {
		return err
	}
	updateProcessCompletionStatusStatement, err = s.DB.Prepare(`
	UPDATE daily_querying_status SET status = $1`)
	if err != nil {
		return err
	}
	return nil
}
