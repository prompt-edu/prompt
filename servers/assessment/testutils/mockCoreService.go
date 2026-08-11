package testutils

import (
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MockCoursePhaseResponse struct {
	CourseID uuid.UUID `json:"courseID"`
}

type MockCourseResponse struct {
	Name        string `json:"name"`
	SemesterTag string `json:"semesterTag"`
}

// Team layout served by the mock core for course phase data resolution. Peer and tutor
// evaluation targets are validated against it, so tests can rely on these fixed IDs.
var (
	MockTeamID      = uuid.MustParse("7ea11111-1111-1111-1111-111111111111")
	MockTeamMembers = []uuid.UUID{
		uuid.MustParse("01234567-1234-1234-1234-123456789012"),
		uuid.MustParse("02234567-1234-1234-1234-123456789012"),
		uuid.MustParse("03234567-1234-1234-1234-123456789012"),
	}
	MockTeamTutors = []uuid.UUID{
		uuid.MustParse("0a234567-1234-1234-1234-123456789012"),
	}
	MockOutsiderParticipationID = uuid.MustParse("0f234567-1234-1234-1234-123456789012")
)

func mockPersons(ids []uuid.UUID) []map[string]any {
	persons := make([]map[string]any, 0, len(ids))
	for i, id := range ids {
		persons = append(persons, map[string]any{
			"id":        id.String(),
			"firstName": "Person",
			"lastName":  string(rune('A' + i)),
		})
	}
	return persons
}

func SetupMockCoreService() (*httptest.Server, func()) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock course phase data resolution, used to resolve teams for a course phase
	router.GET("/api/course_phases/:id/course_phase_data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"prevData": gin.H{
				"teams": []map[string]any{
					{
						"id":      MockTeamID.String(),
						"name":    "Mock Team",
						"members": mockPersons(MockTeamMembers),
						"tutors":  mockPersons(MockTeamTutors),
					},
				},
			},
			"resolutions": []any{},
		})
	})

	// Mock course phases endpoint
	router.GET("/api/course_phases/:id", func(c *gin.Context) {
		phaseID := c.Param("id")

		// Map known test phase IDs to course IDs
		var courseID uuid.UUID
		switch phaseID {
		case "10000000-0000-0000-0000-000000000001":
			courseID = uuid.MustParse("20000000-0000-0000-0000-000000000001")
		case "10000000-0000-0000-0000-000000000002":
			courseID = uuid.MustParse("20000000-0000-0000-0000-000000000001")
		case "10000000-0000-0000-0000-000000000003":
			courseID = uuid.MustParse("20000000-0000-0000-0000-000000000001")
		case "10000000-0000-0000-0000-000000000004":
			courseID = uuid.MustParse("20000000-0000-0000-0000-000000000001")
		case "4179d58a-d00d-4fa7-94a5-397bc69fab02":
			courseID = uuid.MustParse("30000000-0000-0000-0000-000000000001")
		default:
			// Default course ID for any other phase
			courseID = uuid.MustParse("90000000-0000-0000-0000-000000000001")
		}

		response := MockCoursePhaseResponse{
			CourseID: courseID,
		}
		c.JSON(http.StatusOK, response)
	})

	router.GET("/api/courses/:id", func(c *gin.Context) {
		courseID := c.Param("id")

		var name, semesterTag string
		switch courseID {
		case "20000000-0000-0000-0000-000000000001":
			name = "ios"
			semesterTag = "2526"
		case "30000000-0000-0000-0000-000000000001":
			name = "testcourse"
			semesterTag = "2425"
		default:
			name = "defaultcourse"
			semesterTag = "2526"
		}

		response := MockCourseResponse{
			Name:        name,
			SemesterTag: semesterTag,
		}
		c.JSON(http.StatusOK, response)
	})

	server := httptest.NewServer(router)

	oldCoreHost := os.Getenv("SERVER_CORE_HOST")
	_ = os.Setenv("SERVER_CORE_HOST", server.URL)

	cleanup := func() {
		server.Close()
		if oldCoreHost != "" {
			_ = os.Setenv("SERVER_CORE_HOST", oldCoreHost)
		} else {
			_ = os.Unsetenv("SERVER_CORE_HOST")
		}
	}

	return server, cleanup
}

func GetMockCourseIdentifier(coursePhaseID uuid.UUID) string {
	switch coursePhaseID.String() {
	case "10000000-0000-0000-0000-000000000001",
		"10000000-0000-0000-0000-000000000002",
		"10000000-0000-0000-0000-000000000003",
		"10000000-0000-0000-0000-000000000004":
		return "mockedios2526"
	case "4179d58a-d00d-4fa7-94a5-397bc69fab02":
		return "testcourse2425"
	default:
		return "defaultcourse2526"
	}
}
