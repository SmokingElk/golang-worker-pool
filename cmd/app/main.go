package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/SmokingElk/golang-worker-pull/internal/config"
	"github.com/SmokingElk/golang-worker-pull/internal/worker_pull"
	"github.com/joho/godotenv"
)

type OrderDTO struct {
	OrderID    int     `json:"orderID"`
	TotalPrice float32 `json:"totalPrice"`
}

var logs []string = []string{
	`{ "orderID": 1, "totalPrice": 5.5 }`,
	`{ "orderID": 874, "totalPrice": 23.75 }`,
	`{ "orderID": 1562, "totalPrice": 12.99 }`,
	`{ "orderID": 2987, "totalPrice": 45.3 }`,
	`{ "orderID": 523, "totalPrice": 8.15 }`,
	`{ "orderID": 2041, "totalPrice": 67.89 }`,
	`{ "orderID": 3765, "totalPrice": 3.25 }`,
	`{ "orderID": 987, "totalPrice": 19.5 }`,
	`{ "orderID": 4312, "totalPrice": 34.67 }`,
	`{ "orderID": 2598, "totalPrice": 9.99 }`,
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env file: %s", err)
	}

	cfg := config.MustLoadConfig()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cfg.TimeoutSeconds))

	wp := worker_pull.NewWorkerPoolConfigured(ctx, &worker_pull.WorkerPoolConfig{
		QueueSize:       cfg.Worker.QueueSize,
		NumberOfWorkers: cfg.Worker.NumberOfWorkers,
	})

	sumMtx := sync.Mutex{}
	totalSum := float32(0.0)

	for _, i := range logs {
		wp.Submit(func() {
			var order OrderDTO

			err := json.Unmarshal([]byte(i), &order)

			if err != nil {
				log.Println(fmt.Errorf("failed to parse order: %w", err))
			}

			sumMtx.Lock()
			defer sumMtx.Unlock()
			totalSum += order.TotalPrice
		})
	}

	wp.StopWait()
	cancel()

	fmt.Printf("Sum of orders prices: %f", totalSum)
}
