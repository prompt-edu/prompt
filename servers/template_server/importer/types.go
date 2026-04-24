package importer

import "github.com/google/uuid"

type PromptReadyTemplate struct {
	TemplateMeta                 templateMeta                 `json:"templateMeta"`
	CourseCreationDefaults       courseCreationDefaults       `json:"courseCreationDefaults"`
	ApplicationFormUpdatePayload applicationFormUpdatePayload `json:"applicationFormUpdatePayload"`
	TeamAllocationPhaseSetup     teamAllocationPhaseSetup     `json:"teamAllocationPhaseSetup"`
}

type templateMeta struct {
	TemplateID             string `json:"templateId"`
	TemplateVersion        int    `json:"templateVersion"`
	Name                   string `json:"name"`
	RecommendedCourseType  string `json:"recommendedCourseType"`
	TeamAllocationMode     string `json:"teamAllocationMode"`
	DefaultStrategyInTease string `json:"defaultStrategyInTease"`
	ReplaceExistingByID    bool   `json:"replaceExistingById"`
}

type courseCreationDefaults struct {
	Name                string         `json:"name"`
	SemesterTag         string         `json:"semesterTag"`
	Ects                int            `json:"ects"`
	ShortDescription    string         `json:"shortDescription"`
	LongDescription     string         `json:"longDescription"`
	StudentReadableData map[string]any `json:"studentReadableData"`
}

type applicationFormUpdatePayload struct {
	DeleteQuestionsText        []string                    `json:"deleteQuestionsText"`
	DeleteQuestionsMultiSelect []string                    `json:"deleteQuestionsMultiSelect"`
	DeleteQuestionsFileUpload  []string                    `json:"deleteQuestionsFileUpload"`
	CreateQuestionsText        []applicationQuestionText   `json:"createQuestionsText"`
	CreateQuestionsMultiSelect []applicationQuestionSelect `json:"createQuestionsMultiSelect"`
	CreateQuestionsFileUpload  []applicationQuestionFile   `json:"createQuestionsFileUpload"`
	UpdateQuestionsText        []applicationQuestionText   `json:"updateQuestionsText"`
	UpdateQuestionsMultiSelect []applicationQuestionSelect `json:"updateQuestionsMultiSelect"`
	UpdateQuestionsFileUpload  []applicationQuestionFile   `json:"updateQuestionsFileUpload"`
}

type applicationQuestionText struct {
	ID                       string `json:"id,omitempty"`
	CoursePhaseID            string `json:"coursePhaseID,omitempty"`
	Title                    string `json:"title"`
	Description              string `json:"description"`
	Placeholder              string `json:"placeholder"`
	ValidationRegex          string `json:"validationRegex,omitempty"`
	ErrorMessage             string `json:"errorMessage"`
	IsRequired               bool   `json:"isRequired"`
	AllowedLength            int    `json:"allowedLength"`
	OrderNum                 int    `json:"orderNum"`
	AccessibleForOtherPhases bool   `json:"accessibleForOtherPhases"`
	AccessKey                string `json:"accessKey"`
}

type applicationQuestionSelect struct {
	ID                       string   `json:"id,omitempty"`
	CoursePhaseID            string   `json:"coursePhaseID,omitempty"`
	Title                    string   `json:"title"`
	Description              string   `json:"description"`
	Placeholder              string   `json:"placeholder"`
	ErrorMessage             string   `json:"errorMessage"`
	IsRequired               bool     `json:"isRequired"`
	MinSelect                int      `json:"minSelect"`
	MaxSelect                int      `json:"maxSelect"`
	Options                  []string `json:"options"`
	OrderNum                 int      `json:"orderNum"`
	AccessibleForOtherPhases bool     `json:"accessibleForOtherPhases"`
	AccessKey                string   `json:"accessKey"`
}

type applicationQuestionFile struct {
	ID                       string   `json:"id,omitempty"`
	CoursePhaseID            string   `json:"coursePhaseID,omitempty"`
	Title                    string   `json:"title"`
	Description              string   `json:"description"`
	Placeholder              string   `json:"placeholder"`
	ErrorMessage             string   `json:"errorMessage"`
	IsRequired               bool     `json:"isRequired"`
	AllowedFileTypes         []string `json:"allowedFileTypes"`
	OrderNum                 int      `json:"orderNum"`
	AccessibleForOtherPhases bool     `json:"accessibleForOtherPhases"`
	AccessKey                string   `json:"accessKey"`
}

type teamAllocationPhaseSetup struct {
	Skills []string `json:"skills"`
	Teams  []string `json:"teams"`
}

