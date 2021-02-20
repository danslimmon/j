package main

import (
	"errors"
	"sync"
)

var (
	errDequeueOnEmptyQueue = errors.New("Dequeue called on empty queue")
)

// QueueDiscipline determines what action will next be dequeued.
//
// It's a function that accepts the list of actions to choose from, and returns an integer
// indicating the index of the next action that should be dequeued.
//
// If an action can't be selected (because the actions slice is empty), the function should return
// -1.
type QueueDiscipline func(actions []Action) int

// FifoDiscipline is a QueueDiscipline that always returns the index of the oldest action
func FifoDiscipline(actions []Action) int {
	if len(actions) <= 0 {
		return -1
	}
	return 0
}

type Action interface {
	Run() error
}

type Queue struct {
	mu         sync.Mutex
	discipline QueueDiscipline
	actions    []Action
}

// Enqueue adds the given Action to the queue.
func (q *Queue) Enqueue(a Action) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.actions = append(q.actions, a)
}

// Dequeue returns the next Action due for processing.
//
// Dequeue decides what Action to dequeue based on the QueueDiscipline with which it was
// initialized.
//
// Return values are as follows: a, the next Action to be processed; final, a bool which indicates
// whether this action is the last one remaining in the queue, and an error. Callers should check
// the value of final on each call. If Dequeue is called on an empty Queue, an error will be
// returned.
func (q *Queue) Dequeue() (a Action, final bool, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.actions) <= 0 {
		return nil, false, errDequeueOnEmptyQueue
	}

	i := q.discipline(q.actions)
	if i == -1 {
		return nil, false, errors.New("Unable to pick next action")
	}

	a = q.actions[i]
	q.actions = append(q.actions[0:i], q.actions[i+1:]...)
	if len(q.actions) <= 0 {
		final = true
	}
	return
}

// NewQueue returns a new, empty Queue.
func NewQueue(disc QueueDiscipline) *Queue {
	return &Queue{
		discipline: disc,
		actions:    make([]Action, 0),
	}
}
