package workers

import (
	"context"
	"log"
	"sync"
	"time"
)

// Pool manages a pool of workers for parallel processing
type Pool struct {
	workers    int
	jobQueue   chan func() error
	resultChan chan error
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	active     bool
	mu         sync.RWMutex
}

// PoolConfig contains configuration for the pool
type PoolConfig struct {
	Workers   int
	QueueSize int
}

// NewPool creates a new worker pool
func NewPool(config PoolConfig) *Pool {
	if config.Workers <= 0 {
		config.Workers = 5
	}
	if config.QueueSize <= 0 {
		config.QueueSize = 100
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &Pool{
		workers:    config.Workers,
		jobQueue:   make(chan func() error, config.QueueSize),
		resultChan: make(chan error, config.QueueSize),
		ctx:        ctx,
		cancel:     cancel,
		active:     true,
	}

	// Start workers
	for i := 0; i < config.Workers; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	log.Printf("Worker pool started with %d workers", config.Workers)

	return pool
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()

	log.Printf("Pool worker %d started", id)

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("Pool worker %d stopping", id)
			return
		case job, ok := <-p.jobQueue:
			if !ok {
				log.Printf("Pool worker %d: job queue closed", id)
				return
			}

			// Execute job
			err := job()
			if err != nil {
				log.Printf("Pool worker %d: job failed: %v", id, err)
			}

			// Send result
			select {
			case p.resultChan <- err:
			default:
			}
		}
	}
}

// Submit adds a job to the pool
func (p *Pool) Submit(job func() error) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.active {
		return ErrPoolNotActive
	}

	select {
	case p.jobQueue <- job:
		return nil
	default:
		return ErrPoolFull
	}
}

// SubmitWithTimeout adds a job with timeout
func (p *Pool) SubmitWithTimeout(job func() error, timeout time.Duration) error {
	done := make(chan error, 1)

	err := p.Submit(func() error {
		done <- job()
		return nil
	})

	if err != nil {
		return err
	}

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return ErrTimeout
	}
}

// GetResults returns a channel of results
func (p *Pool) GetResults() <-chan error {
	return p.resultChan
}

// Stop stops the pool
func (p *Pool) Stop() {
	log.Println("Stopping worker pool...")

	p.mu.Lock()
	p.active = false
	p.cancel()
	p.mu.Unlock()

	// Close job queue
	close(p.jobQueue)

	// Wait for workers to finish
	p.wg.Wait()

	// Close result channel
	close(p.resultChan)

	log.Println("Worker pool stopped")
}

// GetStats returns pool statistics
func (p *Pool) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"workers":      p.workers,
		"queue_length": len(p.jobQueue),
		"queue_cap":    cap(p.jobQueue),
		"results_len":  len(p.resultChan),
		"active":       p.active,
	}
}

// Errors for pool operations
var (
	ErrPoolNotActive = &PoolError{"pool is not active"}
	ErrPoolFull      = &PoolError{"pool queue is full"}
	ErrTimeout       = &PoolError{"operation timed out"}
)

type PoolError struct {
	msg string
}

func (e *PoolError) Error() string {
	return e.msg
}
