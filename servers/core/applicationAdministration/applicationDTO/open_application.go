package applicationDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type OpenApplication struct {
	CourseName               string      `json:"courseName"`
	CoursePhaseID            uuid.UUID   `json:"id"`
	CourseType               string      `json:"courseType"`
	ECTS                     int         `json:"ects"`
	StartDate                pgtype.Date `json:"startDate" swaggertype:"string"`
	EndDate                  pgtype.Date `json:"endDate" swaggertype:"string"`
	ApplicationDeadline      string      `json:"applicationDeadline"`
	ExternalStudentsAllowed  bool        `json:"externalStudentsAllowed"`
	UniversityLoginAvailable bool        `json:"universityLoginAvailable"`
	ShortDescription         *string     `json:"shortDescription,omitempty"`
	LongDescription          *string     `json:"longDescription,omitempty"`
	// WelcomeText is instructor-authored HTML shown above the application form.
	WelcomeText *string `json:"welcomeText,omitempty"`
	// ProfilePicture tells the form whether and why to ask logged-in applicants for a picture.
	// Only set on the form of one application, not in the list of open applications.
	ProfilePicture *ApplicationProfilePicture `json:"profilePicture,omitempty"`
}

// ApplicationProfilePicture is the profile picture configuration an applicant sees.
type ApplicationProfilePicture struct {
	// Requirement is "off", "optional", or "required".
	Requirement string `json:"requirement"`
	Explanation string `json:"explanation,omitempty"`
	// HiddenUntilAccepted tells applicants that reviewers do not see their picture.
	HiddenUntilAccepted bool `json:"hiddenUntilAccepted"`
}

func GetOpenApplicationPhaseDTO(dbModel db.GetAllOpenApplicationPhasesRow) OpenApplication {
	var shortDesc *string
	if dbModel.ShortDescription.Valid {
		shortDesc = &dbModel.ShortDescription.String
	}

	var longDesc *string
	if dbModel.LongDescription.Valid {
		longDesc = &dbModel.LongDescription.String
	}

	return OpenApplication{
		CourseName:               dbModel.CourseName,
		CoursePhaseID:            dbModel.CoursePhaseID,
		CourseType:               string(dbModel.CourseType),
		ECTS:                     int(dbModel.Ects.Int32),
		StartDate:                dbModel.StartDate,
		EndDate:                  dbModel.EndDate,
		ApplicationDeadline:      dbModel.ApplicationEndDate,
		ExternalStudentsAllowed:  dbModel.ExternalStudentsAllowed,
		UniversityLoginAvailable: dbModel.UniversityLoginAvailable,
		ShortDescription:         shortDesc,
		LongDescription:          longDesc,
	}
}
