package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/SmokingElk/golang-worker-pool/internal/config"
	"github.com/SmokingElk/golang-worker-pool/internal/worker_pool"
	"github.com/joho/godotenv"
)

type OrderDTO struct {
	OrderID    int     `json:"orderID"`
	TotalPrice float32 `json:"totalPrice"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env file: %s", err)
	}

	cfg := config.MustLoadConfig()

	file, err := os.Open(cfg.LogsPath)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to load logs file: %w", err))
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cfg.TimeoutSeconds))

	wp := worker_pool.NewWorkerPoolConfigured(ctx, &worker_pool.WorkerPoolConfig{
		QueueSize:       cfg.Worker.QueueSize,
		NumberOfWorkers: cfg.Worker.NumberOfWorkers,
	})

	sumMtx := sync.Mutex{}
	totalSum := float32(0.0)
	processedCount := 0

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		logRecord := scanner.Bytes()

		var task worker_pool.Task = func() {
			var order OrderDTO

			err := json.Unmarshal(logRecord, &order)

			if err != nil {
				log.Println(fmt.Errorf("failed to parse order: %w", err))
			}

			sumMtx.Lock()
			defer sumMtx.Unlock()
			totalSum += order.TotalPrice
			processedCount++
		}

		wp.Submit(task)
	}

	wp.StopWait()
	cancel()

	averagePrice := totalSum / float32(processedCount)

	fmt.Printf("Average price of orders: %f", averagePrice)
}
