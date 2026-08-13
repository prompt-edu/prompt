package student

import (
	"errors"
	"regexp"

	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	log "github.com/sirupsen/logrus"
)

func Validate(c studentDTO.CreateStudent) error {
	if err := validateName(c.FirstName, c.LastName); err != nil {
		return err
	}
	if err := validateEmail(c.Email); err != nil {
		return err
	}
	if err := validateUniversityData(c.HasUniversityAccount, c.MatriculationNumber, c.UniversityLogin); err != nil {
		return err
	}
	if c.Nationality == "" {
		log.Error("nationality is missing")
		return errors.New("nationality is missing")
	}

	if !c.CurrentSemester.Valid || c.CurrentSemester.Int32 < 1 {
		log.Error("semester is invalid")
		return errors.New("semester is invalid")
	}
	if c.StudyProgram == "" {
		log.Error("study program is invalid")
		return errors.New("study program is invalid")
	}
	if c.StudyDegree != db.StudyDegreeBachelor && c.StudyDegree != db.StudyDegreeMaster {
		log.Error("study degree is invalid")
		return errors.New("study degree is invalid")
	}
	return nil
}

// ValidateImport is a lenient variant of Validate used for CSV student imports.
// It only enforces the mandatory identity fields (first name, last name, email and university
// login) plus their formats. Optional attributes (nationality, study program, current semester)
// and the enum defaults (gender, study degree) are filled in by the import service, so they are
// not required here.
func ValidateImport(c studentDTO.CreateStudent) error {
	if err := validateName(c.FirstName, c.LastName); err != nil {
		return err
	}
	if err := validateEmail(c.Email); err != nil {
		return err
	}
	if c.UniversityLogin == "" {
		return errors.New("university login is required")
	}
	// Imports always identify students by their university login, so validate the university data as
	// if the student has an account regardless of the (import-service-managed) HasUniversityAccount flag.
	if err := validateUniversityData(true, c.MatriculationNumber, c.UniversityLogin); err != nil {
		return err
	}
	// Optional attributes are omitted-then-defaulted by the import service, but a value the CSV does
	// provide is written verbatim to the globally shared student row, so reject an invalid one here.
	if c.CurrentSemester.Valid && c.CurrentSemester.Int32 < 1 {
		return errors.New("semester is invalid")
	}
	if c.StudyDegree != "" && c.StudyDegree != db.StudyDegreeBachelor && c.StudyDegree != db.StudyDegreeMaster {
		return errors.New("study degree is invalid")
	}
	if c.Gender != "" && c.Gender != db.GenderMale && c.Gender != db.GenderFemale &&
		c.Gender != db.GenderDiverse && c.Gender != db.GenderPreferNotToSay {
		return errors.New("gender is invalid")
	}
	return nil
}

func validateName(firstName, lastName string) error {
	if firstName == "" {
		log.Error("first name is required")
		return errors.New("first name is required")
	}
	if lastName == "" {
		log.Error("last name is required")
		return errors.New("last name is required")
	}
	return nil
}

func validateEmail(email string) error {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	if matched, _ := regexp.MatchString(emailRegex, email); !matched {
		return errors.New("invalid email address")
	}

	return nil
}

func validateUniversityData(hasUniversityAccount bool, matriculationNumber, universityLogin string) error {
	if hasUniversityAccount {
		// Schema for TUM matriculation number
		matriculationNumberRegex := `^0\d{7}$`

		// Schema for TUM ID
		universityLoginRegex := `^[a-zA-Z]{2}\d{2}[a-zA-Z]{3}$`

		// Matriculation number is optional (external TUM members may not have one)
		if matriculationNumber != "" {
			if matched, _ := regexp.MatchString(matriculationNumberRegex, matriculationNumber); !matched {
				return errors.New("invalid matriculation number")
			}
		}

		if matched, _ := regexp.MatchString(universityLoginRegex, universityLogin); !matched {
			return errors.New("invalid university login")
		}
	} else if matriculationNumber != "" || universityLogin != "" {
		return errors.New("student has no university account but has university data")
	}

	return nil
}

