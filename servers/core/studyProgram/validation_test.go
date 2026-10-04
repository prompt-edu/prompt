package studyProgram

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateStudyProgram(t *testing.T) {
	tests := []struct {
		name      string
		program   string
		shortName string
		wantError bool
	}{
		{name: "name and short name", program: "Computer Science", shortName: "CS"},
		{name: "no short name", program: "Computer Science"},
		{name: "surrounding spaces", program: "  Computer Science  ", shortName: "  CS  "},
		{name: "name at the limit", program: strings.Repeat("a", maxNameLength)},
		{name: "multibyte name at the limit", program: strings.Repeat("ä", maxNameLength)},
		{name: "short name at the limit", program: "Physics", shortName: strings.Repeat("a", maxShortNameLength)},
		{name: "empty name", program: "", wantError: true},
		{name: "blank name", program: "   ", wantError: true},
		{name: "reserved name", program: "Other", wantError: true},
		{name: "reserved name in other casing", program: "  oTHER ", wantError: true},
		{name: "reserved short name", program: "Computer Science", shortName: " other ", wantError: true},
		{name: "unknown name", program: "Unknown", wantError: true},
		{name: "unknown name in other casing", program: " uNKNOWN  ", wantError: true},
		{name: "unknown short name", program: "Computer Science", shortName: "Unknown", wantError: true},
		{name: "name containing a reserved word", program: "Other Sciences", shortName: "Unknown Studies"},
		{name: "name too long", program: strings.Repeat("a", maxNameLength+1), wantError: true},
		{name: "short name too long", program: "Physics", shortName: strings.Repeat("a", maxShortNameLength+1), wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStudyProgram(tt.program, tt.shortName)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
