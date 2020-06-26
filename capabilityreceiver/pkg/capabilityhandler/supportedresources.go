package capabilityhandler

func RunSupportedResourcesChecks(capInt map[string]interface{}) []string {
	if capInt == nil {
		return nil
	}
	var supportedResources []string

	if capInt["rest"] == nil {
		return nil
	}
	restArr := capInt["rest"].([]interface{})
	restInt := restArr[0].(map[string]interface{})
	if restInt["resource"] == nil {
		return nil
	}
	resourceArr := restInt["resource"].([]interface{})

	for _, resource := range resourceArr {
		resourceInt := resource.(map[string]interface{})
		if resourceInt["type"] == nil {
			return nil
		}
		resourceType := resourceInt["type"].(string)
		supportedResources = append(supportedResources, resourceType)
	}

	return supportedResources
}
