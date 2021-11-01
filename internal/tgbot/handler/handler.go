package handler

import "github.com/gin-gonic/gin"

type Handler struct {
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	api := router.Group("/api")
	{
		api.GET("/logs", sendLog)
		api.POST("/logs", sendLogWithFile)
	}

	return router
}
