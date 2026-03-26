package participants

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
)

func InitParticipantsModule(routerGroup *gin.RouterGroup, queries db.Queries) {
	setupParticipantsRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	ParticipantsServiceSingleton = NewParticipantsService(queries)
}
