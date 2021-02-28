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

		// Most basic
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

		// Body empty
		testCase{
			In: []byte(`---
class: thought
tags:
---`),
			Match: func(obj *Thought) {
				assert.Equal("thought", obj.Meta.Class)
				assert.Equal(0, len(obj.Meta.Tags))
				assert.Equal("", obj.Body)
			},
		},

		// Tags present
		testCase{
			In: []byte(`---
class: thought
tags:
  - foo
  - bar
---
# blah blah

some text`),
			Match: func(obj *Thought) {
				assert.Equal("thought", obj.Meta.Class)
				assert.Equal(2, len(obj.Meta.Tags))
				assert.Equal("foo", obj.Meta.Tags[0])
				assert.Equal("bar", obj.Meta.Tags[1])
				assert.Equal("# blah blah\n\nsome text", obj.Body)
			},
		},
	}

	for i, tc := range testCases {
		t.Logf("test case %d", i)
		obj := NewThought()
		err := obj.Unmarshal(tc.In)
		assert.Equal(nil, err)
		tc.Match(obj)
	}
}
