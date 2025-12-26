package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/user/jsbug/internal/captcha"
	"github.com/user/jsbug/internal/chrome"
	"github.com/user/jsbug/internal/config"
	"github.com/user/jsbug/internal/fetcher"
	"github.com/user/jsbug/internal/logger"
	"github.com/user/jsbug/internal/parser"
	"github.com/user/jsbug/internal/robots"
	"github.com/user/jsbug/internal/screenshot"
	"github.com/user/jsbug/internal/server"
	"github.com/user/jsbug/internal/session"
)

func main() {
	configPath := flag.String("c", "config.yaml", "config file path")
	flag.Parse()

	fmt.Println("jsbug starting...")

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Initialize Chrome pool
	pool, err := chrome.NewChromePool(chrome.InstanceConfig{
		Headless:          true,  // Always headless in production
		NoSandbox:         false, // Sandbox enabled for security
		PoolSize:          cfg.Chrome.PoolSize,
		WarmupURL:         cfg.Chrome.WarmupURL,
		Timeout:           time.Duration(cfg.ChromeTimeout()) * time.Second,
		RestartAfterCount: cfg.Chrome.RestartAfterCount,
		RestartAfterTime:  cfg.Chrome.RestartAfterTime,
	}, log)

	if err != nil {
		log.Fatal("Failed to initialize Chrome pool",
			zap.Error(err),
		)
	}
	defer pool.Shutdown()

	// Initialize HTTP fetcher (always available)
	httpFetcher := fetcher.NewFetcher(log)

	// Initialize HTML parser
	htmlParser := parser.NewParser()

	// Initialize session token manager if captcha is enabled
	var tokenManager *session.TokenManager
	var captchaVerifier *captcha.Verifier
	if cfg.Captcha.Enabled {
		tokenManager, err = session.NewTokenManager(cfg.Captcha.SecretKey, log)
		if err != nil {
			log.Fatal("Failed to create token manager",
				zap.Error(err),
				zap.Int("key_length", len(cfg.Captcha.SecretKey)),
				zap.Int("required_min", session.MinSecretKeyLength),
			)
		}
		captchaVerifier = captcha.NewVerifier(cfg.Captcha.SecretKey, log)
		log.Info("Captcha/session token verification enabled")
	}

	// Create screenshot store with 5-minute TTL
	screenshotStore := screenshot.NewStore(5 * time.Minute)

	// Create a context for background cleanup that will be cancelled on shutdown
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()

	// Start screenshot cleanup goroutine (runs every minute)
	screenshotStore.StartCleanup(cleanupCtx, 1*time.Minute)

	// Create server (SSE manager is created internally)
	srv := server.New(cfg, log)

	// Create and configure auth handler (if captcha enabled)
	if captchaVerifier != nil && tokenManager != nil {
		authHandler := server.NewAuthHandler(captchaVerifier, tokenManager, log)
		srv.SetAuthHandler(authHandler)
	}

	// Create and configure render handler with pool and screenshot store
	renderHandler := server.NewRenderHandler(pool, httpFetcher, htmlParser, cfg, log, tokenManager, screenshotStore)
	renderHandler.SetSSEManager(srv.SSEManager())
	srv.SetRenderHandler(renderHandler)

	// Create and configure screenshot handler
	screenshotHandler := server.NewScreenshotHandler(screenshotStore)
	srv.SetScreenshotHandler(screenshotHandler)

	// Create and configure robots handler
	robotsChecker := robots.NewChecker(log)
	robotsHandler := server.NewRobotsHandler(robotsChecker, log)
	srv.SetRobotsHandler(robotsHandler)

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed", zap.Error(err))
		}
	}()

	log.Info("jsbug started",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
		zap.Int("pool_size", cfg.Chrome.PoolSize),
	)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutdown signal received")

	// Graceful shutdown sequence:
	// 1. Stop accepting new HTTP requests first
	log.Info("Shutting down HTTP server...")
	ctx, cancel := context.WithTimeout(context.Background(), chrome.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", zap.Error(err))
	}

	// 2. Then shutdown pool (in-flight renders should already be done)
	log.Info("Shutting down Chrome pool...")
	if err := pool.Shutdown(); err != nil {
		log.Error("Pool shutdown error", zap.Error(err))
	}

	log.Info("jsbug stopped")
}
