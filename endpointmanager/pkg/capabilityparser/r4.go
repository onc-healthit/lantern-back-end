package capabilityparser

type r4CapabilityParser struct {
	baseParser
}

func newR4(capStat map[string]interface{}) CapabilityStatement {
	return &r4CapabilityParser{
		baseParser: baseParser{
			capStat: capStat,
			version: "R4",
		},
	}
}
