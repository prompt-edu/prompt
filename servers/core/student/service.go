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

	dtoStudentsWithCourses := make( []studentDTO.StudentWithCourseParticipationsDTO, 0, len(studentsWithCourses))

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

func GetStudentByEmail(ctx context.Context, email string) (studentDTO.Student, error) {
	student, err := StudentServiceSingleton.queries.GetStudentByEmail(ctx, pgtype.Text{String: email, Valid: true})
	if err != nil {
		return studentDTO.Student{}, err
	}

	return studentDTO.GetStudentDTOFromDBModel(student), nil
}

func GetStudentByMatriculationNumberAndUniversityLogin(ctx context.Context, matriculation_number string, universityLogin string) (studentDTO.Student, error) {
	student, err := StudentServiceSingleton.queries.GetStudentByMatriculationNumberAndUniversityLogin(ctx, db.GetStudentByMatriculationNumberAndUniversityLoginParams{
		UniversityLogin:     pgtype.Text{String: universityLogin, Valid: true},
		MatriculationNumber: pgtype.Text{String: matriculation_number, Valid: true},
	})
	if err != nil {
		return studentDTO.Student{}, err
	}
	return studentDTO.GetStudentDTOFromDBModel(student), nil
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

func CreateOrUpdateStudent(ctx context.Context, transactionQueries *db.Queries, studentObj studentDTO.CreateStudent) (studentDTO.Student, error) {
	queries := utils.GetQueries(transactionQueries, &StudentServiceSingleton.queries)

	var studentByEmail studentDTO.Student
	var err error
	if !studentObj.HasUniversityAccount {
		// Student added by lecturer but without university account
		studentByEmail, err = GetStudentByEmail(ctx, studentObj.Email)
	} else {
		studentByEmail, err = GetStudentByMatriculationNumberAndUniversityLogin(ctx, studentObj.MatriculationNumber, studentObj.UniversityLogin)
	}

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return CreateStudent(ctx, &queries, studentObj)
	}
	if err != nil {
		return studentDTO.Student{}, err
	}

	// student exists
	if studentObj.ID != uuid.Nil && studentByEmail.ID != studentObj.ID {
		return studentDTO.Student{}, errors.New("student has wrong ID")
	} else {
		return UpdateStudent(ctx, &queries, studentByEmail.ID, studentDTO.CreateStudent{
			ID:                   studentByEmail.ID, // make sure the id is not overwritten
			FirstName:            studentObj.FirstName,
			LastName:             studentObj.LastName,
			Email:                studentByEmail.Email,
			MatriculationNumber:  studentObj.MatriculationNumber,
			UniversityLogin:      studentObj.UniversityLogin,
			HasUniversityAccount: studentObj.HasUniversityAccount,
			Gender:               studentObj.Gender,
			Nationality:          studentObj.Nationality,
			StudyDegree:          studentObj.StudyDegree,
			StudyProgram:         studentObj.StudyProgram,
			CurrentSemester:      studentObj.CurrentSemester,
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

  return  studentDTO.GetStudentEnrollmentsDTOFromDB(studentWithEnrollments)
}
