package endpointmanager

// NPIContact represents the digitial contact information for an NPI Contact provided by the NPPES database
type NPIContact struct {
	ID                           int
	NPI_ID                       string
	EndpointType                 string
	EndpointTypeDescription      string
	Endpoint                     string
	ValidURL                     bool
	Affiliation                  string
	EndpointDescription          string
	AffiliationLegalBusinessName string
	UseCode                      string
	UseDescription               string
	OtherUseDescription          string
	ContentType                  string
	ContentDescription           string
	OtherContentDescription      string
	Location                     *Location
}
