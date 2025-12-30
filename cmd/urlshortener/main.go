package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const serverAddr = ":8080"

func main() {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	if err := router.Run(serverAddr); err != nil {
		panic(fmt.Errorf("run server: %w", err))
	}
}
