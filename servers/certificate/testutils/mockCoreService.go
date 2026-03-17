package testutils

import (
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SetupMockCoreService() (*httptest.Server, func()) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// We need to create the server first to know its URL for resolution references
	var serverURL string

	// Mock course phases endpoint
	router.GET("/api/course_phases/:id", func(c *gin.Context) {
		phaseID := c.Param("id")

		var courseID uuid.UUID
		switch phaseID {
		case "10000000-0000-0000-0000-000000000001":
			courseID = uuid.MustParse("20000000-0000-0000-0000-000000000001")
		default:
			courseID = uuid.MustParse("90000000-0000-0000-0000-000000000001")
		}

		c.JSON(http.StatusOK, gin.H{
			"id":       phaseID,
			"courseId": courseID.String(),
			"course": gin.H{
				"id":   courseID.String(),
				"name": "Test Course",
			},
		})
	})

	// Mock course phase data endpoint (used by prompt-sdk resolution system)
	// Returns prevData + resolutions pointing to our mock team service endpoint
	router.GET("/api/course_phases/:id/course_phase_data", func(c *gin.Context) {
		phaseID := c.Param("id")

		c.JSON(http.StatusOK, gin.H{
			"prevData": gin.H{},
			"resolutions": []gin.H{
				{
					"dtoName":       "teams",
					"baseURL":       serverURL,
					"endpointPath":  "/team",
					"coursePhaseID": phaseID,
				},
			},
		})
	})

	// Mock team allocation service endpoint (called by SDK resolution)
	// URL pattern: /course_phase/{coursePhaseID}/team
	router.GET("/course_phase/:coursePhaseID/team", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"teams": []gin.H{
				{
					"id":   "60000000-0000-0000-0000-000000000001",
					"name": "BMW",
					"members": []gin.H{
						{"id": "30000000-0000-0000-0000-000000000001", "firstName": "John", "lastName": "Doe"},
					},
					"tutors": []gin.H{},
				},
				{
					"id":   "60000000-0000-0000-0000-000000000002",
					"name": "Siemens",
					"members": []gin.H{
						{"id": "30000000-0000-0000-0000-000000000002", "firstName": "Jane", "lastName": "Smith"},
					},
					"tutors": []gin.H{},
				},
			},
		})
	})

	// Mock participations endpoint
	// Returns both formats: "participations" (SDK resolution format) and
	// "coursePhaseParticipations" (direct API format) so both code paths work
	router.GET("/api/course_phases/:id/participations", func(c *gin.Context) {
		phaseID := c.Param("id")

		participationData := []gin.H{
			{
				"coursePhaseID":         phaseID,
				"courseParticipationID": "50000000-0000-0000-0000-000000000001",
				"passStatus":           "not_assessed",
				"prevData":             gin.H{},
				"restrictedData":       gin.H{},
				"studentReadableData":  gin.H{},
				"student": gin.H{
					"id":        "30000000-0000-0000-0000-000000000001",
					"firstName": "John",
					"lastName":  "Doe",
					"email":     "john.doe@example.com",
				},
			},
			{
				"coursePhaseID":         phaseID,
				"courseParticipationID": "50000000-0000-0000-0000-000000000002",
				"passStatus":           "not_assessed",
				"prevData":             gin.H{},
				"restrictedData":       gin.H{},
				"studentReadableData":  gin.H{},
				"student": gin.H{
					"id":        "30000000-0000-0000-0000-000000000002",
					"firstName": "Jane",
					"lastName":  "Smith",
					"email":     "jane.smith@example.com",
				},
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"participations": participationData,
			"resolutions": []gin.H{
				{
					"dtoName":       "teamAllocation",
					"baseURL":       serverURL,
					"endpointPath":  "/team-allocation",
					"coursePhaseID": phaseID,
				},
			},
		})
	})

	// Mock team allocation resolution endpoint (returns per-participation team assignments)
	// URL pattern: /course_phase/{coursePhaseID}/team-allocation
	// Returns array of [{courseParticipationID, teamAllocation}]
	router.GET("/course_phase/:coursePhaseID/team-allocation", func(c *gin.Context) {
		c.JSON(http.StatusOK, []gin.H{
			{
				"courseParticipationID": "50000000-0000-0000-0000-000000000001",
				"teamAllocation":       "60000000-0000-0000-0000-000000000001",
			},
			{
				"courseParticipationID": "50000000-0000-0000-0000-000000000002",
				"teamAllocation":       "60000000-0000-0000-0000-000000000002",
			},
		})
	})

	// Mock single participation endpoint
	router.GET("/api/course_phases/:id/participations/:studentID", func(c *gin.Context) {
		studentID := c.Param("studentID")

		c.JSON(http.StatusOK, gin.H{
			"id":                    "40000000-0000-0000-0000-000000000001",
			"courseParticipationId": "50000000-0000-0000-0000-000000000001",
			"student": gin.H{
				"id":        studentID,
				"firstName": "John",
				"lastName":  "Doe",
				"email":     "john.doe@example.com",
			},
		})
	})

	// Mock students endpoint (returns all students in a course phase by their core student ID)
	router.GET("/api/course_phases/:id/participations/students", func(c *gin.Context) {
		c.JSON(http.StatusOK, []gin.H{
			{
				"id":        "30000000-0000-0000-0000-000000000001",
				"firstName": "John",
				"lastName":  "Doe",
				"email":     "john.doe@example.com",
			},
			{
				"id":        "30000000-0000-0000-0000-000000000002",
				"firstName": "Jane",
				"lastName":  "Smith",
				"email":     "jane.smith@example.com",
			},
		})
	})

	// Mock self participation endpoint (returns the student's own participation with core student ID)
	router.GET("/api/course_phases/:id/participations/self", func(c *gin.Context) {
		phaseID := c.Param("id")

		c.JSON(http.StatusOK, gin.H{
			"coursePhaseID":         phaseID,
			"courseParticipationID": "50000000-0000-0000-0000-000000000001",
			"student": gin.H{
				"id":        "30000000-0000-0000-0000-000000000001",
				"firstName": "John",
				"lastName":  "Doe",
				"email":     "john.doe@example.com",
			},
		})
	})

	server := httptest.NewServer(router)
	serverURL = server.URL

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
