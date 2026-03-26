package instructorNote

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/instructorNote/instructorNoteDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

func setupInstructorNoteRouter(router *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, permissionRoleMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	instructorNoteRouter := router.Group("/instructor-notes", authMiddleware())
	instructorNoteRouter.GET("/", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer), getAllInstructorNotes)
	instructorNoteRouter.DELETE("/:note-uuid", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer), deleteInstructorNote)

	instructorNoteRouter.GET("/s/:student-uuid", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer), getInstructorNoteForStudentByID)
	instructorNoteRouter.POST("/s/:student-uuid", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer), createInstructorNoteForStudentByID)

	instructorNoteRouter.GET("/tags", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer), getAllNoteTags)
	instructorNoteRouter.POST("/tags", permissionRoleMiddleware(permissionValidation.PromptAdmin), createNoteTag)
	instructorNoteRouter.PUT("/tags/:tag-uuid", permissionRoleMiddleware(permissionValidation.PromptAdmin), updateNoteTag)
	instructorNoteRouter.DELETE("/tags/:tag-uuid", permissionRoleMiddleware(permissionValidation.PromptAdmin), deleteNoteTag)
}

// getAllInstructorNotes godoc
// @Summary Get all notes
// @Description Get all instructor notes with note versions
// @Tags instructorNotes
// @Produce json
// @Success 200 {object} []instructorNoteDTO.InstructorNote
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /instructor-notes [get]
func getAllInstructorNotes(c *gin.Context) {
	studentNotes, err := GetStudentNotes(c)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, studentNotes)
}

// getInstructorNoteForStudentByID godoc
// @Summary Get all notes for a student
// @Description Get all instructor notes with note versions for a specific student, provided the student ID
// @Tags instructorNotes
// @Produce json
// @Param student-uuid path string true "Student UUID"
// @Success 200 {object} []instructorNoteDTO.InstructorNote
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /instructor-notes/s/{student-uuid} [get]
func getInstructorNoteForStudentByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("student-uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	studentNotes, err := GetStudentNotesByID(c, id)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, studentNotes)
}

// createInstructorNoteForStudentByID godoc
// @Summary Create an instructor Note for a student
// @Description Create a new instructor note or a new edit for a specific student given its ID
// @Tags instructorNotes
// @Accept json
// @Produce json
// @Param student-uuid path string true "Student UUID"
// @Param note body instructorNoteDTO.CreateInstructorNote true "Note to create"
// @Success 200 {object} []instructorNoteDTO.InstructorNote
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /instructor-notes/s/{student-uuid} [post]
func createInstructorNoteForStudentByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("student-uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var newNote instructorNoteDTO.CreateInstructorNote
	if err := c.BindJSON(&newNote); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	userID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	authorName := utils.GetUserNameFromContext(c)
	authorEmail := utils.GetUserEmailFromContext(c)

	// validate Request
	if err := ValidateCreateNote(newNote); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	if err := ValidateReferencedNote(newNote, c, userID); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	_, err = NewStudentNote(c, id, newNote, userID, authorName, authorEmail)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	studentNotes, err := GetStudentNotesByID(c, id)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, studentNotes)
}

// deleteInstructorNote godoc
// @Summary Delete an instructor Note
// @Description Delete an instructor note by UUID
// @Tags instructorNotes
// @Produce json
// @Param note-uuid path string true "Note UUID"
// @Success 200 {object} instructorNoteDTO.InstructorNote
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /instructor-notes/{note-uuid} [delete]
func deleteInstructorNote(c *gin.Context) {
	note_id, err := uuid.Parse(c.Param("note-uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	userID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	if _, err := VerifyNoteOwnership(c, note_id, userID); err != nil {
		handleError(c, http.StatusForbidden, err)
		return
	}

	note, err := DeleteInstructorNote(c, note_id, userID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, note)
}

// createNoteTag godoc
// @Summary Create a note tag
// @Description Create a new note tag with a name and color
// @Tags instructorNotes
// @Accept json
// @Produce json
// @Param tag body instructorNoteDTO.CreateNoteTag true "Tag to create"
// @Success 200 {object} instructorNoteDTO.NoteTag
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /instructor-notes/tags [post]
func createNoteTag(c *gin.Context) {
	var newTag instructorNoteDTO.CreateNoteTag
	if err := c.BindJSON(&newTag); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	tag, err := CreateNoteTag(c, newTag)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, tag)
}

// updateNoteTag godoc
// @Summary Update a note tag
// @Description Update the name and color of an existing note tag
// @Tags instructorNotes
// @Accept json
// @Produce json
// @Param tag-uuid path string true "Tag UUID"
// @Param tag body instructorNoteDTO.UpdateNoteTag true "Updated tag data"
// @Success 200 {object} instructorNoteDTO.NoteTag
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /instructor-notes/tags/{tag-uuid} [put]
func updateNoteTag(c *gin.Context) {
	id, err := uuid.Parse(c.Param("tag-uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var updatedTag instructorNoteDTO.UpdateNoteTag
	if err := c.BindJSON(&updatedTag); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	tag, err := UpdateNoteTag(c, id, updatedTag)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, tag)
}

// deleteNoteTag godoc
// @Summary Delete a note tag
// @Description Delete a note tag by UUID
// @Tags instructorNotes
// @Produce json
// @Param tag-uuid path string true "Tag UUID"
// @Success 204
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /instructor-notes/tags/{tag-uuid} [delete]
func deleteNoteTag(c *gin.Context) {
	id, err := uuid.Parse(c.Param("tag-uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := DeleteNoteTag(c, id); err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// getAllNoteTags godoc
// @Summary Get all note tags
// @Description Get all available note tags
// @Tags instructorNotes
// @Produce json
// @Success 200 {object} []instructorNoteDTO.NoteTag
// @Failure 500 {object} utils.ErrorResponse
// @Router /instructor-notes/tags [get]
func getAllNoteTags(c *gin.Context) {
	tags, err := GetAllTags(c)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, tags)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, utils.ErrorResponse{
		Error: err.Error(),
	})
}
