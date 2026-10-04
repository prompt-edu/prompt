package studyProgram

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	maxNameLength      = 100
	maxShortNameLength = 20
)

var reservedNames = []struct {
	name   string
	reason string
}{
	{name: "Other", reason: "study programs entered as free text"},
	{name: "Unknown", reason: "applications without a study program"},
}

func validateStudyProgram(name, shortName string) error {
	name = strings.TrimSpace(name)
	shortName = strings.TrimSpace(shortName)

	if name == "" {
		return errors.New("study program name is required")
	}
	for _, reserved := range reservedNames {
		if strings.EqualFold(name, reserved.name) || strings.EqualFold(shortName, reserved.name) {
			return fmt.Errorf("%q is reserved for %s", reserved.name, reserved.reason)
		}
	}
	if utf8.RuneCountInString(name) > maxNameLength {
		return fmt.Errorf("study program name must be at most %d characters", maxNameLength)
	}
	if utf8.RuneCountInString(shortName) > maxShortNameLength {
		return fmt.Errorf("short name must be at most %d characters", maxShortNameLength)
	}
	return nil
}
