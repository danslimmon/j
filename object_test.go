package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Thought should satisfy the Object interface
func TestThought_SatisfiesObjectInterface(t *testing.T) {
	var _ Object = (*Thought)(nil)
}

// Thought should be populated from Markdown correctly
func TestThought_Unmarshal(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	type testCase struct {
		In    []byte
		Match func(*Thought)
	}
	testCases := []testCase{
		testCase{
			In: []byte(`---
class: thought
tags:
---
# blah blah`),
			Match: func(obj *Thought) {
				assert.Equal("thought", obj.Meta.Class)
				assert.Equal(0, len(obj.Meta.Tags))
				assert.Equal("# blah blah", obj.Body)
			},
		},
	}

	for _, tc := range testCases {
		obj := NewThought()
		obj.Unmarshal(tc.In)
		tc.Match(obj)
	}
}
