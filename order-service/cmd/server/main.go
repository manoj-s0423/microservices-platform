package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shopstream/order-service/internal/client"
	"github.com/shopstream/order-service/internal/config"
	"github.com/shopstream/order-service/internal/handlers"
	"github.com/shopstream/order-service/internal/middleware"
	"github.com/shopstream/order-service/internal/repository"
	"github.com/shopstream/order-service/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// Incorrect/missing required env var -> fail fast with a clear
		// message rather than starting into a broken state.
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(cfg.LogLevel),
	})).With("service", "order-service")

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseDSN())
	if err != nil {
		logger.Error("invalid database configuration", "error", err)
		os.Exit(1)
	}
	poolConfig.MaxConns = int32(cfg.DBMaxOpenConns)
	poolConfig.MinConns = int32(cfg.DBMaxIdleConns)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logger.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	pingCtx, cancel := context.WithTimeout(context.Background(), cfg.DBConnectTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		// Database connection failure at startup - a deliberate,
		// reproducible failure scenario (wrong DB_HOST/DB_PASSWORD).
		logger.Error("database is unreachable at startup", "error", err)
		os.Exit(1)
	}

	userRC := client.NewResilientClient(cfg.UserServiceURL, "user-service", cfg.HTTPClientTimeout, cfg.HTTPRetryAttempts, cfg.HTTPRetryBackoff, logger)
	productRC := client.NewResilientClient(cfg.ProductServiceURL, "product-service", cfg.HTTPClientTimeout, cfg.HTTPRetryAttempts, cfg.HTTPRetryBackoff, logger)
	paymentRC := client.NewResilientClient(cfg.PaymentServiceURL, "payment-service", cfg.HTTPClientTimeout, cfg.HTTPRetryAttempts, cfg.HTTPRetryBackoff, logger)

	userClient := client.NewUserClient(userRC)
	productClient := client.NewProductClient(productRC)
	paymentClient := client.NewPaymentClient(paymentRC)

	orderRepo := repository.NewPostgresOrderRepository(pool)
	orderService := service.NewOrderService(orderRepo, userClient, productClient, paymentClient, logger)

	orderHandler := handlers.NewOrderHandler(orderService)
	healthHandler := handlers.NewHealthHandler(pool)

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery(), middleware.RequestID(), middleware.StructuredLogging(logger))

	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	v1 := router.Group("/api/v1")
	{
		v1.POST("/orders", orderHandler.CreateOrder)
		v1.GET("/orders/:id", orderHandler.GetOrder)
		v1.GET("/orders", orderHandler.ListOrders)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	go func() {
		logger.Info("order-service listening", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown: on SIGTERM/SIGINT, stop accepting new connections
	// and let in-flight requests finish within ShutdownTimeout before
	// forcing an exit. This is what turns a k8s rolling deploy into a
	// clean handover instead of dropped requests.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutdown signal received")
	ctx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", "error", err)
		os.Exit(1)
	}
	logger.Info("server shut down cleanly")
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
