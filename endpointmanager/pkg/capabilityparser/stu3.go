package capabilityparser

type stu3CapabilityParser struct {
	baseParser
}

func newSTU3(capStat map[string]interface{}) *stu3CapabilityParser {
	return &stu3CapabilityParser{
		baseParser: baseParser{
			capStat: capStat,
			version: "STU3",
		},
	}
}
