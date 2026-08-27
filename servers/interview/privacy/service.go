package privacy

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	"github.com/prompt-edu/prompt/servers/interview/interviewReview/interviewReviewDTO"
)

type PrivacyService struct {
	Queries db.Queries
	conn    *pgxpool.Pool
}

var PrivacyServiceSingleton *PrivacyService

type interviewReviewExport struct {
	CoursePhaseID         uuid.UUID                            `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID                            `json:"courseParticipationID"`
	Score                 *int32                               `json:"score,omitempty"`
	Interviewer           string                               `json:"interviewer"`
	InterviewAnswers      []interviewReviewDTO.InterviewAnswer `json:"interviewAnswers"`
	CreatedAt             *time.Time                           `json:"createdAt,omitempty"`
	UpdatedAt             *time.Time                           `json:"updatedAt,omitempty"`
}

func toInterviewReviewExports(reviews []db.InterviewReview) []interviewReviewExport {
	exports := make([]interviewReviewExport, 0, len(reviews))
	for _, review := range reviews {
		dto := interviewReviewDTO.GetInterviewReviewFromDB(review)
		export := interviewReviewExport{
			CoursePhaseID:         review.CoursePhaseID,
			CourseParticipationID: dto.CourseParticipationID,
			Score:                 dto.Score,
			Interviewer:           dto.Interviewer,
			InterviewAnswers:      dto.InterviewAnswers,
		}
		if review.CreatedAt.Valid {
			export.CreatedAt = &review.CreatedAt.Time
		}
		if review.UpdatedAt.Valid {
			export.UpdatedAt = &review.UpdatedAt.Time
		}
		exports = append(exports, export)
	}
	return exports
}

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	exp.AddJSON("Interview Assignments", "interview_assignments.json", func() (any, error) {
		return PrivacyServiceSingleton.Queries.GetInterviewAssignmentsByParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Interview Reviews", "interview_reviews.json", func() (any, error) {
		reviews, err := PrivacyServiceSingleton.Queries.GetInterviewReviewsByParticipationIDs(c, subject.CourseParticipationIDs)
		if err != nil {
			return nil, err
		}
		return toInterviewReviewExports(reviews), nil
	})
	return nil
}

func PrivacyDataDeletionHandler(c *gin.Context, subject sdkAuth.SubjectIdentifiers) error {
	tx, err := PrivacyServiceSingleton.conn.Begin(c)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, c)
	qtx := PrivacyServiceSingleton.Queries.WithTx(tx)

	if err := qtx.DeleteInterviewAssignmentsByParticipationIDs(c, subject.CourseParticipationIDs); err != nil {
		return err
	}
	if err := qtx.DeleteInterviewReviewsByParticipationIDs(c, subject.CourseParticipationIDs); err != nil {
		return err
	}
	return tx.Commit(c)
}
