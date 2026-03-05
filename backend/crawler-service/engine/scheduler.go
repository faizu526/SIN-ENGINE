package engine

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/sin-engine/crawler-service/queue"
)

// Scheduler manages crawl job scheduling
type Scheduler struct {
	jobQueue *queue.JobQueue
	crawler  *CrawlerEngine
	workers  int
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
	paused   bool
	mu       sync.RWMutex
}

func NewScheduler(jobQueue *queue.JobQueue, crawler *CrawlerEngine) *Scheduler {
	return &Scheduler{
		jobQueue: jobQueue,
		crawler:  crawler,
		workers:  3,
		interval: 10 * time.Second,
		stopCh:   make(chan struct{}),
		paused:   false,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	log.Printf("Starting crawler scheduler with %d workers", s.workers)

	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(ctx, i)
	}
}

func (s *Scheduler) worker(ctx context.Context, id int) {
	defer s.wg.Done()

	log.Printf("Crawler worker %d started", id)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Crawler worker %d stopping (context canceled)", id)
			return
		case <-s.stopCh:
			log.Printf("Crawler worker %d stopping", id)
			return
		case <-ticker.C:
			s.processNextJob(ctx)
		}
	}
}

func (s *Scheduler) processNextJob(ctx context.Context) {
	s.mu.RLock()
	paused := s.paused
	s.mu.RUnlock()

	if paused {
		return
	}

	// Get next job from queue
	jobData, err := s.jobQueue.Pop(ctx)
	if err != nil {
		// No jobs available
		return
	}

	// Parse job data
	// This would deserialize the job
	_ = jobData

	log.Println("Processing next crawl job from queue")
}

func (s *Scheduler) Stop() {
	log.Println("Stopping crawler scheduler")

	close(s.stopCh)
	s.wg.Wait()

	log.Println("Crawler scheduler stopped")
}

func (s *Scheduler) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paused = true
	log.Println("Crawler scheduler paused")
}

func (s *Scheduler) Resume() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paused = false
	log.Println("Crawler scheduler resumed")
}

func (s *Scheduler) SetWorkers(workers int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workers = workers
	log.Printf("Scheduler workers set to %d", workers)
}

func (s *Scheduler) SetInterval(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interval = interval
	log.Printf("Scheduler interval set to %v", interval)
}

func (s *Scheduler) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"workers":  s.workers,
		"interval": s.interval.String(),
		"paused":   s.paused,
	}
}
