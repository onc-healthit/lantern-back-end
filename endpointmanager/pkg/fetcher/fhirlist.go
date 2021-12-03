package fetcher

// FHIRList implements the Endpoints interface for an endpoint list in FHIR
type FHIRList struct{}

func (fl FHIRList) GetEndpoints(fhirList []map[string]interface{}, listURL string) ListOfEndpoints {
	return GetBundleEndpoints(fhirList, "FHIR", listURL, true)
}
