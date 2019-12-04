package mock

import (
	"context"
	"database/sql"
	"errors"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

type BasicMockStore struct {
	HealthITProductData []*endpointmanager.HealthITProduct
	Store
}

func NewBasicMockHealthITProductStore() endpointmanager.HealthITProductStore {
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

	return &store
}
