package mock

import (
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"
)

// GetProviderOrganization mocks endpointmanager.ProviderOrganizationStore.GetProviderOrganization and sets s.GetProviderOrganizationInvoked to true and calls s.GetProviderOrganizationFn with the given arguments.
func (s *Store) GetProviderOrganization(id int) (*endpointmanager.ProviderOrganization, error) {
	s.GetProviderOrganizationInvoked = true
	return s.GetProviderOrganizationFn(id)
}

// AddProviderOrganization mocks endpointmanager.ProviderOrganizationStore.AddProviderOrganization and sets s.AddProviderOrganizationInvoked to true and calls s.AddProviderOrganizationFn with the given arguments.
func (s *Store) AddProviderOrganization(po *endpointmanager.ProviderOrganization) error {
	s.AddProviderOrganizationInvoked = true
	return s.AddProviderOrganizationFn(po)
}

// UpdateProviderOrganization mocks endpointmanager.ProviderOrganizationStore.UpdateProviderOrganization and sets s.UpdateProviderOrganizationInvoked to true and calls s.UpdateProviderOrganizationFn with the given arguments.
func (s *Store) UpdateProviderOrganization(po *endpointmanager.ProviderOrganization) error {
	s.UpdateProviderOrganizationInvoked = true
	return s.UpdateProviderOrganizationFn(po)
}

// DeleteProviderOrganization mocks endpointmanager.ProviderOrganizationStore.DeleteProviderOrganization and sets s.DeleteProviderOrganizationInvoked to true and calls s.DeleteProviderOrganizationFn with the given arguments.
func (s *Store) DeleteProviderOrganization(po *endpointmanager.ProviderOrganization) error {
	s.DeleteProviderOrganizationInvoked = true
	return s.DeleteProviderOrganizationFn(po)
}
