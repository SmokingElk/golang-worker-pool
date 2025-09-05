package worker_pull

import (
	"context"
	"sync"
)

const DEFAULT_QUEUE_SIZE = 128

type Task func()

type WorkerPoolConfig struct {
	QueueSize       int
	NumberOfWorkers int
}

type WorkerPool struct {
	tasksQueue  chan Task
	workerGroup sync.WaitGroup
	stopped     bool
	cancelTasks context.CancelFunc
}

// создание WorkerPool с параметрами
func NewWorkerPoolConfigured(ctx context.Context, cfg *WorkerPoolConfig) *WorkerPool {
	ctx, cancel := context.WithCancel(ctx)

	pool := &WorkerPool{
		tasksQueue:  make(chan Task, cfg.QueueSize),
		stopped:     false,
		cancelTasks: cancel,
	}

	for range cfg.NumberOfWorkers {
		// локализуем инкремент и декремент wait группы в одном месте,
		// чтобы снизить вероятность утечки горутин
		pool.workerGroup.Add(1)
		go func() {
			defer pool.workerGroup.Done()
			pool.worker(ctx)
		}()
	}

	return pool
}

// NewWorkerPool - создать WorkerPool по умолчанию
func NewWorkerPool(numberOfWorkers int) *WorkerPool {
	cfg := &WorkerPoolConfig{
		NumberOfWorkers: numberOfWorkers,
		QueueSize:       DEFAULT_QUEUE_SIZE,
	}

	return NewWorkerPoolConfigured(context.Background(), cfg)
}

// Submit - добавить таску в воркер пул
func (wp *WorkerPool) Submit(task Task) {
	if wp.stopped {
		return
	}

	wp.tasksQueue <- task
}

// SubmitWait - добавить таску в воркер пул и дождаться окончания ее выполнения
func (wp *WorkerPool) SubmitWait(task Task) {
	if wp.stopped {
		return
	}

	var waitTask sync.WaitGroup
	waitTask.Add(1)

	wp.Submit(func() {
		task()
		waitTask.Done()
	})

	waitTask.Wait()
}

// Stop - остановить воркер пул, дождаться выполнения только тех тасок, которые выполняются сейчас
func (wp *WorkerPool) Stop() {
	if wp.stopped {
		return
	}

	wp.cancelTasks()
	wp.StopWait()
}

// StopWait - остановить воркер пул, дождаться выполнения всех тасок, даже тех,
// что не начали выполняться, но лежат в очереди
func (wp *WorkerPool) StopWait() {
	if wp.stopped {
		return
	}

	wp.stopped = true
	close(wp.tasksQueue)
	wp.workerGroup.Wait()
}

func (wp *WorkerPool) worker(cancelContext context.Context) {
	workDone := false

	for !workDone {
		select {
		case <-cancelContext.Done():
			workDone = true
		default:
			task, ok := <-wp.tasksQueue
			if !ok {
				workDone = true
				break
			}

			task()
		}
	}
}
