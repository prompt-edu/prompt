package coursePhaseType

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

func InitCoursePhaseTypeModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool, isDevEnvironment bool) {

	setupCoursePhaseTypeRouter(routerGroup)
	CoursePhaseTypeServiceSingleton = &CoursePhaseTypeService{
		queries:          queries,
		conn:             conn,
		isDevEnvironment: isDevEnvironment,
	}

	// initialize course phase types
	err := initInterview()
	if err != nil {
		log.Fatal("failed to init interview phase type: ", err)
	}

	err = initMatching()
	if err != nil {
		log.Fatal("failed to init matching phase type: ", err)
	}

	err = initIntroCourseDeveloper()
	if err != nil {
		log.Fatal("failed to init intro course developer phase type: ", err)
	}

	err = initDevOpsChallenge()
	if err != nil {
		log.Fatal("failed to init dev ops challenge phase type: ", err)
	}

	err = initAssessment()
	if err != nil {
		log.Fatal("failed to init assessment phase type: ", err)
	}

	err = initTeamAllocation()
	if err != nil {
		log.Fatal("failed to init team allocation phase type: ", err)
	}

	err = initSelfTeamAllocation()
	if err != nil {
		log.Fatal("failed to init self team allocation phase type: ", err)
	}

	err = initCertificate()
	if err != nil {
		log.Fatal("failed to init certificate phase type: ", err)
	}
}
