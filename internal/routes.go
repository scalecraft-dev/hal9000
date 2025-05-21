package internal

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, incidentService *IncidentService, slackService *SlackService) {
	router.Use(func(c *gin.Context) {
		c.Set("slackApi", slackService.GetClient())
		c.Next()
	})

	router.GET("/health", HealthHandler(slackService))
	router.POST("/incident", IncidentHandler(incidentService))
	router.POST("/interaction", InteractionHandler(incidentService))
}
