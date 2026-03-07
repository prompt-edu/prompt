package generator

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/ls1intum/prompt-sdk"
	db "github.com/ls1intum/prompt2/servers/certificate/db/sqlc"
)

func InitGeneratorModule(routerGroup *gin.RouterGroup, queries db.Queries) {
	setupGeneratorRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	GeneratorServiceSingleton = NewGeneratorService(queries)
}
