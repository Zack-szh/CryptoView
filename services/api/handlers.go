package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/szh/cryptoview/services/market-data/binance"
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
	if limit <= 0 || limit > 500 {
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

	if limit <= 0 || limit > 500 {
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

func (s *Server) getKline(c *gin.Context) {
	symbol := c.Param("symbol")
	interval := c.DefaultQuery("interval", "1m")

	var since time.Time
	if sinceStr := c.Query("since"); sinceStr != "" {
		if ms, err := strconv.ParseInt(sinceStr, 10, 64); err == nil {
			since = time.UnixMilli(ms).UTC()
		} else {
			since = time.Now().UTC().AddDate(0, 0, -30)
		}
	} else {
		since = time.Now().UTC().AddDate(0, 0, -30)
	}

	klines, err := s.store.GetKline(c.Request.Context(), symbol, interval, since)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, klines)
}

func (s *Server) getOrderBook(c *gin.Context) {
	symbol := c.Param("symbol")
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit <= 0 || limit > 5000 {
		limit = 20
	}

	book, err := binance.FetchOrderBook(c.Request.Context(), symbol, limit)

	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, book)
}
