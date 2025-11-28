package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vantutran2k1/flowfleet/internal/adapter/handler"
	"github.com/vantutran2k1/flowfleet/internal/adapter/logger"
	"github.com/vantutran2k1/flowfleet/internal/adapter/storage/postgres"
	"github.com/vantutran2k1/flowfleet/internal/config"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	appLogger, _ := logger.New()
	defer appLogger.Sync()

	dbConfig, err := pgxpool.ParseConfig(cfg.DBUrl)
	if err != nil {
		appLogger.Fatal("unable to parse db config", zap.Error(err))
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		appLogger.Fatal("unable to create db pool", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		appLogger.Fatal("cannot connect to db", zap.Error(err))
	}

	appLogger.Info("connected to database via pgxpool")

	store := postgres.New(pool)
	driverHandler := handler.NewDriverHandler(store)

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "UP", "env": cfg.Env})
	})

	api := r.Group("/api/v1")
	{
		api.POST("/drivers", driverHandler.CreateDriver)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		appLogger.Info("starting server", zap.String("port", cfg.ServerPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("listen: %s\n", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatal("server forced to shutdown:", zap.Error(err))
	}

	appLogger.Info("server exiting")
}
