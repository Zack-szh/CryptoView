package api

import (
	"github.com/gin-gonic/gin"
	"github.com/szh/cryptoview/services/api/db"
)

// gin.New() or gin.Default() returns a gin.Engine type
type Server struct {
	router *gin.Engine // accessing server specifics
	store  *db.Store   // entry point for db query
}

func New(store *db.Store) *Server {
	// init server
	r := gin.Default()

	s := &Server{router: r, store: store}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	// all endpoints under v1 for now
	v1 := s.router.Group("/api/v1")

	// -- symbols --
	v1.GET("/symbols", s.getSymbol)
	// -- tickers --
	v1.GET("/ticker/:symbol", s.getTicker)
	// -- trades --
	v1.GET("/trade/:symbol", s.getTrade)
	// -- klines --
	v1.GET("/kline/:symbol", s.getKline)
	// more endpoints: WORK IN PROGRESS
}

// Run() starts the server on given port (Ex: 8080)
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
