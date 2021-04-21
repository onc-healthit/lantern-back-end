package capabilityhandler

// RunSupportedResourcesChecks takes the given capability statement and creates a map
// of the operations to the endpoint's resources that specified that operation. Example:
// { "read": ["AllergyInformation", "Medication"...],
//   "search-type": ["Medication", "Document"...], ...}
func RunSupportedResourcesChecks(capInt map[string]interface{}) map[string][]string {
	var mapOpToResList = make(map[string][]string)
	if capInt == nil {
		return mapOpToResList
	}

	// Get the resource field from the Capability Statement, which is a list of resources
	if capInt["rest"] == nil {
		return mapOpToResList
	}
	restArr := capInt["rest"].([]interface{})
	restInt := restArr[0].(map[string]interface{})
	if restInt["resource"] == nil {
		return mapOpToResList
	}
	resourceArr := restInt["resource"].([]interface{})

	for _, resource := range resourceArr {
		resourceInt := resource.(map[string]interface{})
		if resourceInt["type"] == nil {
			continue
		}
		resourceType := resourceInt["type"].(string)

		// Keep track of the operations defined by each resource
		notSpec := false
		hasCodes := false
		operations, ok := resourceInt["interaction"].([]interface{})
		// if there is no interaction field, or the list is empty,
		// then no operations were specified
		if !ok || len(operations) == 0 {
			notSpec = true
		} else {
			// For each given operation, make sure it's valid and then
			// add it to the list of operation and resource pairs
			for _, op := range operations {
				opMap, ok := op.(map[string]interface{})
				if !ok {
					continue
				}
				code, ok := opMap["code"].(string)
				if !ok {
					continue
				}
				hasCodes = true
				if mapOpToResList[code] == nil {
					mapOpToResList[code] = []string{resourceType}
				} else {
					mapOpToResList[code] = append(mapOpToResList[code], resourceType)
				}
			}
		}
		// If the interaction field was not specified or it has no valid operations
		if notSpec || !hasCodes {
			if mapOpToResList["not specified"] == nil {
				mapOpToResList["not specified"] = []string{resourceType}
			} else {
				mapOpToResList["not specified"] = append(mapOpToResList["not specified"], resourceType)
			}
		}
	}

	return mapOpToResList
}