type coursePhaseType struct {
	ID                              uuid.UUID             `json:"id"`
	Name                            string                `json:"name"`
	BaseURL                         string                `json:"baseUrl"`
	RequiredParticipationInputDTOs  []participationDTORef `json:"requiredParticipationInputDTOs"`
	ProvidedParticipationOutputDTOs []participationDTORef `json:"providedParticipationOutputDTOs"`
}

type participationDTORef struct {
	ID      uuid.UUID `json:"id"`
	DtoName string    `json:"dtoName"`
}

type createCourseRequest struct {
	Name                string         `json:"name"`
	StartDate           string         `json:"startDate"`
	EndDate             string         `json:"endDate"`
	SemesterTag         string         `json:"semesterTag"`
	RestrictedData      map[string]any `json:"restrictedData"`
	StudentReadableData map[string]any `json:"studentReadableData"`
	ShortDescription    string         `json:"shortDescription"`
	LongDescription     string         `json:"longDescription"`
	CourseType          string         `json:"courseType"`
	Ects                int            `json:"ects"`
	Template            bool           `json:"template"`
}

type courseResponse struct {
	ID uuid.UUID `json:"id"`
}

type templateCourseResponse struct {
	ID             uuid.UUID      `json:"id"`
	Name           string         `json:"name"`
	SemesterTag    string         `json:"semesterTag"`
	CourseType     string         `json:"courseType"`
	RestrictedData map[string]any `json:"restrictedData"`
	Template       bool           `json:"template"`
}

type createCoursePhaseRequest struct {
	CourseID            uuid.UUID      `json:"courseID"`
	Name                string         `json:"name"`
	IsInitialPhase      bool           `json:"isInitialPhase"`
	RestrictedData      map[string]any `json:"restrictedData"`
	StudentReadableData map[string]any `json:"studentReadableData"`
	CoursePhaseTypeID   uuid.UUID      `json:"coursePhaseTypeID"`
}

type coursePhaseResponse struct {
	ID uuid.UUID `json:"id"`
}

type coursePhaseGraphItem struct {
	FromCoursePhaseID uuid.UUID `json:"fromCoursePhaseID"`
	ToCoursePhaseID   uuid.UUID `json:"toCoursePhaseID"`
}

type updateCoursePhaseGraphRequest struct {
	InitialPhase     uuid.UUID              `json:"initialPhase"`
	CoursePhaseGraph []coursePhaseGraphItem `json:"coursePhaseGraph"`
}

type metaDataGraphItem struct {
	FromCoursePhaseID    uuid.UUID `json:"fromCoursePhaseID"`
	ToCoursePhaseID      uuid.UUID `json:"toCoursePhaseID"`
	FromCoursePhaseDtoID uuid.UUID `json:"fromCoursePhaseDtoID"`
	ToCoursePhaseDtoID   uuid.UUID `json:"toCoursePhaseDtoID"`
}

type createSkillsRequest struct {
	SkillNames []string `json:"skillNames"`
}

type createTeamsRequest struct {
	TeamNames []string `json:"teamNames"`
}

type ImportPromptReadyTemplateResponse struct {
	CourseID                uuid.UUID `json:"courseId"`
	ApplicationPhaseID      uuid.UUID `json:"applicationPhaseId"`
	TeamAllocationPhaseID   uuid.UUID `json:"teamAllocationPhaseId"`
	TeamAllocationMode      string    `json:"teamAllocationMode"`
	DefaultMatchingStrategy string    `json:"defaultMatchingStrategy"`
	Warnings                []string  `json:"warnings"`
}

func (p applicationFormUpdatePayload) withCoursePhaseID(coursePhaseID uuid.UUID) applicationFormUpdatePayload {
	phaseID := coursePhaseID.String()
	payload := p

	for i := range payload.CreateQuestionsText {
		payload.CreateQuestionsText[i].CoursePhaseID = phaseID
	}
	for i := range payload.CreateQuestionsMultiSelect {
		payload.CreateQuestionsMultiSelect[i].CoursePhaseID = phaseID
	}
	for i := range payload.CreateQuestionsFileUpload {
		payload.CreateQuestionsFileUpload[i].CoursePhaseID = phaseID
	}

	// Importer intentionally only creates questions in the first version.
	payload.UpdateQuestionsText = nil
	payload.UpdateQuestionsMultiSelect = nil
	payload.UpdateQuestionsFileUpload = nil
	payload.DeleteQuestionsText = []string{}
	payload.DeleteQuestionsMultiSelect = []string{}
	payload.DeleteQuestionsFileUpload = []string{}

	return payload
}
