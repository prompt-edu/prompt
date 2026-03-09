package generator

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
)

func InitGeneratorModule(routerGroup *gin.RouterGroup, queries db.Queries) {
	setupGeneratorRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	GeneratorServiceSingleton = NewGeneratorService(queries)
}
