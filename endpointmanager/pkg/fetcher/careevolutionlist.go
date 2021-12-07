package fetcher

// CareEvolutionList implements the Endpoints interface for the CareEvolution endpoint lists
type CareEvolutionList struct{}

// GetEndpoints takes the list of CareEvolution endpoints and formats it into a ListOfEndpoints
func (cl CareEvolutionList) GetEndpoints(careEvolutionList []map[string]interface{}, listURL string) ListOfEndpoints {
	return getDefaultEndpoints(careEvolutionList, "CareEvolution", listURL)
}
