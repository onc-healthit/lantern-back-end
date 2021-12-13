package fetcher

// EpicList implements the Endpoints interface for the Epic endpoint lists
type EpicList struct{}

// GetEndpoints takes the list of epic endpoints and formats it into a ListOfEndpoints
func (el EpicList) GetEndpoints(epicList []map[string]interface{}, source string, listURL string) ListOfEndpoints {
	if source == "" {
		source = "Epic"
	}
	return getDefaultEndpoints(epicList, source, listURL)
}
