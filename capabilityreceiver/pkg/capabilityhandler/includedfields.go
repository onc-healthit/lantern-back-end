package capabilityhandler

// RunIncludedFieldsChecks stores whether each field in capability statement is populated or not populated
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
		{"messaging"},
		{"document"},
	}

	for _, fieldNames := range fieldsList {
		var stringIndex string
		if len(fieldNames) != 1 {
			for index, name := range fieldNames {
				if index == (len(fieldNames) - 1) {
					stringIndex = stringIndex + name
				} else if index == 0 {
					stringIndex = name + "."
				} else {
					stringIndex = stringIndex + "." + name
				}
			}
		} else {
			stringIndex = fieldNames[0]
		}

		includedFields[stringIndex] = checkField(capInt, fieldNames)
	}

	return includedFields
}

// Checks whether the given field is populated in the capability statement
func checkField(capInt map[string]interface{}, fieldNames []string) bool {
	for index, name := range fieldNames {
		if capInt[name] == nil {
			return false
		}

		field := capInt[name]

		if index == (len(fieldNames) - 1) {
			return field != nil
		}

		capInt = field.(map[string]interface{})

	}

	return false
}
