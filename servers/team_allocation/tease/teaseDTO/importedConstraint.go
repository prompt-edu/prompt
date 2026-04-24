package teaseDTO

type ImportedConstraint struct {
	Type       string `json:"type"`
	LowerBound int32  `json:"lowerBound"`
	UpperBound int32  `json:"upperBound"`
}
