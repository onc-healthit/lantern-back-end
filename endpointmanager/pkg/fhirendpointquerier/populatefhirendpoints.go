package populatefhirendpoints

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/fetcher"

	"regexp"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// AddEndpointData iterates through the list of endpoints and adds each one to the database
func AddEndpointData(ctx context.Context, store *postgresql.Store, endpoints *fetcher.ListOfEndpoints) error {
	var firstUpdate time.Time
	var firstUpdateOrg time.Time
	var listsource = endpoints.Entries[0].ListSource
	for i, endpoint := range endpoints.Entries {
		select {
		case <-ctx.Done():
			return errors.Wrapf(ctx.Err(), "saved %d out of %d endpoints before context ended", i, len(endpoints.Entries))
		default:
			// ok
		}

		// Add trailing "/" to URIs that do not have it for consistency
		uri := endpoint.FHIRPatientFacingURI
		if len(uri) > 0 && uri[len(uri)-1:] != "/" {
			uri = uri + "/"
		}

		splitEndpoint := strings.Split(uri, "://")
		header := "http://"

		if len(splitEndpoint) > 1 {
			header = strings.ToLower(splitEndpoint[0]) + "://"
		}
		uri = header + splitEndpoint[len(splitEndpoint)-1]
		endpoint.FHIRPatientFacingURI = uri

		if isValidURL(uri) {
			err := saveEndpointData(ctx, store, &endpoint)
			if err != nil {
				log.Warn(err)
				continue
			}
			if firstUpdate.IsZero() {
				// get time of update for first endpoint
				fhirURL := endpoint.FHIRPatientFacingURI
				if fhirURL[len(fhirURL)-1:] != "/" {
					fhirURL = fhirURL + "/"
				}

				splitEndpoint := strings.Split(fhirURL, "://")
				header := "http://"

				if len(splitEndpoint) > 1 {
					header = strings.ToLower(splitEndpoint[0]) + "://"
				}
				fhirURL = header + splitEndpoint[len(splitEndpoint)-1]

				existingEndpt, err := store.GetFHIREndpointUsingURLAndListSource(ctx, fhirURL, endpoint.ListSource)
				if err != nil {
					log.Warn(err)
					continue
				} else {
					firstUpdate = existingEndpt.UpdatedAt
				}
			}
			if firstUpdateOrg.IsZero() {
				// get time of update for first endpoint organization
				fhirURL := endpoint.FHIRPatientFacingURI
				if fhirURL[len(fhirURL)-1:] != "/" {
					fhirURL = fhirURL + "/"
				}

				existingOrg, err := store.GetFHIREndpointOrganizationByURLandListSource(ctx, fhirURL, endpoint.ListSource)
				if err == sql.ErrNoRows {
					continue
				} else if err != nil {
					log.Warn(err)
					continue
				} else {
					firstUpdateOrg = existingOrg.UpdatedAt
				}
			}
		}
	}

	err := RemoveOldEndpointOrganizations(ctx, store, firstUpdateOrg, listsource)
	if err != nil {
		log.Warn(err)
	}

	err = RemoveOldEndpoints(ctx, store, firstUpdate, listsource)
	if err != nil {
		log.Warn(err)
	}

	return nil
}

// saveEndpointData formats the endpoint as a FHIREndpoint and then checks to see if it's in the database.
// If it is, ignore it, if it isn't, add it to the database.
func saveEndpointData(ctx context.Context, store *postgresql.Store, endpoint *fetcher.EndpointEntry) error {
	fhirEndpoint, err := formatToFHIREndpt(endpoint)
	if err != nil {
		return err
	}

	err = store.AddOrUpdateFHIREndpoint(ctx, fhirEndpoint)

	return err
}

