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

	"github.com/user/jsbug/internal/chrome"
	"github.com/user/jsbug/internal/config"
	"github.com/user/jsbug/internal/fetcher"
	"github.com/user/jsbug/internal/logger"
	"github.com/user/jsbug/internal/parser"
	"github.com/user/jsbug/internal/robots"
	"github.com/user/jsbug/internal/server"
)

const shutdownTimeout = 30 * time.Second

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

	// Initialize Chrome instance
	var chromeInstance *chrome.Instance
	var renderer *chrome.RendererV2

	chromeInstance, err = chrome.New(chrome.InstanceConfig{
		ExecutablePath: cfg.Chrome.ExecutablePath,
		Headless:       cfg.Chrome.Headless,
		DisableGPU:     cfg.Chrome.DisableGPU,
		NoSandbox:      cfg.Chrome.NoSandbox,
		ViewportWidth:  cfg.Chrome.ViewportWidth,
		ViewportHeight: cfg.Chrome.ViewportHeight,
	}, log)

	if err != nil {
		log.Warn("Chrome initialization failed, JS rendering will be unavailable",
			zap.Error(err),
		)
	} else {
		defer chromeInstance.Close()
		renderer = chrome.NewRendererV2(chromeInstance, log)
	}

	// Initialize HTTP fetcher (always available)
	httpFetcher := fetcher.NewFetcher(log)

	// Initialize HTML parser
	htmlParser := parser.NewParser()

	// Create server (SSE manager is created internally)
	srv := server.New(cfg, log)

	// Create and configure render handler
	renderHandler := server.NewRenderHandler(renderer, httpFetcher, htmlParser, cfg, log)
	renderHandler.SetSSEManager(srv.SSEManager())
	srv.SetRenderHandler(renderHandler)

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
		zap.Bool("chrome_available", chromeInstance != nil),
	)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", zap.Error(err))
	}

	log.Info("jsbug stopped")
}
