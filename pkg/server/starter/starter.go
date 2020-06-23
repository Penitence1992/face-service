package starter

import (
	"github.com/gin-gonic/gin"
	"org.penitence/face-service/pkg/server"
	"org.penitence/face-service/pkg/server/handler"
)

func StartDaemon() {
	httpServer := server.CreateServer(8080, "", nil)
	httpServer.RegisterRoute(func(engine *gin.Engine) {
		engine.GET("/:deviceId", handler.CreateStream())
		//database := engine.Group("/databases")
		//database.GET("", router.FindDatabasePage)
		//database.POST("", router.CreateDatabase)
		//kube := engine.Group("/kube")
		//kube.GET("/:namespace/:pod/logs", router.GetLogs)
	})
	httpServer.StartListen()
}
