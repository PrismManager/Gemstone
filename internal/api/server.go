package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/PrismManager/gemstone/internal/config"
	"github.com/PrismManager/gemstone/internal/process"
	"github.com/PrismManager/gemstone/internal/stats"
	"github.com/PrismManager/gemstone/internal/types"
)

// Server represents the API server
type Server struct {
	config    *config.Config
	manager   *process.Manager
	collector *stats.Collector
	server    *http.Server
	router    *gin.Engine
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, manager *process.Manager, collector *stats.Collector) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	// router.Use(gin.Logger())

	s := &Server{
		config:    cfg,
		manager:   manager,
		collector: collector,
		router:    router,
	}

	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	if s.config.API.EnableCORS {
		s.router.Use(corsMiddleware())
	}

	if s.config.API.AuthToken != "" {
		s.router.Use(authMiddleware(s.config.API.AuthToken))
	}

	api := s.router.Group("/api/v1")
	{
		api.GET("/health", s.healthCheck)
		api.GET("/system", s.getSystemInfo)
		api.GET("/system/stats", s.getSystemStats)
		api.GET("/system/stats/history", s.getSystemStatsHistory)
		api.GET("/processes", s.listProcesses)
		api.POST("/processes", s.startProcess)
		api.GET("/processes/:id", s.getProcess)
		api.DELETE("/processes/:id", s.deleteProcess)
		api.POST("/processes/:id/stop", s.stopProcess)
		api.POST("/processes/:id/restart", s.restartProcess)
		api.GET("/processes/:id/stats", s.getProcessStats)
		api.GET("/processes/:id/stats/history", s.getProcessStatsHistory)
		api.GET("/processes/:id/logs", s.getProcessLogs)
	}
}

// Start starts the API server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.API.Host, s.config.API.Port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	fmt.Printf("API server listening on %s\n", addr)
	return s.server.ListenAndServe()
}

// Stop stops the API server
func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Message: "Gemstone daemon is running",
	})
}

func (s *Server) getSystemInfo(c *gin.Context) {
	sysStats := s.collector.GetCurrentSystemStats()

	info := types.DaemonInfo{
		Version:      "0.1.0",
		ProcessCount: s.manager.Count(),
		SystemStats:  sysStats,
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Data:    info,
	})
}

func (s *Server) getSystemStats(c *gin.Context) {
	sysStats := s.collector.GetCurrentSystemStats()
	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Data:    sysStats,
	})
}

func (s *Server) getSystemStatsHistory(c *gin.Context) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	sysStats := s.collector.GetSystemStatsHistory(limit)
	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Data:    sysStats,
	})
}

func (s *Server) listProcesses(c *gin.Context) {
	processes := s.manager.List()
	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Data:    processes,
	})
}

func (s *Server) startProcess(c *gin.Context) {
	var req types.StartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if req.Name == "" || req.Command == "" {
		c.JSON(http.StatusBadRequest, types.Response{
			Success: false,
			Error:   "name and command are required",
		})
		return
	}

	info, err := s.manager.Start(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, types.Response{
		Success: true,
		Message: "Process started",
		Data:    info,
	})
}

func (s *Server) getProcess(c *gin.Context) {
	id := c.Param("id")
	info := s.manager.Get(id)

	if info == nil {
		c.JSON(http.StatusNotFound, types.Response{
			Success: false,
			Error:   "process not found",
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Data:    info,
	})
}

func (s *Server) deleteProcess(c *gin.Context) {
	id := c.Param("id")

	if err := s.manager.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Message: "Process deleted",
	})
}

func (s *Server) stopProcess(c *gin.Context) {
	id := c.Param("id")

	if err := s.manager.Stop(id); err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Message: "Process stopped",
	})
}

func (s *Server) restartProcess(c *gin.Context) {
	id := c.Param("id")

	if err := s.manager.Restart(id); err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Message: "Process restarted",
	})
}

func (s *Server) getProcessStats(c *gin.Context) {
	id := c.Param("id")
	procStats := s.manager.Stats(id)

	if procStats == nil {
		c.JSON(http.StatusNotFound, types.Response{
			Success: false,
			Error:   "process not found or not running",
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Data:    procStats,
	})
}

func (s *Server) getProcessStatsHistory(c *gin.Context) {
	id := c.Param("id")
	limit := 100
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	procStats := s.manager.GetStatsHistory(id, limit)
	if procStats == nil {
		c.JSON(http.StatusNotFound, types.Response{
			Success: false,
			Error:   "process not found",
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Data:    procStats,
	})
}

func (s *Server) getProcessLogs(c *gin.Context) {
	id := c.Param("id")
	lines := 100
	logType := c.Query("type")

	if l := c.Query("lines"); l != "" {
		fmt.Sscanf(l, "%d", &lines)
	}

	logs, err := s.manager.GetLogs(id, lines, logType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Success: true,
		Data:    logs,
	})
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func authMiddleware(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth != "Bearer "+token {
			c.JSON(http.StatusUnauthorized, types.Response{
				Success: false,
				Error:   "unauthorized",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
