package capabilityhandler

func RunSupportedResourcesChecks(capInt map[string]interface{}) []string {
	if capInt == nil {
		return nil
	}
	var supportedResources []string
	supportedResources = append(supportedResources, "testing")
	return supportedResources
}
