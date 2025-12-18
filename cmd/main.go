package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"linkedin-automation/internal/auth"
	"linkedin-automation/internal/browser"
	"linkedin-automation/internal/config"
	"linkedin-automation/internal/connect"
	"linkedin-automation/internal/logger"
	"linkedin-automation/internal/message"
	"linkedin-automation/internal/scheduler"
	"linkedin-automation/internal/search"
	"linkedin-automation/internal/storage"
)

func main() {
	// Initialize logger
	log := logger.Init()
	log.Info("Starting LinkedIn Automation Bot")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage
	store, err := storage.New(cfg.Storage.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize browser
	browserCtx, err := browser.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize browser: %v", err)
	}
	defer browserCtx.Close()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Received shutdown signal, cleaning up...")
		cancel()
	}()

	// Authenticate
	log.Info("Authenticating with LinkedIn...")
	authService := auth.New(browserCtx, store, cfg)
	if err := authService.Login(ctx); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	log.Info("Authentication successful")

	// Initialize services
	searchService := search.New(browserCtx, store, cfg)
	connectService := connect.New(browserCtx, store, cfg)
	messageService := message.New(browserCtx, store, cfg)
	schedulerService := scheduler.New(cfg)

	// Main automation loop
	log.Info("Starting automation workflow...")
	
	for {
		select {
		case <-ctx.Done():
			log.Info("Shutting down gracefully...")
			return
		default:
			// Check if we should run based on schedule
			if !schedulerService.ShouldRun() {
				log.Info("Outside active hours, sleeping...")
				time.Sleep(30 * time.Minute)
				continue
			}

			// Check rate limits
			if !canProceed(store, cfg) {
				log.Info("Rate limits reached, waiting...")
				time.Sleep(1 * time.Hour)
				continue
			}

			// Execute workflow
			if err := runWorkflow(ctx, searchService, connectService, messageService, store, cfg); err != nil {
				log.Errorf("Workflow error: %v", err)
				time.Sleep(5 * time.Minute)
				continue
			}

			// Wait before next iteration
			log.Info("Workflow completed, taking a break...")
			time.Sleep(time.Duration(cfg.Stealth.IdleBreak.MinDurationSeconds) * time.Second)
		}
	}
}

func runWorkflow(
	ctx context.Context,
	searchSvc *search.Service,
	connectSvc *connect.Service,
	messageSvc *message.Service,
	store *storage.Storage,
	cfg *config.Config,
) error {
	log := logger.Get()

	// Phase 1: Search for profiles
	log.Info("Phase 1: Searching for target profiles...")
	profiles, err := searchSvc.SearchProfiles(ctx)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}
	log.Infof("Found %d profiles", len(profiles))

	// Phase 2: Send connection requests
	log.Info("Phase 2: Sending connection requests...")
	sent, err := connectSvc.SendConnectionRequests(ctx, profiles)
	if err != nil {
		return fmt.Errorf("connection requests failed: %w", err)
	}
	log.Infof("Sent %d connection requests", sent)

	// Phase 3: Send messages to accepted connections
	log.Info("Phase 3: Messaging accepted connections...")
	messaged, err := messageSvc.SendMessages(ctx)
	if err != nil {
		return fmt.Errorf("messaging failed: %w", err)
	}
	log.Infof("Sent %d messages", messaged)

	return nil
}

func canProceed(store *storage.Storage, cfg *config.Config) bool {
	stats := store.GetTodayStats()
	
	if stats.ConnectionsSent >= cfg.RateLimits.Connections.PerDay {
		return false
	}
	
	if stats.MessagesSent >= cfg.RateLimits.Messages.PerDay {
		return false
	}
	
	return true
}
