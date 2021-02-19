package capabilityhandler

func RunSupportedResourcesChecks(capInt map[string]interface{}) ([]string, map[string][]string) {
	if capInt == nil {
		return nil, nil
	}
	// @TODO Remove?
	var supportedResources []string

	if capInt["rest"] == nil {
		return nil, nil
	}
	restArr := capInt["rest"].([]interface{})
	restInt := restArr[0].(map[string]interface{})
	if restInt["resource"] == nil {
		return nil, nil
	}
	resourceArr := restInt["resource"].([]interface{})

	opToRes := make(map[string][]string)
	for _, resource := range resourceArr {
		resourceInt := resource.(map[string]interface{})
		if resourceInt["type"] == nil {
			return nil, nil
		}
		resourceType := resourceInt["type"].(string)
		// @TODO Remove?
		supportedResources = append(supportedResources, resourceType)

		// Keep track of each resource type's given operations specified in the
		// capability statement
		notSpec := false
		hasCodes := false
		operations, ok := resourceInt["interaction"].([]interface{})
		if !ok {
			notSpec = true
		} else if len(operations) == 0 {
			notSpec = true
		} else {
			// Add the above resourceType to each specified code in the map
			// e.g. { "read": ["AllergyIntolerance", "Conformance"],
			// "write": ["AllergyIntolerance"] }
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
				if c, ok := opToRes[code]; ok {
					opToRes[code] = append(c, resourceType)
				} else {
					opToRes[code] = []string{resourceType}
				}
			}
		}
		// If the interaction field was not specified or has no given codes
		// Then add the resource to the "not specified" key in the map
		if notSpec || !hasCodes {
			if c, ok := opToRes["not specified"]; ok {
				opToRes["not specified"] = append(c, resourceType)
			} else {
				opToRes["not specified"] = []string{resourceType}
			}
		}
	}

	// @TODO Remove
	// url, ok := capInt["url"].(string)
	// if ok {
	// 	log.Infof("THE URL (hopefully), %s", url)
	// }
	// log.Infof("The Operations: %+v", opToRes)

	return supportedResources, opToRes
}
