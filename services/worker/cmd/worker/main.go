package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"clan-budget/services/worker/internal/config"
	"clan-budget/services/worker/internal/db"
	"clan-budget/services/worker/internal/processor"
	"clan-budget/services/worker/internal/queue"
	"clan-budget/services/worker/internal/scheduler"
)

func main() {
	// Initialize Envs
	config.Load()
	log.Println("Starting Clan Budget Worker Service...")

	// Setup Connections
	database := db.Connect()
	defer database.Close()

	mq := queue.Connect()
	defer mq.Close()

	// Initialize Modules
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Setup processing worker pool (Backpressure limited to 5 concurrent jobs)
	consumerPool := processor.NewProcessor(mq, database, 5)
	
	// Start consuming messages
	go consumerPool.Start(ctx)

	// 2. Setup Monthly Schedule
	cronJob := scheduler.NewScheduler(database, mq)
	cronJob.Start()
	defer cronJob.Stop()

	// Development trigger: If TEST_RUN is true, trigger cron job immediately
	if config.GetEnv("TEST_RUN") == "true" {
		log.Println("[DEV] Running cron immediately due to TEST_RUN=true")
		cronJob.RunNow()
	}

	// Wait for shutdown signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Worker service is up and running. Received signal: %v to shutdown", <-sig)
	
	log.Println("Shutting down cleanly...")
}
