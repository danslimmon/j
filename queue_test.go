package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A basic enqueue/enqueue works correctly.
//func TestQueue_EnqueueDequeue(t *testing.T) {
//	q := NewQueue()
//
//}

// Dequeue on a freshly created, empty queue throws the right error.
func TestQueue_EnqueueDequeue(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	q := NewQueue(FifoDiscipline)
	_, _, err := q.Dequeue()
	assert.Equal(errDequeueOnEmptyQueue, err)
}
