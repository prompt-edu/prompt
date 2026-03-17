package privacy

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/privacy/service"
)

func setupPrivacyRouter(router *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, premissionRoleMiddleware func(allowedRoles ...string) gin.HandlerFunc) {

  privacyRouter := router.Group("/privacy", authMiddleware())

  privacyRouter.POST("/student-data-export", premissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseEditor, permissionValidation.CourseLecturer, permissionValidation.CourseStudent), studentDataExport)

}


// studentDataExport exports all student related data from core and all microservices.
//
// @Summary Export all student related data
// @Description Get an export of all student data from core and all microservices
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.ErrorResponse
// @Router /privacy/student-data-export [post]
func studentDataExport(c *gin.Context) {

  subjectIdentifiers, errSI := service.GetSubjectIdentifiers(c)
  if errSI != nil {
    c.JSON(500, "err");
    return
  }
  fmt.Printf("subjectIdentifiers: %+v\n", subjectIdentifiers)

  service.GetSubjectData(subjectIdentifiers)

  c.JSON(200, subjectIdentifiers);


  // from core:
  //
  // main:
  // - student record export
  // - all course phase participations
  // - - do we export restricted_data?
  // - all course participations
  // - all instructor notes
  // application:
  // - 
  //
  // - do we export course info? for convenience?
  // 

}
