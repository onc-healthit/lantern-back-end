package fetcher

// OneUpList implements the Endpoints interface for the OneUpList endpoint lists
type OneUpList struct{}

// GetEndpoints takes the list of 1Up endpoints and formats it into a ListOfEndpoints
func (ul OneUpList) GetEndpoints(oneUpList []map[string]interface{}, listURL string) ListOfEndpoints {
	return getDefaultEndpoints(oneUpList, "1Up", listURL)
}
