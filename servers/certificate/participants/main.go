package participants

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/ls1intum/prompt-sdk"
	db "github.com/ls1intum/prompt2/servers/certificate/db/sqlc"
)

func InitParticipantsModule(routerGroup *gin.RouterGroup, queries db.Queries) {
	setupParticipantsRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	ParticipantsServiceSingleton = NewParticipantsService(queries)
}
