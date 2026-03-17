package course

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/course/courseDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CourseTestSuite struct {
	suite.Suite
	router        *gin.Engine
	ctx           context.Context
	cleanup       func()
	courseService CourseService
}

func (suite *CourseTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up PostgreSQL container
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/course_test.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.courseService = CourseService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	CourseServiceSingleton = &suite.courseService
	suite.router = gin.Default()
}

func (suite *CourseTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *CourseTestSuite) TestValidateCreateCourse() {
	tests := []struct {
		name          string
		input         courseDTO.CreateCourse
		expectedError string
	}{
		{
			name: "valid course",
			input: courseDTO.CreateCourse{
				Name:                "Valid Course",
				StartDate:           pgtype.Date{Valid: true, Time: time.Now()},
				EndDate:             pgtype.Date{Valid: true, Time: time.Now().Add(24 * time.Hour)},
				SemesterTag:         pgtype.Text{String: "WS2024", Valid: true},
				RestrictedData:      meta.MetaData{"key": "value"},
				StudentReadableData: meta.MetaData{"differentKey": "value"},
				CourseType:          "practical course",
				Ects:                pgtype.Int4{Int32: 10, Valid: true},
			},
			expectedError: "",
		},
		{
			name: "missing course name",
			input: courseDTO.CreateCourse{
				Name:                "",
				StartDate:           pgtype.Date{Valid: true, Time: time.Now()},
				EndDate:             pgtype.Date{Valid: true, Time: time.Now().Add(24 * time.Hour)},
				SemesterTag:         pgtype.Text{String: "WS2024", Valid: true},
				RestrictedData:      meta.MetaData{"key": "value"},
				StudentReadableData: meta.MetaData{"differentKey": "value"},
				CourseType:          "practical course",
				Ects:                pgtype.Int4{Int32: 10, Valid: true},
			},
			expectedError: "course name is required",
		},
		{
			name: "start date after end date",
			input: courseDTO.CreateCourse{
				Name:                "Invalid Date Course",
				StartDate:           pgtype.Date{Valid: true, Time: time.Now().Add(24 * time.Hour)},
				EndDate:             pgtype.Date{Valid: true, Time: time.Now()},
				SemesterTag:         pgtype.Text{String: "WS2024", Valid: true},
				RestrictedData:      meta.MetaData{"key": "value"},
				StudentReadableData: meta.MetaData{"differentKey": "value"},
				CourseType:          "practical course",
				Ects:                pgtype.Int4{Int32: 10, Valid: true},
			},
			expectedError: "start date must be before end date",
		},
		{
			name: "missing semester tag",
			input: courseDTO.CreateCourse{
				Name:                "Invalid Semester Tag",
				StartDate:           pgtype.Date{Valid: true, Time: time.Now()},
				EndDate:             pgtype.Date{Valid: true, Time: time.Now().Add(24 * time.Hour)},
				SemesterTag:         pgtype.Text{String: "", Valid: false},
				RestrictedData:      meta.MetaData{"key": "value"},
				StudentReadableData: meta.MetaData{"differentKey": "value"},
				CourseType:          "practical course",
				Ects:                pgtype.Int4{Int32: 10, Valid: true},
			},
			expectedError: "semester tag is required",
		},
		{
			name: "missing course type",
			input: courseDTO.CreateCourse{
				Name:                "Invalid Course Type",
				StartDate:           pgtype.Date{Valid: true, Time: time.Now()},
				EndDate:             pgtype.Date{Valid: true, Time: time.Now().Add(24 * time.Hour)},
				SemesterTag:         pgtype.Text{String: "ios2425", Valid: true},
				RestrictedData:      meta.MetaData{"key": "value"},
				StudentReadableData: meta.MetaData{"differentKey": "value"},
				CourseType:          "",
				Ects:                pgtype.Int4{Int32: 10, Valid: true},
			},
			expectedError: "course type is required",
		},
		{
			name: "short description too long",
			input: courseDTO.CreateCourse{
				Name:                "Course With Long Short Description",
				StartDate:           pgtype.Date{Valid: true, Time: time.Now()},
				EndDate:             pgtype.Date{Valid: true, Time: time.Now().Add(24 * time.Hour)},
				SemesterTag:         pgtype.Text{String: "ios2425", Valid: true},
				RestrictedData:      meta.MetaData{"key": "value"},
				StudentReadableData: meta.MetaData{"differentKey": "value"},
				CourseType:          "practical course",
				Ects:                pgtype.Int4{Int32: 10, Valid: true},
				ShortDescription:    pgtype.Text{String: strings.Repeat("a", 260), Valid: true},
			},
			expectedError: "short description must be 255 characters or fewer",
		},
		{
			name: "long description too long",
			input: courseDTO.CreateCourse{
				Name:                "Course With Extremely Long Description",
				StartDate:           pgtype.Date{Valid: true, Time: time.Now()},
				EndDate:             pgtype.Date{Valid: true, Time: time.Now().Add(24 * time.Hour)},
				SemesterTag:         pgtype.Text{String: "ios2425", Valid: true},
				RestrictedData:      meta.MetaData{"key": "value"},
				StudentReadableData: meta.MetaData{"differentKey": "value"},
				CourseType:          "practical course",
				Ects:                pgtype.Int4{Int32: 10, Valid: true},
				LongDescription:     pgtype.Text{String: strings.Repeat("b", 6000), Valid: true},
			},
			expectedError: "long description must be 5000 characters or fewer",
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			err := validateCreateCourse(tt.input)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

func (suite *CourseTestSuite) TestValidateUpdateCourseData() {
	tests := []struct {
		name          string
		input         courseDTO.UpdateCourseData
		expectedError string
	}{
		{
			name: "valid short description",
			input: courseDTO.UpdateCourseData{
				ShortDescription: pgtype.Text{String: "Concise summary", Valid: true},
			},
			expectedError: "",
		},
		{
			name: "short description too long",
			input: courseDTO.UpdateCourseData{
				ShortDescription: pgtype.Text{String: strings.Repeat("x", 300), Valid: true},
			},
			expectedError: "short description must be 255 characters or fewer",
		},
		{
			name: "long description too long",
			input: courseDTO.UpdateCourseData{
				LongDescription: pgtype.Text{String: strings.Repeat("y", 6000), Valid: true},
			},
			expectedError: "long description must be 5000 characters or fewer",
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			err := validateUpdateCourseData(tt.input)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

func (suite *CourseTestSuite) TestValidateUpdateCourseOrder() {
	// set up CoursePhaseService
	coursePhase.InitCoursePhaseModule(suite.router.Group("/api"), suite.courseService.queries, suite.courseService.conn)

	tests := []struct {
		name          string
		courseID      uuid.UUID
		orderedPhases []uuid.UUID
		expectedError string
	}{
		{
			name:     "valid course phase order",
			courseID: uuid.MustParse("3f42d322-e5bf-4faa-b576-51f2cab14c2e"),
			orderedPhases: []uuid.UUID{
				uuid.MustParse("3d1f3b00-87f3-433b-a713-178c4050411b"),
				uuid.MustParse("92bb0532-39e5-453d-bc50-fa61ea0128b2"),
			},
			expectedError: "",
		},
		// Example of a failing test:
		// {
		// 	name:     "invalid course ID in phase",
		// 	courseID: uuid.MustParse("3f42d322-e5bf-4faa-b576-51f2cab14c2e"),
		// 	orderedPhases: []uuid.UUID{
		// 		uuid.MustParse("3d1f3b00-87f3-433b-a713-178c4050411c"),
		// 	},
		// 	expectedError: "course id must be the same for all course phases",
		// },
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			// Convert orderedPhases to a slice of CoursePhaseGraph entries
			var phaseGraph []courseDTO.CoursePhaseGraph
			for _, phaseID := range tt.orderedPhases {
				phaseGraph = append(phaseGraph, courseDTO.CoursePhaseGraph{
					FromCoursePhaseID: uuid.Nil, // or another valid UUID if needed
					ToCoursePhaseID:   phaseID,
				})
			}

			err := validateUpdateCourseOrder(context.Background(), tt.courseID, phaseGraph)
			if tt.expectedError == "" {
				assert.NoError(t, err, "Expected no error, got %v", err)
			} else {
				assert.Error(t, err, "Expected an error but got none")
				assert.EqualError(t, err, tt.expectedError, "Error message did not match")
			}
		})
	}
}

func TestCourseTestSuite(t *testing.T) {
	suite.Run(t, new(CourseTestSuite))
}
