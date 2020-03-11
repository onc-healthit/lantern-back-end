package mock

import (
	"context"
	"database/sql"
	"errors"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// BasicMockStore is a mocked store that includes both the HealthItProduct and FHIREndpoint stores
type BasicMockStore struct {
	HealthITProductData []*endpointmanager.HealthITProduct
	FhirEndpointData    []*endpointmanager.FHIREndpoint
	Store
}

// NewBasicMockHealthITProductStore returns mocked implementations of the functions
// used in the HealthITProductStore store
func NewBasicMockHealthITProductStore() endpointmanager.HealthITProductStore {
	return newBasicMockStore()
}

// NewBasicMockFhirEndpointStore returns mocked implementations of the AddFhirEndpoint and GetFHIREndpointUsingURL functions
// used in the FhirEndpoint store
func NewBasicMockFhirEndpointStore() endpointmanager.FHIREndpointStore {
	return newBasicMockStore()
}

func newBasicMockStore() *BasicMockStore {
	store := BasicMockStore{}

	store.AddHealthITProductFn = func(ctx context.Context, hitp *endpointmanager.HealthITProduct) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// ok
		}

		for _, existingHitp := range store.HealthITProductData {
			if existingHitp.ID == hitp.ID {
				return errors.New("HealthITProduct with that ID already exists")
			}
		}
		// want to store a copy
		newHitp := *hitp
		newHitp.ID = len(store.HealthITProductData) + 1
		store.HealthITProductData = append(store.HealthITProductData, &newHitp)
		return nil
	}

	store.GetHealthITProductUsingNameAndVersionFn = func(ctx context.Context, name string, version string) (*endpointmanager.HealthITProduct, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// ok
		}

		for _, existingHitp := range store.HealthITProductData {
			if existingHitp.Name == name && existingHitp.Version == version {
				// want to return a copy
				hitp := *existingHitp
				return &hitp, nil
			}
		}
		return nil, sql.ErrNoRows
	}

	store.UpdateHealthITProductFn = func(ctx context.Context, hitp *endpointmanager.HealthITProduct) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// ok
		}

		var existingHitp *endpointmanager.HealthITProduct
		var i int
		replace := false
		for i, existingHitp = range store.HealthITProductData {
			if existingHitp.ID == hitp.ID {
				replace = true
				break
			}
		}
		if replace {
			// replacing with copy
			updatedHitp := *hitp
			store.HealthITProductData[i] = &updatedHitp
		} else {
			return errors.New("No existing entry exists")
		}

		return nil
	}

	store.GetHealthITProductDevelopersFn = func(ctx context.Context) ([]string, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// ok
		}

		devList := []string{
			"Epic Systems Corporation",
			"Cerner Corporation",
			"Cerner Health Services, Inc.",
			"Medical Information Technology, Inc. (MEDITECH)",
			"Allscripts",
		}

		return devList, nil
	}

	// FHIREndpointStore Functions

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

	store.UpdateFHIREndpointFn = func(ctx context.Context, e *endpointmanager.FHIREndpoint) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// ok
		}

		var existingFhirEndpt *endpointmanager.FHIREndpoint
		var i int
		replace := false
		for i, existingFhirEndpt = range store.FhirEndpointData {
			if existingFhirEndpt.ID == e.ID {
				replace = true
				break
			}
		}
		if replace {
			// replacing with copy
			updatedFhirEndpt := *e
			store.FhirEndpointData[i] = &updatedFhirEndpt
		} else {
			return errors.New("No existing entry exists")
		}

		return nil
	}

	return &store
}
