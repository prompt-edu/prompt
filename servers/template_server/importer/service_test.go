package importer

import (
	"testing"

	"github.com/google/uuid"
)

func TestNormalizeCourseType(t *testing.T) {
	tests := map[string]string{
		"Seminar":          "seminar",
		"lecture":          "lecture",
		"Practical Course": "practical course",
		"something else":   "seminar",
	}

	for input, expected := range tests {
		if actual := normalizeCourseType(input); actual != expected {
			t.Fatalf("normalizeCourseType(%q) = %q, expected %q", input, actual, expected)
		}
	}
}

func TestFindParticipationDTOIDHelpers(t *testing.T) {
	outputID := uuid.New()
	inputID := uuid.New()

	phaseType := coursePhaseType{
		Name: "Team Allocation",
		ProvidedParticipationOutputDTOs: []participationDTORef{
			{ID: outputID, DtoName: "applicationAnswers"},
		},
		RequiredParticipationInputDTOs: []participationDTORef{
			{ID: inputID, DtoName: "applicationAnswers"},
		},
	}

	foundOutputID, err := findParticipationOutputDTOID(phaseType, "applicationAnswers")
	if err != nil {
		t.Fatalf("unexpected error finding output DTO: %v", err)
	}
	if foundOutputID != outputID {
		t.Fatalf("unexpected output DTO id: got %s want %s", foundOutputID, outputID)
	}

	foundInputID, err := findParticipationInputDTOID(phaseType, "applicationAnswers")
	if err != nil {
		t.Fatalf("unexpected error finding input DTO: %v", err)
	}
	if foundInputID != inputID {
		t.Fatalf("unexpected input DTO id: got %s want %s", foundInputID, inputID)
	}
}

func TestMatchesTemplateID(t *testing.T) {
	restrictedData := map[string]any{
		"templateSource": map[string]any{
			"templateId": "project_week_1000plus_prompt_ready",
		},
	}

	if !matchesTemplateID(restrictedData, "project_week_1000plus_prompt_ready") {
		t.Fatalf("expected template ID to match")
	}

	if matchesTemplateID(restrictedData, "other_template") {
		t.Fatalf("expected template ID not to match")
	}
}

func TestValidatePromptReadyTemplateRequiresTemplateIDWhenReplacing(t *testing.T) {
	err := validatePromptReadyTemplate(PromptReadyTemplate{
		TemplateMeta: templateMeta{
			ReplaceExistingByID:    true,
			TeamAllocationMode:     "project_week_1000plus",
			DefaultStrategyInTease: "project-week-greedy",
		},
		CourseCreationDefaults: courseCreationDefaults{
			Name:        "Project Week 1000+ Template",
			SemesterTag: "template",
		},
	})
	if err == nil {
		t.Fatalf("expected validation error when replaceExistingById is true without templateId")
	}
}

func TestShouldReplaceTemplateCourse(t *testing.T) {
	course := templateCourseResponse{
		Name:        "Project Week 1000+ Template",
		SemesterTag: "template",
		Template:    true,
		RestrictedData: map[string]any{
			"templateSource": map[string]any{
				"templateId": "project_week_1000plus_prompt_ready",
			},
		},
	}

	if !shouldReplaceTemplateCourse(course, "project_week_1000plus_prompt_ready", "other", "other") {
		t.Fatalf("expected course to match by templateId")
	}

	course.RestrictedData = nil
	if !shouldReplaceTemplateCourse(course, "other_template", "Project Week 1000+ Template", "template") {
		t.Fatalf("expected course to match by name and semesterTag")
	}

	if shouldReplaceTemplateCourse(course, "other_template", "Different Name", "template") {
		t.Fatalf("did not expect course to match different template identity")
	}
}
