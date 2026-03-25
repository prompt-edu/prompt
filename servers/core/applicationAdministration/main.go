package applicationAdministration

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/meta"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

func getScoreLevelSpecificationBytes() ([]byte, error) {
	scoreLevelSpecificationJson := meta.MetaData{}
	scoreLevelSpecificationJson["type"] = "string"
	scoreLevelSpecificationJson["enum"] = []string{"veryBad", "bad", "ok", "good", "veryGood"}
	return scoreLevelSpecificationJson.GetDBModel()
}

func InitApplicationAdministrationModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupApplicationRouter(routerGroup, keycloakTokenVerifier.KeycloakMiddleware, keycloakTokenVerifier.ApplicationMiddleware, checkAccessControlByIDWrapper)
	ApplicationServiceSingleton = &ApplicationService{
		queries: queries,
		conn:    conn,
	}

	err := initializeApplicationCoursePhaseType()
	if err != nil {
		log.Fatal("failed to init application phase type: ", err)
	}
}

// initializes the handler func with CheckCoursePermissions
func checkAccessControlByIDWrapper(allowedRoles ...string) gin.HandlerFunc {
	return permissionValidation.CheckAccessControlByID(permissionValidation.CheckCoursePhasePermission, "coursePhaseID", allowedRoles...)
}

func initializeApplicationCoursePhaseType() error {
	// check if the application module exists in the types
	ctx := context.Background()
	exists, err := ApplicationServiceSingleton.queries.TestApplicationPhaseTypeExists(ctx)
	if err != nil {
		log.Error("failed to check if application phase type exists: ", err)
		return err
	}

	if !exists {
		// 1.) start transaction
		tx, err := ApplicationServiceSingleton.conn.Begin(ctx)
		if err != nil {
			return err
		}
		defer sdkUtils.DeferRollback(tx, ctx)
		qtx := ApplicationServiceSingleton.queries.WithTx(tx)

		// 2.) create the application module
		newApplicationPhaseType := db.CreateCoursePhaseTypeParams{
			ID:           uuid.New(),
			Name:         "Application",
			InitialPhase: true,
			BaseUrl:      "core",
		}
		err = qtx.CreateCoursePhaseType(ctx, newApplicationPhaseType)
		if err != nil {
			log.Error("failed to create application module: ", err)
		}

		// 3.) create the provided output meta data
		// 3.1 Application Score
		scoreSpecificationJson := meta.MetaData{}
		scoreSpecificationJson["type"] = "integer"
		scoreSpecificationBytes, err := scoreSpecificationJson.GetDBModel()
		if err != nil {
			log.Error("failed to parse score specification")
			return err
		}

		scoreLevelSpecificationBytes, err := getScoreLevelSpecificationBytes()
		if err != nil {
			log.Error("failed to parse score level specification")
			return err
		}

		newProvidedOutput := db.CreateCoursePhaseTypeProvidedOutputParams{
			ID:                uuid.New(),
			CoursePhaseTypeID: newApplicationPhaseType.ID,
			DtoName:           "score",
			Specification:     scoreSpecificationBytes,
			VersionNumber:     1,
			EndpointPath:      "core",
		}
		err = qtx.CreateCoursePhaseTypeProvidedOutput(ctx, newProvidedOutput)
		if err != nil {
			log.Error("failed to create required score input: ", err)
			return err
		}

		newProvidedScoreLevelOutput := db.CreateCoursePhaseTypeProvidedOutputParams{
			ID:                uuid.New(),
			CoursePhaseTypeID: newApplicationPhaseType.ID,
			DtoName:           "scoreLevel",
			Specification:     scoreLevelSpecificationBytes,
			VersionNumber:     1,
			EndpointPath:      "core",
		}
		err = qtx.CreateCoursePhaseTypeProvidedOutput(ctx, newProvidedScoreLevelOutput)
		if err != nil {
			log.Error("failed to create score level output: ", err)
			return err
		}

		// 3.2 Application Answers
		err = qtx.InsertCourseProvidedApplicationAnswers(ctx, newApplicationPhaseType.ID)
		if err != nil {
			log.Error("failed to create required application answers: ", err)
			return err
		}

		// 3.3 Additional Scores
		err = qtx.InsertCourseProvidedAdditionalScores(ctx, newApplicationPhaseType.ID)
		if err != nil {
			log.Error("failed to create required additional scores: ", err)
			return err
		}

		// 4.) commit the transaction
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

	} else {
		log.Debug("application module already exists")
	}

	return nil
}
