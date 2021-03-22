package capabilityhandler

import "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager"

// RunSupportedResourcesChecks  @TODO update this
func RunSupportedResourcesChecks(capInt map[string]interface{}) ([]string, []endpointmanager.OperationAndResource) {
	if capInt == nil {
		return nil, nil
	}
	// @TODO Remove
	var supportedResources []string

	// Get the resource field from the Capability Statement, which is a list of resources
	if capInt["rest"] == nil {
		return nil, nil
	}
	restArr := capInt["rest"].([]interface{})
	restInt := restArr[0].(map[string]interface{})
	if restInt["resource"] == nil {
		return nil, nil
	}
	resourceArr := restInt["resource"].([]interface{})

	var opAndRes []endpointmanager.OperationAndResource
	for _, resource := range resourceArr {
		resourceInt := resource.(map[string]interface{})
		if resourceInt["type"] == nil {
			return nil, nil
		}
		resourceType := resourceInt["type"].(string)
		// @TODO Remove?
		supportedResources = append(supportedResources, resourceType)

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
				item := endpointmanager.OperationAndResource{
					Operation: code,
					Resource:  resourceType,
				}
				opAndRes = append(opAndRes, item)
			}
		}
		// If the interaction field was not specified or it has no valid operations
		if notSpec || !hasCodes {
			item := endpointmanager.OperationAndResource{
				Operation: "not specified",
				Resource:  resourceType,
			}
			opAndRes = append(opAndRes, item)
		}
	}

	return supportedResources, opAndRes
}
