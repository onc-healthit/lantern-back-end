package capabilityparser

type dstu2CapabilityParser struct {
	baseParser
}

func newDSTU2(capStat map[string]interface{}) *dstu2CapabilityParser {
	return &dstu2CapabilityParser{
		baseParser: baseParser{
			capStat: capStat,
			version: "DSTU2",
		},
	}
}
