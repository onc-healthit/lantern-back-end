package mock

import (
	"context"
	"database/sql"
	"errors"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// FhirMockStore is a mocked store based on the FhirEndpoint store
type FhirMockStore struct {
	FhirEndpointData []*endpointmanager.FHIREndpoint
	Store
}

// NewBasicMockFhirEndpointStore returns mocked implementations of the AddFhirEndpoint and GetFHIREndpointUsingURL functions
// used in the FhirEndpoint store
func NewBasicMockFhirEndpointStore() endpointmanager.FHIREndpointStore {
	return newBasicFhirStore()
}

func newBasicFhirStore() *FhirMockStore {
	store := FhirMockStore{}

	store.AddFHIREndpointFn = func(ctx context.Context, fhirEndpt *endpointmanager.FHIREndpoint) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// ok
		}

		for _, existingFhirEndpt := range store.FhirEndpointData {
			if existingFhirEndpt.URL == fhirEndpt.URL {
				return errors.New("FHIR Endpoint with that URL already exists")
			}
		}
		// want to store a copy
		newFhirEndpt := *fhirEndpt
		newFhirEndpt.ID = len(store.FhirEndpointData) + 1
		store.FhirEndpointData = append(store.FhirEndpointData, &newFhirEndpt)
		return nil
	}

	store.GetFHIREndpointUsingURLFn = func(ctx context.Context, url string) (*endpointmanager.FHIREndpoint, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// ok
		}

		for _, existingFhirEndpt := range store.FhirEndpointData {
			if existingFhirEndpt.URL == url {
				// want to return a copy
				fhirEndpt := *existingFhirEndpt
				return &fhirEndpt, nil
			}
		}
		return nil, sql.ErrNoRows
	}

	return &store
}
