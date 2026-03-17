package student

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

func setupStudentRouter(router *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, permissionRoleMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	student := router.Group("/students", authMiddleware())
	student.GET("/with-courses", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer), getAllStudentsWithCourses)
	student.GET("/search/:searchString", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer), searchStudents)
	student.GET("/:uuid", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer), getStudentByID)
	student.GET("/", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer), getAllStudents)
	student.POST("/", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer), createStudent)
	student.PUT("/:uuid", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer), updateStudent)
	student.GET("/:uuid/enrollments", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer), getStudentEnrollments)
}

// getAllStudents godoc
// @Summary Get all students
// @Description Get a list of all students
// @Tags students
// @Produce json
// @Success 200 {array} studentDTO.Student
// @Failure 500 {object} utils.ErrorResponse
// @Router /students/ [get]
func getAllStudents(c *gin.Context) {
	students, err := GetAllStudents(c)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, students)
}

// getAllStudentsWithCourses() godoc
// @Summary Get all students with courses
// @Description Get a list of all students with the property 'courses' a list of courses that the student is taking part of or was
// @Tags students
// @Produce json
// @Success 200 {array} studentDTO.StudentWithCourseParticipationsDTO
// @Failure 500 {object} utils.ErrorResponse
// @Router /students/with-courses [get]
func getAllStudentsWithCourses(c *gin.Context) {
	studentsWithCourses, err := GetAllStudentsWithCourses(c)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, studentsWithCourses)
}

// getStudentByID godoc
// @Summary Get student by ID
// @Description Get a student by UUID
// @Tags students
// @Produce json
// @Param uuid path string true "Student UUID"
// @Success 200 {object} studentDTO.Student
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /students/{uuid} [get]
func getStudentByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	student, err := GetStudentByID(c, id)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, student)
}

// createStudent godoc
// @Summary Create a student
// @Description Create a new student
// @Tags students
// @Accept json
// @Produce json
// @Param student body studentDTO.CreateStudent true "Student to create"
// @Success 201 {object} studentDTO.Student
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /students/ [post]
func createStudent(c *gin.Context) {
	var newStudent studentDTO.CreateStudent
	if err := c.BindJSON(&newStudent); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// validate student
	if err := Validate(newStudent); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	student, err := CreateStudent(c, nil, newStudent)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, student)
}

// updateStudent godoc
// @Summary Update a student
// @Description Update an existing student by UUID
// @Tags students
// @Accept json
// @Produce json
// @Param uuid path string true "Student UUID"
// @Param student body studentDTO.CreateStudent true "Student to update"
// @Success 200 {object} studentDTO.Student
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /students/{uuid} [put]
func updateStudent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var updateStudent studentDTO.CreateStudent
	if err := c.BindJSON(&updateStudent); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// make sure that the UUID matches
	if id != updateStudent.ID {
		handleError(c, http.StatusBadRequest, errors.New("UUID in URL does not match UUID in body"))
		return
	}

	// validate student
	if err := Validate(updateStudent); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	student, err := UpdateStudent(c, nil, id, updateStudent)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, student)
}

// searchStudents godoc
// @Summary Search students
// @Description Search students by a search string
// @Tags students
// @Produce json
// @Param searchString path string true "Search string"
// @Success 200 {array} studentDTO.Student
// @Failure 500 {object} utils.ErrorResponse
// @Router /students/search/{searchString} [get]
func searchStudents(c *gin.Context) {
	searchString := c.Param("searchString")

	students, err := SearchStudents(c, searchString)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, students)
}

// getStudentEnrollments godoc
// @Summary Get student enrollments by ID
// @Description Get all of a students enrollments, provide student UUID
// @Tags students
// @Produce json
// @Param uuid path string true "Student UUID"
// @Success 200 {object} studentDTO.StudentEnrollmentsDTO
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /students/{uuid}/enrollments [get]
func getStudentEnrollments(c *gin.Context) {
	id, parseErr := uuid.Parse(c.Param("uuid"))
	if parseErr != nil {
		handleError(c, http.StatusBadRequest, parseErr)
		return
	}
  studentEnrollments, err := GetStudentEnrollmentsByID(c, id)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, studentEnrollments)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, utils.ErrorResponse{
		Error: err.Error(),
	})
}
