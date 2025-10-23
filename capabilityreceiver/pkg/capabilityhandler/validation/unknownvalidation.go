package validation

type unknownValidation struct {
	baseVal
}

func newUnknownVal() *unknownValidation {
	return &unknownValidation{
		baseVal: baseVal{},
	}
}
