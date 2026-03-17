package teaseDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	log "github.com/sirupsen/logrus"
)

const (
	KEY_LANGUAGE_EN = "language_proficiency_english"
	KEY_LANGUAGE_DE = "language_proficiency_german"
)

// Student is the TEASE representation of a student's data.
type Student struct {
	CourseParticipationID  uuid.UUID              `json:"id"` // use the coursePhaseID, though in TEASE it's called StudentID
	FirstName              string                 `json:"firstName"`
	LastName               string                 `json:"lastName"`
	Gender                 string                 `json:"gender"`
	Nationality            string                 `json:"nationality"`
	Email                  string                 `json:"email"`
	StudyDegree            string                 `json:"studyDegree"`
	StudyProgram           string                 `json:"studyProgram"`
	Semester               pgtype.Int4            `json:"semester"`
	Languages              []Language             `json:"languages"`
	IntroSelfEvaluation    string                 `json:"introSelfEvaluation"`
	IntroCourseProficiency string                 `json:"introCourseProficiency"`
	Skills                 []StudentSkillResponse `json:"skills"`
	Devices                []string               `json:"devices"`
	StudentComments        []string               `json:"studentComments"` // @rappm pls update once your assessment is done
	TutorComments          []string               `json:"tutorComments"`   // @rappm pls update once your assessment is done
	ProjectPreferences     []ProjectPreference    `json:"projectPreferences"`
}

// convertPromptGenderToTease maps the promptSDK gender to TEASE gender labels.
func convertPromptGenderToTease(gender promptTypes.Gender) string {
	switch gender {
	case promptTypes.GenderFemale:
		return "Female"
	case promptTypes.GenderMale:
		return "Male"
	case promptTypes.GenderDiverse:
		return "Other"
	case promptTypes.GenderPreferNotToSay:
		return "Prefer not to say"
	default:
		return "Unknown"
	}
}

// ConvertCourseParticipationToTeaseStudent transforms a promptSDK course participation
// into a TEASE-compatible Student struct.
func ConvertCourseParticipationToTeaseStudent(
	cp promptTypes.CoursePhaseParticipationWithStudent,
	projectPreferences []ProjectPreference,
	skillResponses []StudentSkillResponse,
) (Student, error) {
	// 1) Read the application answers from meta data
	_, multiSelectAnswers, err := promptTypes.ReadApplicationAnswersFromMetaData(cp.PrevData["applicationAnswers"])
	if err != nil {
		log.WithField("courseParticipationID", cp.CourseParticipationID).
			WithError(err).
			Error("Could not read application answers from metadata")
		return Student{}, err
	}

	// 2) Attempt to read the "devices" field as []string
	devices := parseDeviceData(cp.PrevData["devices"])

	// 3) Attempt to read the scoreLevel field as a string
	scoreLevel, ok := cp.PrevData["scoreLevel"].(string)
	if !ok {
		log.WithField("courseParticipationID", cp.CourseParticipationID).
			Error("Field 'scoreLevel' in PrevData is not a string; using empty string")
		scoreLevel = ""
	}

	teaseProficiency := getTeaseScoreLevel(scoreLevel)

	// 3) Build a Student object
	student := Student{
		CourseParticipationID:  cp.CourseParticipationID,
		FirstName:              cp.Student.FirstName,
		LastName:               cp.Student.LastName,
		Gender:                 convertPromptGenderToTease(cp.Student.Gender),
		Nationality:            cp.Student.Nationality,
		Email:                  cp.Student.Email,
		StudyDegree:            string(cp.Student.StudyDegree),
		StudyProgram:           cp.Student.StudyProgram,
		Semester:               cp.Student.CurrentSemester,
		Languages:              []Language{},
		Devices:                devices,
		Skills:                 skillResponses,
		ProjectPreferences:     projectPreferences,
		StudentComments:        []string{},
		TutorComments:          []string{},
		IntroCourseProficiency: teaseProficiency,
		IntroSelfEvaluation:    teaseProficiency, // we currently use the same for both as we do not yet have a self-assessment
	}

	// 4) Process multi-select answers for language proficiency
	for _, ans := range multiSelectAnswers {
		switch ans.Key {
		case KEY_LANGUAGE_EN:
			addLanguageProficiency(&student, cp, ans, "en")
		case KEY_LANGUAGE_DE:
			addLanguageProficiency(&student, cp, ans, "de")
		}
	}

	return student, nil
}

// addLanguageProficiency is a helper that checks that the multi-select answer
// has exactly one item, and then appends it to the student's Languages slice.
func addLanguageProficiency(
	stud *Student,
	cp promptTypes.CoursePhaseParticipationWithStudent,
	answer promptTypes.AnswersMultiSelect,
	langCode string,
) {
	if len(answer.Answer) != 1 {
		log.WithField("courseParticipationID", cp.CourseParticipationID).
			WithField("key", answer.Key).
			Error("Unexpected number of answers for language proficiency: ", answer.Answer)
		return
	}
	proficiency := LanguageProficiency(answer.Answer[0])
	stud.Languages = append(stud.Languages, Language{
		Language:    langCode,
		Proficiency: proficiency,
	})
}

func parseDeviceData(rawDevices interface{}) []string {
	if rawDevices == nil {
		log.Error("devices is nil")
		return []string{}
	}
	devices := []string{}
	for i, v := range rawDevices.([]interface{}) {
		s, ok := v.(string)
		if !ok {
			log.Errorf("devices[%d] is not a string", i)
			continue
		}
		devices = append(devices, s)
	}
	return devices
}

// Tease defines the Scores upper case, while the DB (and Assessment) defines them lower case
// Maps 5-level score levels (VeryBad-VeryGood) to 4-level skill proficiency (Novice-Expert)
func getTeaseScoreLevel(skillLevel string) string {
	switch skillLevel {
	case "veryBad", "bad":
		return "Novice"
	case "ok":
		return "Intermediate"
	case "good":
		return "Advanced"
	case "veryGood":
		return "Expert"
	default:
		return "Unknown"
	}
}
