package validation

type dstu2Validation struct {
	baseVal
}

func newDSTU2Val() *dstu2Validation {
	return &dstu2Validation{
		baseVal: baseVal{},
	}
}
