package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/szh/cryptoview/services/indicators"
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

// shared by getKline and getIndicatorSeries so both cover the same default range
// (30 days back) when the caller doesn't pin an explicit `since`
func parseSince(c *gin.Context) time.Time {
	if sinceStr := c.Query("since"); sinceStr != "" {
		if ms, err := strconv.ParseInt(sinceStr, 10, 64); err == nil {
			return time.UnixMilli(ms).UTC()
		}
	}
	return time.Now().UTC().AddDate(0, 0, -30)
}

func (s *Server) getKline(c *gin.Context) {
	symbol := c.Param("symbol")
	interval := c.DefaultQuery("interval", "1m")
	since := parseSince(c)

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

func (s *Server) getIndicators(c *gin.Context) {
	symbol := c.Param("symbol")
	interval := c.DefaultQuery("interval", "1m")

	snap, err := indicators.BuildSnapshot(c.Request.Context(), s.store, symbol, interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, snap)
}

func (s *Server) getIndicatorSeries(c *gin.Context) {
	symbol := c.Param("symbol")
	interval := c.DefaultQuery("interval", "1m")
	since := parseSince(c)

	series, err := indicators.BuildIndicatorSeries(c.Request.Context(), s.store, symbol, interval, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, series)
}
