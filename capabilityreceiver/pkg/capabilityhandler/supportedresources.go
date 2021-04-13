package capabilityhandler

// RunSupportedResourcesChecks takes the given capability statement and creates an array
// of the resources and their specified operations in OperationAndResource format. Example:
// [ { Resource: "AllergyInformation", Operation: "read" },
//   { Resource: "AllergyInforamtion", Operation: "search-type" },
//   { Resource: "Medication": Operation: "read" }, ...]
// @TODO go through commented out stuff
func RunSupportedResourcesChecks(capInt map[string]interface{}) map[string][]string {
	// var opAndRes []endpointmanager.OperationAndResource
	var mapOpToResList = make(map[string][]string)
	if capInt == nil {
		// return opAndRes
		return mapOpToResList
	}

	// Get the resource field from the Capability Statement, which is a list of resources
	if capInt["rest"] == nil {
		// return opAndRes
		return mapOpToResList
	}
	restArr := capInt["rest"].([]interface{})
	restInt := restArr[0].(map[string]interface{})
	if restInt["resource"] == nil {
		// return opAndRes
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
		if !ok {
			notSpec = true
		} else if len(operations) == 0 {
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
				// item := endpointmanager.OperationAndResource{
				// 	Operation: code,
				// 	Resource:  resourceType,
				// }
				// opAndRes = append(opAndRes, item)
				if mapOpToResList[code] == nil {
					mapOpToResList[code] = []string{resourceType}
				} else {
					mapOpToResList[code] = append(mapOpToResList[code], resourceType)
				}
			}
		}
		// If the interaction field was not specified or it has no valid operations
		if notSpec || !hasCodes {
			// item := endpointmanager.OperationAndResource{
			// 	Operation: "not specified",
			// 	Resource:  resourceType,
			// }
			// opAndRes = append(opAndRes, item)
			if mapOpToResList["not specified"] == nil {
				mapOpToResList["not specified"] = []string{resourceType}
			} else {
				mapOpToResList["not specified"] = append(mapOpToResList["not specified"], resourceType)
			}
		}
	}

	// return opAndRes
	return mapOpToResList
}
