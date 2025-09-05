package worker_pool_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/SmokingElk/golang-worker-pool/internal/worker_pool"
	"github.com/stretchr/testify/assert"
)

func TestSubmit(t *testing.T) {
	t.Run("completes asynchronously after submit", func(t *testing.T) {
		wp := worker_pool.NewWorkerPool(4)
		defer wp.Stop()

		mtx := sync.Mutex{}
		counter := 0
		expectedCounter := 5

		ctx, cancel := context.WithCancel(context.Background())

		var task worker_pool.Task = func() {
			time.Sleep(time.Millisecond * 35)
			mtx.Lock()
			defer mtx.Unlock()
			counter++

			if counter == expectedCounter {
				cancel()
			}
		}

		for range expectedCounter {
			wp.Submit(task)
		}

		timeoutExit := false

		select {
		case <-time.After(time.Millisecond * 100):
			timeoutExit = true
		case <-ctx.Done():
		}

		cancel()

		assert.Equal(t, expectedCounter, counter)
		assert.Equal(t, false, timeoutExit)
	})

	t.Run("ignored for stopped worker pool", func(t *testing.T) {
		wp := worker_pool.NewWorkerPool(4)
		wp.Stop()

		ctx, cancel := context.WithCancel(context.Background())
		value := 0
		expectedValue := 0

		wp.Submit(func() {
			time.Sleep(time.Millisecond * 100)
			value = 5
		})

		timeoutExit := false

		select {
		case <-time.After(time.Millisecond * 500):
			timeoutExit = true
		case <-ctx.Done():
		}

		cancel()

		assert.Equal(t, expectedValue, value)
		assert.Equal(t, true, timeoutExit)
	})
}

func TestSubmitWait(t *testing.T) {
	t.Run("completes synchronously after submit", func(t *testing.T) {
		wp := worker_pool.NewWorkerPool(4)
		defer wp.Stop()

		ctx, cancel := context.WithCancel(context.Background())
		value := 0
		expectedValue := 5

		go func() {
			localValue := 0
			mtx := sync.Mutex{}

			wp.SubmitWait(func() {
				time.Sleep(time.Millisecond * 100)
				mtx.Lock()
				defer mtx.Unlock()
				localValue = expectedValue
			})

			mtx.Lock()
			defer mtx.Unlock()
			value = localValue

			cancel()
		}()

		timeoutExit := false

		select {
		case <-time.After(time.Millisecond * 500):
			timeoutExit = true
		case <-ctx.Done():
		}

		cancel()

		assert.Equal(t, expectedValue, value)
		assert.Equal(t, false, timeoutExit)
	})

	t.Run("ignored for stopped worker pool", func(t *testing.T) {
		wp := worker_pool.NewWorkerPool(4)
		wp.Stop()

		ctx, cancel := context.WithCancel(context.Background())
		value := 0
		expectedValue := 0

		wp.SubmitWait(func() {
			time.Sleep(time.Millisecond * 100)
			value = 5
		})

		timeoutExit := false

		select {
		case <-time.After(time.Millisecond * 500):
			timeoutExit = true
		case <-ctx.Done():
		}

		cancel()

		assert.Equal(t, expectedValue, value)
		assert.Equal(t, true, timeoutExit)
	})
}

func TestStop(t *testing.T) {
	t.Run("waits for current tasks", func(t *testing.T) {
		wp := worker_pool.NewWorkerPool(4)

		mtx := sync.Mutex{}
		counter := 0
		expectedCounter := 4

		for range expectedCounter {
			wp.Submit(func() {
				time.Sleep(time.Millisecond * 100)
				mtx.Lock()
				defer mtx.Unlock()
				counter++
			})
		}

		time.Sleep(time.Millisecond * 10)
		wp.Stop()

		assert.Equal(t, expectedCounter, counter)
	})

	t.Run("does not wait for tasks in queue", func(t *testing.T) {
		wp := worker_pool.NewWorkerPool(4)

		mtx := sync.Mutex{}
		counter := 0
		expectedCounter := 4
		tasksCount := 8

		for range tasksCount {
			wp.Submit(func() {
				time.Sleep(time.Millisecond * 100)
				mtx.Lock()
				defer mtx.Unlock()
				counter++
			})
		}

		time.Sleep(time.Millisecond * 10)
		wp.Stop()

		assert.Equal(t, expectedCounter, counter)
	})

	t.Run("does nothing for stopped worker", func(t *testing.T) {
		wp := worker_pool.NewWorkerPool(4)

		mtx := sync.Mutex{}
		counter := 0
		expectedCounter := 8
		tasksCount := 8

		for range tasksCount {
			wp.Submit(func() {
				time.Sleep(time.Millisecond * 100)
				mtx.Lock()
				defer mtx.Unlock()
				counter++
			})
		}

		time.Sleep(time.Millisecond * 10)
		wp.StopWait()
		wp.Stop()

		assert.Equal(t, expectedCounter, counter)
	})
}

func TestStopWait(t *testing.T) {
	t.Run("wait for all added tasks", func(t *testing.T) {
		wp := worker_pool.NewWorkerPool(4)

		mtx := sync.Mutex{}
		counter := 0
		expectedCounter := 8
		tasksCount := 8

		for range tasksCount {
			wp.Submit(func() {
				time.Sleep(time.Millisecond * 100)
				mtx.Lock()
				defer mtx.Unlock()
				counter++
			})
		}

		time.Sleep(time.Millisecond * 10)
		wp.StopWait()

		assert.Equal(t, expectedCounter, counter)
	})

	t.Run("does nothing for stopped worker", func(t *testing.T) {
		wp := worker_pool.NewWorkerPool(4)

		mtx := sync.Mutex{}
		counter := 0
		expectedCounter := 4
		tasksCount := 8

		for range tasksCount {
			wp.Submit(func() {
				time.Sleep(time.Millisecond * 100)
				mtx.Lock()
				defer mtx.Unlock()
				counter++
			})
		}

		time.Sleep(time.Millisecond * 10)
		wp.Stop()
		wp.StopWait()

		assert.Equal(t, expectedCounter, counter)
	})
}