// formatToFHIREndpt takes an entry in the list of endpoints and formats it for the fhir_endpoints table in the database
func formatToFHIREndpt(endpoint *fetcher.EndpointEntry) (*endpointmanager.FHIREndpoint, error) {
	// Add trailing "/" to URIs that do not have it for consistency
	uri := endpoint.FHIRPatientFacingURI
	if len(uri) > 0 && uri[len(uri)-1:] != "/" {
		uri = uri + "/"
	}

	splitEndpoint := strings.Split(uri, "://")
	header := "http://"

	if len(splitEndpoint) > 1 {
		header = strings.ToLower(splitEndpoint[0]) + "://"
	}
	uri = header + splitEndpoint[len(splitEndpoint)-1]

	// convert the endpoint entry to the fhirDatabase format
	dbEntry := endpointmanager.FHIREndpoint{
		URL:        uri,
		ListSource: endpoint.ListSource,
	}

	if endpoint.OrganizationName != "" || endpoint.NPIID != "" || endpoint.OrganizationZipCode != "" {
		dbOrgEntry := endpointmanager.FHIREndpointOrganization{
			OrganizationName:        endpoint.OrganizationName,
			OrganizationNPIID:       endpoint.NPIID,
			OrganizationZipCode:     endpoint.OrganizationZipCode,
			OrganizationIdentifiers: endpoint.OrganizationIdentifiers,
			OrganizationAddresses:   endpoint.OrganizationAddresses,
			OrganizationActive:      endpoint.OrganizationActive,
		}

		dbEntry.OrganizationList = []*endpointmanager.FHIREndpointOrganization{&dbOrgEntry}
	} else {
		dbEntry.OrganizationList = []*endpointmanager.FHIREndpointOrganization{}
	}

	return &dbEntry, nil
}

// RemoveOldEndpoints removes fhir endpoints from fhir_endpoints and fhir_endpoints_info
// that are no longer in the given listsource
func RemoveOldEndpoints(ctx context.Context, store *postgresql.Store, updateTime time.Time, listSource string) error {
	// get endpoints that are from this listsource and have an update time before this time
	fhirEndpoints, err := store.GetFHIREndpointsUsingListSourceAndUpdateTime(ctx, updateTime, listSource)
	if err != nil {
		return err
	}

	for _, endpoint := range fhirEndpoints {
		err = store.DeleteFHIREndpoint(ctx, endpoint)
		if err != nil {
			log.Warn(err)
			continue
		}
		existingEndpointList, err := store.GetFHIREndpointInfosUsingURL(ctx, endpoint.URL)
		if err == sql.ErrNoRows {
			log.Warn(err)
			continue
		} else {
			endpointList, err := store.GetFHIREndpointUsingURL(ctx, endpoint.URL)
			if err != nil {
				log.Warn(err)
				continue
			}
			if len(endpointList) == 0 {
				for _, existingEndpoint := range existingEndpointList {
					err = store.DeleteFHIREndpointInfo(ctx, existingEndpoint)
					if err != nil {
						log.Warn(err)
						continue
					}
				}
			}
		}
	}

	log.Infof("Removed %d endpoints from list source %s", len(fhirEndpoints), listSource)

	return nil
}

// RemoveOldEndpointOrganizations removes fhir endpoint organizations from fhir_endpoint_organizations
// that are no longer in the given endpoint's list of organizations
func RemoveOldEndpointOrganizations(ctx context.Context, store *postgresql.Store, updateTime time.Time, listSource string) error {
	// get endpoint organizations that are from this listsource and have an update time before this time
	fhirEndpoints, err := store.GetFHIREndpointsByListSourceAndOrganizationsUpdatedAtTime(ctx, updateTime, listSource)
	if err != nil {
		return err
	}

	for _, endpoint := range fhirEndpoints {
		for _, org := range endpoint.OrganizationList {
			err = store.DeleteFHIREndpointOrganization(ctx, org, endpoint.ID)
			if err != nil {
				log.Warn(err)
				continue
			}
		}
	}

	log.Infof("Removed %d endpoints organizations from list source %s", len(fhirEndpoints), listSource)

	return nil
}

func isValidURL(url string) bool {
	urlregex := regexp.MustCompile(`^(?:http(s)?:\/\/)?[\w.-]+(?:\.[\w\.-]+)+[\w\-\._~:/?#[\]@!\$&'\(\)\*\+,;=.]+$`)
	urlmatched := urlregex.MatchString(strings.ToLower(url))

	return urlmatched
}
