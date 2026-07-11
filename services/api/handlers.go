package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) getSymbol(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (s *Server) getTicker(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (s *Server) getTrade(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
