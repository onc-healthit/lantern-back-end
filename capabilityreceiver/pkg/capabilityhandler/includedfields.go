package capabilityhandler

func RunIncludedFieldsChecks(capInt map[string]interface{}) map[string]bool {
	if capInt == nil {
		return nil
	}
	includedFields := make(map[string]bool)

	fieldsList := [][]string{
		{"url"},
		{"version"},
		{"name"},
		{"title"},
		{"status"},
		{"experimental"},
		{"date"},
		{"publisher"},
		{"contact"},
		{"contact", "name"},
		{"contact", "telecom"},
		{"description"},
		{"requirements"},
		{"useContext"},
		{"jurisdiction"},
		{"purpose"},
		{"copyright"},
		{"kind"},
		{"instantiates"},
		{"imports"},
		{"software", "name"},
		{"software", "version"},
		{"software", "releaseDate"},
		{"implementation", "description"},
		{"implementation", "url"},
		{"implementation", "custodian"},
		{"fhirVersion"},
		{"format"},
		{"patchFormat"},
		{"acceptUnknown"},
		{"implementationGuide"},
		{"profile"},
	}

	for _, fieldNames := range fieldsList {
		var stringIndex string
		if len(fieldNames) == 2 {
			stringIndex = fieldNames[0] + "." + fieldNames[1]
		} else {
			stringIndex = fieldNames[0]
		}

		includedFields[stringIndex] = checkField(capInt, fieldNames)
	}

	return includedFields
}

func checkField(capInt map[string]interface{}, fieldNames []string) bool {
	for index, name := range fieldNames {
		if capInt[name] == nil {
			return false
		}

		field := capInt[name]

		if index == (len(fieldNames) - 1) {
			if field != nil {
				return true
			}
			return false
		}

		if name == "contact" {
			capArr := field.([]interface{})
			capInt = capArr[0].(map[string]interface{})
		} else {
			capInt = field.(map[string]interface{})
		}

	}

	return false
}
