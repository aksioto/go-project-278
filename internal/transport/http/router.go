package httptransport

import (
	"log/slog"

	"code/internal/transport/http/handler"
	linkusecase "code/internal/usecase/link"
	visitusecase "code/internal/usecase/visit"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	router *gin.Engine,
	linkService linkusecase.Service,
	visitService visitusecase.Service,
	baseURL string,
	logger *slog.Logger,
) {
	linkHandler := handler.NewLinkHandler(linkService, visitService, baseURL, logger)

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

		visits := api.Group("/link_visits")
		{
			visits.GET("", linkHandler.ListLinkVisits)
			visits.DELETE("/:id", linkHandler.DeleteLinkVisit)
		}
	}

	router.GET("/r/:code", linkHandler.RedirectToOriginalURL)

}
