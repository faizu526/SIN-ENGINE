package workers

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Worker represents a worker that processes jobs
type Worker struct {
	ID      int
	jobChan chan Job
	quit    chan struct{}
}

// Job represents a work unit
type Job struct {
	ID        string
	Type      string
	Payload   interface{}
	Result    chan Result
	StartedAt time.Time
}

// Result represents the result of a job
type Result struct {
	JobID    string
	Data     interface{}
	Error    error
	Duration time.Duration
}

// WorkerPool manages a pool of workers
type WorkerPool struct {
	workers  []*Worker
	jobQueue chan Job
	wg       sync.WaitGroup
	mu       sync.RWMutex
	active   bool
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers int, jobQueueSize int) *WorkerPool {
	pool := &WorkerPool{
		workers:  make([]*Worker, numWorkers),
		jobQueue: make(chan Job, jobQueueSize),
		active:   true,
	}

	// Create workers
	for i := 0; i < numWorkers; i++ {
		pool.workers[i] = &Worker{
			ID:      i,
			jobChan: make(chan Job),
			quit:    make(chan struct{}),
		}
	}

	return pool
}

// Start starts all workers
func (p *WorkerPool) Start(handler JobHandler) {
	for _, worker := range p.workers {
		p.wg.Add(1)
		go worker.run(p.jobQueue, handler, &p.wg)
	}
	log.Printf("Started %d workers", len(p.workers))
}

// Stop stops all workers
func (p *WorkerPool) Stop() {
	log.Println("Stopping worker pool...")

	p.mu.Lock()
	p.active = false
	p.mu.Unlock()

	// Signal all workers to stop
	for _, worker := range p.workers {
		close(worker.quit)
	}

	// Wait for all workers to finish
	p.wg.Wait()

	log.Println("Worker pool stopped")
}

// Submit submits a job to the pool
func (p *WorkerPool) Submit(job Job) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.active {
		return fmt.Errorf("pool is not active")
	}

	select {
	case p.jobQueue <- job:
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}

// GetStats returns worker pool statistics
func (p *WorkerPool) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"total_workers":  len(p.workers),
		"active":         p.active,
		"queue_length":   len(p.jobQueue),
		"queue_capacity": cap(p.jobQueue),
	}
}

func (w *Worker) run(jobQueue <-chan Job, handler JobHandler, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("Worker %d started", w.ID)

	for {
		select {
		case <-w.quit:
			log.Printf("Worker %d stopped", w.ID)
			return
		case job := <-jobQueue:
			w.processJob(job, handler)
		}
	}
}

func (w *Worker) processJob(job Job, handler JobHandler) {
	start := time.Now()

	log.Printf("Worker %d processing job %s", w.ID, job.ID)

	result := Result{
		JobID:     job.ID,
		StartedAt: start,
	}

	// Process job
	data, err := handler.Handle(job)
	result.Data = data
	result.Error = err
	result.Duration = time.Since(start)

	// Send result
	select {
	case job.Result <- result:
	default:
		log.Printf("Worker %d: result channel full for job %s", w.ID, job.ID)
	}

	if err != nil {
		log.Printf("Worker %d: job %s failed: %v", w.ID, job.ID, err)
	} else {
		log.Printf("Worker %d: job %s completed in %v", w.ID, job.ID, result.Duration)
	}
}

// JobHandler interface defines how jobs are processed
type JobHandler interface {
	Handle(job Job) (interface{}, error)
}

// JobFunc is a function that handles jobs
type JobFunc func(job Job) (interface{}, error)

// Handle implements JobHandler interface
func (f JobFunc) Handle(job Job) (interface{}, error) {
	return f(job)
}
