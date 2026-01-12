package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KBM2795/DevArena-Backend/internal/config"
	"github.com/KBM2795/DevArena-Backend/internal/database"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router     *gin.Engine
	db         *database.Database
	config     config.Server
	httpServer *http.Server
}

func NewServer(cfg config.Server, db *database.Database, env string) *Server {
	// Set Gin mode based on environment
	if env != "Dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"https://devarena.dev", // production
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Authorization",
			"Content-Type",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: false, // IMPORTANT
		MaxAge:           12 * time.Hour,
	}))

	server := &Server{
		router: router,
		db:     db,
		config: cfg,
	}

	server.RegisterRoutes()
	return server
}

func (s *Server) RegisterRoutes() {
	s.router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to DevArena Backend",
		})
	})

	s.router.GET("/health", s.HealthHandler)
}

func (s *Server) HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.config.Port)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	// Channel to listen for errors from the server
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", addr)
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		if err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}

	case sig := <-shutdown:
		log.Printf("Received signal %v, starting graceful shutdown...", sig)

		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := s.httpServer.Shutdown(ctx); err != nil {
			// Force shutdown if graceful shutdown fails
			s.httpServer.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}

		log.Println("Server stopped gracefully")
	}

	return nil
}
