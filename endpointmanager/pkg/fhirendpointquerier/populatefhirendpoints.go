package populatefhirendpoints

import (
	"context"
	"database/sql"
	"time"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// AddEndpointData iterates through the list of endpoints and adds each one to the database
func AddEndpointData(ctx context.Context, store *postgresql.Store, endpoints *fetcher.ListOfEndpoints) error {
	var firstUpdate time.Time
	var listsource = endpoints.Entries[0].ListSource
	for i, endpoint := range endpoints.Entries {
		select {
		case <-ctx.Done():
			return errors.Wrapf(ctx.Err(), "saved %d out of %d endpoints before context ended", i, len(endpoints.Entries))
		default:
			// ok
		}

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
			existingEndpt, err := store.GetFHIREndpointUsingURLAndListSource(ctx, fhirURL, endpoint.ListSource)
			if err != nil {
				log.Warn(err)
				continue
			} else {
				firstUpdate = existingEndpt.UpdatedAt
			}
		}
	}

	err := removeOldEndpoints(ctx, store, firstUpdate, listsource)
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

	// convert the endpoint entry to the fhirDatabase format
	dbEntry := endpointmanager.FHIREndpoint{
		URL:               uri,
		OrganizationNames: endpoint.OrganizationNames,
		ListSource:        endpoint.ListSource,
	}

	// @TODO Get Location

	return &dbEntry, nil
}

// removeOldEndpoints removes fhir endpoints from fhir_endpoints and fhir_endpoints_info
// that were not updated from given list source
func removeOldEndpoints(ctx context.Context, store *postgresql.Store, updateTime time.Time, listSource string) error {
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
		existingEndpoint, err := store.GetFHIREndpointInfoUsingURL(ctx, endpoint.URL)
		if err == sql.ErrNoRows {
			log.Warn(err)
			continue
		} else {
			err = store.DeleteFHIREndpointInfo(ctx, existingEndpoint)
			if err != nil {
				log.Warn(err)
				continue
			}
		}
	}

	log.Infof("Removed %d endpoints", len(fhirEndpoints))

	return nil
}
