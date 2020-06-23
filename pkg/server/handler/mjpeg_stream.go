package handler

import (
	"github.com/gin-gonic/gin"
	"org.penitence/face-service/pkg/server/engineer"
	"strconv"
)

func CreateStream() gin.HandlerFunc {
	return func(context *gin.Context) {
		idStr := context.Param("deviceId")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			id = 0
		}
		engineer.GetInstance().HttpHandler(id).ServeHTTP(context.Writer, context.Request)
		engineer.GetInstance().Release()
	}
	//return engineer.GetInstance().HttpHandler(deviceId)
}
