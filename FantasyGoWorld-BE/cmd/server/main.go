package main

import (
	"context"
	"fantasy-go-world-be/internal/api/router"
	"fantasy-go-world-be/internal/config"
	"fantasy-go-world-be/internal/repository/db"
	"fantasy-go-world-be/internal/repository/store"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load config
	configPath := "local/config.yaml"
	if os.Getenv("CONFIG_PATH") != "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Critical Error: LoadConfig failed: %v", err)
	}

	// 2. Initialize Gin mode
	gin.SetMode(cfg.Server.Mode)

	// 3. Initialize MySQL
	if _, err := db.InitDB(cfg.Database); err != nil {
		log.Fatalf("Critical Error: InitDB failed: %v", err)
	}

	// 4. Initialize Redis
	if _, err := store.InitRedis(cfg.Redis); err != nil {
		log.Fatalf("Critical Error: InitRedis failed: %v", err)
	}

	// 5. Initialize Router
	r := router.NewRouter()

	// 6. Start server with graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %d...", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
