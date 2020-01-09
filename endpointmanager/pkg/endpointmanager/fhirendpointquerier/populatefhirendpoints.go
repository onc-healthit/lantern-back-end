package populatefhirendpoints

import (
	"context"
	"database/sql"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
	"github.com/onc-healthit/lantern-back-end/endpoints/fetcher"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// AddEndpointData iterates through the list of endpoints and adds each one to the database
func AddEndpointData(ctx context.Context, store endpointmanager.FHIREndpointStore, endpoints *fetcher.ListOfEndpoints) error {
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
func saveEndpointData(ctx context.Context, store endpointmanager.FHIREndpointStore, endpoint *fetcher.EndpointEntry) error {
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
		// Currently just logging if there is a repeat
		log.Info(existingEndpt)
	}
	return nil
}

// formatToFHIREndpt takes an entry in the list of endpoints and formats it for the fhir_endpoints table in the database
func formatToFHIREndpt(endpoint *fetcher.EndpointEntry) (*endpointmanager.FHIREndpoint, error) {
	// convert the endpoint entry to the fhirDatabase format
	dbEntry := endpointmanager.FHIREndpoint{
		URL:              endpoint.FHIRPatientFacingURI,
		OrganizationName: endpoint.OrganizationName,
	}

	// @TODO Get Location

	return &dbEntry, nil
}
