package main

import (
	"context"
	"log"
	"net/http"

	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JonayMedina/api-music/internal/api"
	"github.com/JonayMedina/api-music/internal/config"
	"github.com/JonayMedina/api-music/internal/handlers"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config     *config.Config
	router     *gin.Engine
	httpServer *http.Server
	services   *api.Services
}

func NewServer(cfg *config.Config, services *api.Services) *Server {
	return &Server{
		config:   cfg,
		services: services,
	}
}

func (s *Server) setupRouter() {
	// Configurar modo de Gin
	if s.config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	s.router = gin.Default()
	s.setupMiddleware()
	s.setupRoutes()
}

func (s *Server) setupMiddleware() {
	s.router.Use(gin.Recovery())
	s.router.Use(gin.Logger())
}

func (s *Server) setupRoutes() {
	api := s.router.Group("/api/" + s.config.APIVersion)

	// Crear handler
	handler := handlers.NewHandler(s.services.SongService)

	// Registrar rutas
	handler.RegisterRoutes(api)
}

// func (s *Server) healthCheck(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{
// 		"status": "OK",
// 		"time":   time.Now().Unix(),
// 	})
// }

// func (s *Server) handleSearch(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{"message": "Search endpoint - To be implemented"})
// }

func (s *Server) Start() error {
	s.setupRouter()

	s.httpServer = &http.Server{
		Addr:    ":" + s.config.Port,
		Handler: s.router,
	}

	// Canal para errores del servidor
	errChan := make(chan error, 1)

	// Iniciar servidor
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Configurar graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case <-quit:
		return s.Shutdown()
	}
}

func (s *Server) Shutdown() error {
	log.Println("Iniciando apagado del servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("Servidor apagado correctamente")
	return nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	services, cleanup, err := api.InitServices(cfg)
	if err != nil {
		log.Fatalf("Error inicializando servicios: %v", err)
	}

	server := NewServer(cfg, services)
	if err := server.Start(); err != nil {
		log.Fatalf("Error en el servidor: %v", err)
	}

	cleanup()
}
