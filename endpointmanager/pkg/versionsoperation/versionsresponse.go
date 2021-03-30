package versionsoperation

// base struct to handle any methods that don't change between the versions of FHIR
// capability statements
type VersionsResponse struct {
	versions []string		`json:"versions"`
	defaultVersion string	`json:"default"`
}
