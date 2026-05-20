package endpointmanager

import "time"

type EndpointQueryError struct {
	ID           int
	ListSource   string
	ErrorMessage string
	QueriedAt    time.Time
	CreatedAt    time.Time
}
