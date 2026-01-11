package http

import (
	"code/internal/transport/http/handler"
	linkusecase "code/internal/usecase/link"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, service linkusecase.Service, baseURL string) {
	linkHandler := handler.NewLinkHandler(service, baseURL)

	api := router.Group("/api")
	{
		links := api.Group("/links")
		{
			links.GET("", linkHandler.ListLinks)
			links.POST("", linkHandler.CreateLink)
			links.GET("/:id", linkHandler.GetLink)
			links.PUT("/:id", linkHandler.UpdateLink)
			links.DELETE("/:id", linkHandler.DeleteLink)
		}
	}
}
