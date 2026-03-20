package student

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

type StudentService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var StudentServiceSingleton *StudentService

func GetAllStudents(ctx context.Context) ([]studentDTO.Student, error) {
	students, err := StudentServiceSingleton.queries.GetAllStudents(ctx)
	if err != nil {
		return nil, err
	}

	dtoStudents := make([]studentDTO.Student, 0, len(students))
	for _, student := range students {
		dtoStudents = append(dtoStudents, studentDTO.GetStudentDTOFromDBModel(student))
	}
	return dtoStudents, nil
}

func GetAllStudentsWithCourses(ctx context.Context) ([]studentDTO.StudentWithCourseParticipationsDTO, error) {
	studentsWithCourses, err := StudentServiceSingleton.queries.GetAllStudentsWithCourseParticipations(ctx)
	if err != nil {
		return nil, err
	}

	dtoStudentsWithCourses := make([]studentDTO.StudentWithCourseParticipationsDTO, 0, len(studentsWithCourses))

	for _, studentWithCourse := range studentsWithCourses {
		dto, err := studentDTO.GetStudentWithCoursesFromDB(studentWithCourse)
		if err != nil {
			return nil, err
		}
		dtoStudentsWithCourses = append(dtoStudentsWithCourses, dto)
	}

	return dtoStudentsWithCourses, nil
}

func GetStudentByID(ctx context.Context, id uuid.UUID) (studentDTO.Student, error) {
	student, err := StudentServiceSingleton.queries.GetStudent(ctx, id)
	if err != nil {
		return studentDTO.Student{}, err
	}

	return studentDTO.GetStudentDTOFromDBModel(student), nil
}

func GetStudentByCourseParticipationID(ctx context.Context, courseParticipationID uuid.UUID) (studentDTO.Student, error) {
	student, err := StudentServiceSingleton.queries.GetStudentByCourseParticipationID(ctx, courseParticipationID)
	if err != nil {
		return studentDTO.Student{}, err
	}

	return studentDTO.GetStudentDTOFromDBModel(student), nil
}

func CreateStudent(ctx context.Context, transactionQueries *db.Queries, student studentDTO.CreateStudent) (studentDTO.Student, error) {
	queries := utils.GetQueries(transactionQueries, &StudentServiceSingleton.queries)
	createStudentParams := student.GetDBModel()

	createStudentParams.ID = uuid.New()
	createdStudent, err := queries.CreateStudent(ctx, createStudentParams)
	if err != nil {
		return studentDTO.Student{}, err
	}

	return studentDTO.GetStudentDTOFromDBModel(createdStudent), nil
}

func GetStudentByEmail(ctx context.Context, queries *db.Queries, email string) (studentDTO.Student, error) {
	student, err := queries.GetStudentByEmail(ctx, pgtype.Text{String: email, Valid: true})
	if err != nil {
		return studentDTO.Student{}, err
	}

	return studentDTO.GetStudentDTOFromDBModel(student), nil
}

func GetStudentByMatriculationNumberAndUniversityLogin(ctx context.Context, queries *db.Queries, matriculationNumber string, universityLogin string) (studentDTO.Student, error) {
	student, err := queries.GetStudentByMatriculationNumberAndUniversityLogin(ctx, db.GetStudentByMatriculationNumberAndUniversityLoginParams{
		UniversityLogin:     pgtype.Text{String: universityLogin, Valid: true},
		MatriculationNumber: pgtype.Text{String: matriculationNumber, Valid: true},
	})
	if err != nil {
		return studentDTO.Student{}, err
	}
	return studentDTO.GetStudentDTOFromDBModel(student), nil
}

func GetStudentByUniversityLogin(ctx context.Context, queries *db.Queries, universityLogin string) (studentDTO.Student, error) {
	student, err := queries.GetStudentByUniversityLogin(ctx, pgtype.Text{String: universityLogin, Valid: true})
	if err != nil {
		return studentDTO.Student{}, err
	}
	return studentDTO.GetStudentDTOFromDBModel(student), nil
}

