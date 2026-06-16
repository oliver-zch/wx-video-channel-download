package api

import (
	"fmt"
	"net/http"

	"wx_channels_web/internal/config"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config    *config.Config
	engine    *gin.Engine
	indexHTML []byte
}

func NewServer(cfg *config.Config, indexHTML []byte) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	s := &Server{config: cfg, engine: engine, indexHTML: indexHTML}

	// CORS 中间件
	engine.Use(s.corsMiddleware())

	// API 路由
	engine.GET("/api/status", s.handleStatus)
	engine.POST("/api/parse", s.handleParse)
	engine.GET("/api/proxy", s.handleProxy)
	engine.POST("/api/config/cookie", s.handleUpdateCookie)

	// 嵌入式前端页面
	engine.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", s.indexHTML)
	})

	return s
}

func (s *Server) Run(addr string) error {
	return s.engine.Run(addr)
}

func (s *Server) RunDefault() error {
	addr := fmt.Sprintf("%s:%d", s.config.API.Hostname, s.config.API.Port)
	return s.engine.Run(addr)
}

func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, HEAD, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Range, Accept, Origin, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Range, Accept-Ranges, Content-Type, Content-Length, Content-Disposition")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
