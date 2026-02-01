package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/zap"

	"github.com/samuelayo/opensms/internal/infrastructure/config"
	"github.com/samuelayo/opensms/internal/infrastructure/database"
	"github.com/samuelayo/opensms/internal/infrastructure/cache"
	"github.com/samuelayo/opensms/internal/infrastructure/eventbus"
	"github.com/samuelayo/opensms/internal/infrastructure/observability"
	"github.com/samuelayo/opensms/internal/infrastructure/storage"
	"github.com/samuelayo/opensms/internal/server"
)

func main() {
	// Initialize logger
	log := observability.NewLogger()
	defer log.Sync()

	log.Info("Starting OpenSMS API Server",
		zap.String("version", "1.0.0"),
		zap.String("env", os.Getenv("APP_ENV")),
	)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize observability
	cleanup, err := observability.InitTracing(cfg)
	if err != nil {
		log.Fatal("Failed to initialize tracing", zap.Error(err))
	}
	defer cleanup()

	observability.InitMetrics()

	// Initialize infrastructure
	ctx := context.Background()

	// Database
	db, err := database.NewPostgresDB(ctx, cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Cache
	redisClient := cache.NewRedisClient(cfg.Redis)
	defer redisClient.Close()

	// Event Bus
	natsConn, err := eventbus.NewNATSConnection(cfg.NATS)
	if err != nil {
		log.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer natsConn.Close()

	// Object Storage
	minioClient, err := storage.NewMinIOClient(cfg.MinIO)
	if err != nil {
		log.Fatal("Failed to initialize MinIO client", zap.Error(err))
	}

	// Initialize Fiber app with security configurations
	app := fiber.New(fiber.Config{
		AppName:               "OpenSMS API v1.0.0",
		ServerHeader:          "OpenSMS",
		StrictRouting:         true,
		CaseSensitive:         true,
		ErrorHandler:          server.ErrorHandler,
		DisableStartupMessage: false,
		EnableTrustedProxyCheck: true,
		TrustedProxies:        cfg.Server.TrustedProxies,
		ProxyHeader:           "X-Forwarded-For",
		ReadTimeout:           time.Second * 30,
		WriteTimeout:          time.Second * 30,
		IdleTimeout:           time.Second * 120,
		BodyLimit:             10 * 1024 * 1024, // 10MB
	})

	// Security Middleware
	app.Use(helmet.New(helmet.Config{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))

	// CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Server.AllowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Tenant-ID,X-Request-ID",
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	// Request ID
	app.Use(requestid.New())

	// Logger
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	}))

	// Recovery
	app.Use(recover.New())

	// Compression
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// Rate Limiting (DDoS protection)
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		LimiterMiddleware: limiter.SlidingWindow{},
		KeyGenerator: func(c *fiber.Ctx) string {
			// Rate limit by IP + User ID (if authenticated)
			return c.IP()
		},
	}))

	// Initialize server with dependencies
	srv := server.NewServer(server.Dependencies{
		Config:      cfg,
		DB:          db,
		Cache:       redisClient,
		EventBus:    natsConn,
		Storage:     minioClient,
		Logger:      log,
	})

	// Register routes
	srv.RegisterRoutes(app)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now().Unix(),
		})
	})

	// Metrics endpoint
	app.Get("/metrics", observability.MetricsHandler)

	// Start server in goroutine
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		log.Info("Starting server", zap.String("address", addr))

		if err := app.Listen(addr); err != nil {
			log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited gracefully")
}
