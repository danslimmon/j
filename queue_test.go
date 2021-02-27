package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testAction struct {
	run func() ([]Action, error)
}

func (a *testAction) Run() ([]Action, error) {
	return a.run()
}

func NewTestAction(runFn func() ([]Action, error)) *testAction {
	return &testAction{run: runFn}
}

// RunNext on a freshly created, empty queue throws the right error.
func TestQueue_RunNextOnEmpty(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	q := NewQueue(FifoDiscipline)
	_, err := q.RunNext()
	assert.Equal(errDequeueOnEmptyQueue, err)
}

// When running an action that returns an error, the error gets passed through.
/*
func TestQueue_RunNextWithError(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	q := NewQueue(FifoDiscipline)
	_, err := q.RunNext()
	assert.Equal(errDequeueOnEmptyQueue, err)
}
*/

// When running an action that returns an error, offspring are still enqueued.
/*
func TestQueue_RunNextWithErrorAndOffspring(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	q := NewQueue(FifoDiscipline)
	_, err := q.RunNext()
	assert.Equal(errDequeueOnEmptyQueue, err)
}
*/

// Enqueue a single action and dequeue it
func TestQueue_EnqueueRunNext_One(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var calls int
	a := NewTestAction(func() ([]Action, error) {
		calls++
		return []Action{}, nil
	})

	q := NewQueue(FifoDiscipline)
	q.Enqueue(a)
	final, err := q.RunNext()
	assert.Equal(nil, err)
	assert.Equal(true, final)
	assert.Equal(1, calls)
}

// Enqueue 3 actions and run them all
func TestQueue_EnqueueRun_Many(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	q := NewQueue(FifoDiscipline)
	var calls int
	for i := 0; i < 3; i++ {
		a := NewTestAction(func() ([]Action, error) {
			calls++
			return []Action{}, nil
		})
		q.Enqueue(a)
	}

	final, err := q.Dequeue()
	assert.Equal(nil, err)
	assert.Equal(false, final)
	assert.Equal(1, calls)

	final, err = q.Dequeue()
	assert.Equal(nil, err)
	assert.Equal(false, final)
	assert.Equal(2, calls)

	final, err = q.Dequeue()
	assert.Equal(nil, err)
	assert.Equal(true, final)
	assert.Equal(3, calls)
}

// Action that adds another action when it's done (i.e. has offspring)
func TestQueue_ActionWithOffspring(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	q := NewQueue(FifoDiscipline)
	var calls int
	a := NewTestAction(func() ([]Action, error) {
		calls++
		offspring := NewTestAction(func() ([]Action, error) {
			return []Action{}, nil
		})
		return []Action{offspring}, nil
	})
	q.Enqueue(a)

	final, err := q.RunNext()
	assert.Equal(nil, err)
	assert.Equal(false, final)
	assert.Equal(1, calls)

	final, err = q.RunNext()
	assert.Equal(nil, err)
	assert.Equal(true, final)
	assert.Equal(2, calls)
}
