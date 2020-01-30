package capabilityparser

type r4CapabilityParser struct {
	baseParser
}

func newR4(capStat map[string]interface{}) *r4CapabilityParser {
	return &r4CapabilityParser{
		baseParser: baseParser{
			capStat: capStat,
			version: "R4",
		},
	}
}
