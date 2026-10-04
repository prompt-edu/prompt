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
	reservedName       = "Other"
)

func validateStudyProgram(name, shortName string) error {
	name = strings.TrimSpace(name)
	shortName = strings.TrimSpace(shortName)

	if name == "" {
		return errors.New("study program name is required")
	}
	if strings.EqualFold(name, reservedName) {
		return fmt.Errorf("%q is reserved for study programs entered as free text", reservedName)
	}
	if utf8.RuneCountInString(name) > maxNameLength {
		return fmt.Errorf("study program name must be at most %d characters", maxNameLength)
	}
	if utf8.RuneCountInString(shortName) > maxShortNameLength {
		return fmt.Errorf("short name must be at most %d characters", maxShortNameLength)
	}
	return nil
}
