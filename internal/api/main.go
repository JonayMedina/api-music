// cmd/api/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JonayMedina/api-music/internal/config"
	"github.com/JonayMedina/api-music/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Configurar modo de Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Inicializar router
	router := gin.Default()

	// Middleware global
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Rutas de API
	api := router.Group("/api/" + cfg.APIVersion)
	{
		// Rutas públicas
		public := api.Group("")
		{
			public.GET("/health", healthCheck)
		}

		// Rutas protegidas
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			protected.GET("/search", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Search endpoint - To be implemented"})
			})
		}
	}

	// Configurar servidor HTTP
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Iniciar servidor en goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Configurar graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
		"time":   time.Now().Unix(),
	})
}
