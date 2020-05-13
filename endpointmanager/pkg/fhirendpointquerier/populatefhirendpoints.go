package populatefhirendpoints

import (
	"context"

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
