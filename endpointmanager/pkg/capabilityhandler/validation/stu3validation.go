package validation

type stu3Validation struct {
	baseVal
}

func newSTU3Val() *stu3Validation {
	return &stu3Validation{
		baseVal: baseVal{},
	}
}
