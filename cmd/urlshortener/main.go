package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const serverAddr = ":8080"

func newRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return router
}

func main() {
	router := newRouter()

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