// ResolveStudentByUniversityCredentials looks up a student by matriculation number and university login.
// If no matriculation number is provided, it falls back to university login alone.
// If the matriculation number is provided but not found, it falls back to university login —
// but only if the found student has no matriculation number yet (migration case).
// If the found student already has a different matriculation number, sql.ErrNoRows is returned.
func ResolveStudentByUniversityCredentials(ctx context.Context, queries *db.Queries, matriculationNumber string, universityLogin string) (studentDTO.Student, error) {
	if matriculationNumber != "" {
		studentObj, err := GetStudentByMatriculationNumberAndUniversityLogin(ctx, queries, matriculationNumber, universityLogin)
		if errors.Is(err, sql.ErrNoRows) {
			// Fallback: student may have been stored before matriculation number was available
			studentObj, err = GetStudentByUniversityLogin(ctx, queries, universityLogin)
			if err == nil && studentObj.MatriculationNumber != "" {
				// Found a student with a different matriculation number — treat as new user
				return studentDTO.Student{}, sql.ErrNoRows
			}
		}
		return studentObj, err
	}
	return GetStudentByUniversityLogin(ctx, queries, universityLogin)
}

func UpdateStudent(ctx context.Context, transactionQueries *db.Queries, id uuid.UUID, student studentDTO.CreateStudent) (studentDTO.Student, error) {
	queries := utils.GetQueries(transactionQueries, &StudentServiceSingleton.queries)
	updateStudentParams := student.GetDBModel()
	updateStudentParams.ID = id

	updatedStudent, err := queries.UpdateStudent(ctx, db.UpdateStudentParams(updateStudentParams))
	if err != nil {
		return studentDTO.Student{}, err
	}

	return studentDTO.GetStudentDTOFromDBModel(updatedStudent), nil
}

func CreateOrUpdateStudent(ctx context.Context, transactionQueries *db.Queries, studentInput studentDTO.CreateStudent) (studentDTO.Student, error) {
	queries := utils.GetQueries(transactionQueries, &StudentServiceSingleton.queries)

	var existingStudent studentDTO.Student
	var err error
	if !studentInput.HasUniversityAccount {
		// Student added by lecturer but without university account
		existingStudent, err = GetStudentByEmail(ctx, &queries, studentInput.Email)
	} else {
		existingStudent, err = ResolveStudentByUniversityCredentials(ctx, &queries, studentInput.MatriculationNumber, studentInput.UniversityLogin)
	}

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return CreateStudent(ctx, &queries, studentInput)
	}
	if err != nil {
		return studentDTO.Student{}, err
	}

	// student exists
	if studentInput.ID != uuid.Nil && existingStudent.ID != studentInput.ID {
		return studentDTO.Student{}, errors.New("student has wrong ID")
	} else {
		matriculationNumber := studentInput.MatriculationNumber
		if matriculationNumber == "" {
			matriculationNumber = existingStudent.MatriculationNumber
		}
		return UpdateStudent(ctx, &queries, existingStudent.ID, studentDTO.CreateStudent{
			ID:                   existingStudent.ID, // make sure the id is not overwritten
			FirstName:            studentInput.FirstName,
			LastName:             studentInput.LastName,
			Email:                existingStudent.Email,
			MatriculationNumber:  matriculationNumber,
			UniversityLogin:      studentInput.UniversityLogin,
			HasUniversityAccount: studentInput.HasUniversityAccount,
			Gender:               studentInput.Gender,
			Nationality:          studentInput.Nationality,
			StudyDegree:          studentInput.StudyDegree,
			StudyProgram:         studentInput.StudyProgram,
			CurrentSemester:      studentInput.CurrentSemester,
		})
	}
}

func SearchStudents(ctx context.Context, searchString string) ([]studentDTO.Student, error) {
	students, err := StudentServiceSingleton.queries.SearchStudents(ctx, pgtype.Text{String: searchString, Valid: true})
	if err != nil {
		return nil, err
	}

	dtoStudents := make([]studentDTO.Student, 0, len(students))
	for _, student := range students {
		dtoStudents = append(dtoStudents, studentDTO.GetStudentDTOFromDBModel(student))
	}
	return dtoStudents, nil
}

func GetStudentEnrollmentsByID(ctx context.Context, id uuid.UUID) (studentDTO.StudentEnrollmentsDTO, error) {
	studentWithEnrollments, err := StudentServiceSingleton.queries.GetStudentEnrollments(ctx, id)
	if err != nil {
		return studentDTO.StudentEnrollmentsDTO{}, err
	}

	return studentDTO.GetStudentEnrollmentsDTOFromDB(studentWithEnrollments)
}
