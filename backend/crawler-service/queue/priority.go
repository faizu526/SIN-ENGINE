package queue

import (
	"sort"
	"sync"
)

// PriorityQueue implements an in-memory priority queue
type PriorityQueue struct {
	items   []*QueueItem
	mu      sync.RWMutex
	maxSize int
}

type QueueItem struct {
	Value    interface{}
	Priority int
	Index    int
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue(maxSize int) *PriorityQueue {
	return &PriorityQueue{
		items:   make([]*QueueItem, 0),
		maxSize: maxSize,
	}
}

// Push adds an item to the queue
func (pq *PriorityQueue) Push(value interface{}, priority int) bool {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	// Check if at max capacity
	if pq.maxSize > 0 && len(pq.items) >= pq.maxSize {
		// Remove lowest priority item if new item has higher priority
		if priority > pq.items[0].Priority {
			pq.pop()
		} else {
			return false
		}
	}

	item := &QueueItem{
		Value:    value,
		Priority: priority,
		Index:    len(pq.items),
	}

	pq.items = append(pq.items, item)
	pq.bubbleUp(len(pq.items) - 1)

	return true
}

// Pop removes and returns the highest priority item
func (pq *PriorityQueue) Pop() (interface{}, bool) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.items) == 0 {
		return nil, false
	}

	return pq.pop()
}

func (pq *PriorityQueue) pop() interface{} {
	item := pq.items[0]
	lastIndex := len(pq.items) - 1

	// Move last item to root
	pq.items[0] = pq.items[lastIndex]
	pq.items[0].Index = 0
	pq.items = pq.items[:lastIndex]

	if len(pq.items) > 0 {
		pq.bubbleDown(0)
	}

	return item.Value
}

// Peek returns the highest priority item without removing
func (pq *PriorityQueue) Peek() (interface{}, bool) {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	if len(pq.items) == 0 {
		return nil, false
	}

	return pq.items[0].Value, true
}

// Len returns the number of items in the queue
func (pq *PriorityQueue) Len() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.items)
}

// IsEmpty returns true if the queue is empty
func (pq *PriorityQueue) IsEmpty() bool {
	return pq.Len() == 0
}

// IsFull returns true if the queue is at capacity
func (pq *PriorityQueue) IsFull() bool {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return pq.maxSize > 0 && len(pq.items) >= pq.maxSize
}

func (pq *PriorityQueue) bubbleUp(index int) {
	for index > 0 {
		parentIndex := (index - 1) / 2
		if pq.items[index].Priority <= pq.items[parentIndex].Priority {
			break
		}
		pq.swap(index, parentIndex)
		index = parentIndex
	}
}

func (pq *PriorityQueue) bubbleDown(index int) {
	length := len(pq.items)

	for {
		leftChild := 2*index + 1
		rightChild := 2*index + 2
		largest := index

		if leftChild < length && pq.items[leftChild].Priority > pq.items[largest].Priority {
			largest = leftChild
		}

		if rightChild < length && pq.items[rightChild].Priority > pq.items[largest].Priority {
			largest = rightChild
		}

		if largest == index {
			break
		}

		pq.swap(index, largest)
		index = largest
	}
}

func (pq *PriorityQueue) swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].Index = i
	pq.items[j].Index = j
}

// PriorityJob wraps a job with priority for the priority queue
type PriorityJob struct {
	Job      interface{}
	Priority int
}

// SortByPriority sorts jobs by priority (highest first)
func SortByPriority(jobs []PriorityJob) {
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].Priority > jobs[j].Priority
	})
}
