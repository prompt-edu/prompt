package teamDTO

type CreateTeamsRequest struct {
	TeamNames           []string                      `json:"teamNames"`
	TeamType            string                        `json:"teamType,omitempty"`
	ReplaceExisting     bool                          `json:"replaceExisting,omitempty"`
	TeamSizeConstraints map[string]TeamSizeConstraint `json:"teamSizeConstraints,omitempty"`
}

type TeamSizeConstraint struct {
	LowerBound int32 `json:"lowerBound"`
	UpperBound int32 `json:"upperBound"`
}
