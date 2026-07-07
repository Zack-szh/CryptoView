package api

import "github.com/gin-gonic/gin"

// gin.New() or gin.Default() returns a gin.Engine type
type Server struct {
	router *gin.Engine
}

func New() *Server {
	// init server
	r := gin.Default()
	s := &Server{router: r}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	// all endpoints under v1 for now
	v1 := s.router.Group("api/v1")

	// -- symbols --
	v1.GET("/symbols", s.getSymbol)
	// -- tickers --
	v1.GET("/tickers/:symbol", s.getTicker)
	// -- trades --
	v1.GET("/trade/:symbol", s.getTrade)
	// more endpoints: WORK IN PROGRESS
}

// Run() starts the server on given port (Ex: 8080)
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
