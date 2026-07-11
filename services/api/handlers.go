package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) getSymbol(c *gin.Context) {
	symbols, err := s.store.GetSymbol(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// return data
	c.JSON(http.StatusOK, symbols)
}

func (s *Server) getTicker(c *gin.Context) {
	// first get parameters from url
	symbol := c.Param("symbol")
	limitStr := c.Query("limit")
	limit, _ := strconv.Atoi(limitStr)

	// get latest 10 rows by default
	if limit <= 0 {
		limit = 10
	}

	// call store method to query db
	tickers, err := s.store.GetTicker(c.Request.Context(), symbol, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// return data
	c.JSON(http.StatusOK, tickers)
}

func (s *Server) getTrade(c *gin.Context) {
	symbol := c.Param("symbol")
	limitStr := c.Query("limit")
	limit, _ := strconv.Atoi(limitStr)

	if limit <= 0 {
		limit = 10
	}

	// query db
	trades, err := s.store.GetTrade(c.Request.Context(), symbol, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// return data
	c.JSON(http.StatusOK, trades)
}
