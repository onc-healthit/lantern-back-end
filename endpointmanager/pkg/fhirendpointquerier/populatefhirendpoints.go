package populatefhirendpoints

import (
	"context"
	"database/sql"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/networkstatsquerier/fetcher"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// AddEndpointData iterates through the list of endpoints and adds each one to the database
func AddEndpointData(ctx context.Context, store *postgresql.Store, endpoints *fetcher.ListOfEndpoints) error {
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

	existingEndpt, err := store.GetFHIREndpointUsingURL(ctx, fhirEndpoint.URL)
	// If the URL doesn't exist, add it to the DB
	if err == sql.ErrNoRows {
		err = store.AddFHIREndpoint(ctx, fhirEndpoint)
		if err != nil {
			return errors.Wrap(err, "adding fhir endpoint to store failed")
		}
	} else if err != nil {
		return errors.Wrap(err, "getting fhir endpoint from store failed")
	} else {
		// Always overwrite the db entry with the new data
		existingEndpt.OrganizationName = fhirEndpoint.OrganizationName
		existingEndpt.ListSource = fhirEndpoint.ListSource
		err = store.UpdateFHIREndpoint(ctx, existingEndpt)
		if err != nil {
			return err
		}
		log.Infof("Endpoint already exists (%s, %s). List source %s is overwriting it.", existingEndpt.URL, existingEndpt.OrganizationName, existingEndpt.ListSource)
	}
	return nil
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
		URL:              uri,
		OrganizationName: endpoint.OrganizationName,
		ListSource:       endpoint.ListSource,
	}

	// @TODO Get Location

	return &dbEntry, nil
}
