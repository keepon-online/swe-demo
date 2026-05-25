package http

import (
	"github.com/gin-gonic/gin"

	"github.com/keepon-online/swe-demo/internal/http/handler"
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/ping", handler.Ping)
	return router
}
