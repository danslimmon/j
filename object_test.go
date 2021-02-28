package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
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
				assert.False(obj.PendingReview)
			},
		},

		// Tags present
		testCase{
			In: []byte(`---
class: thought
pending_review: true
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
				assert.True(obj.PendingReview)
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

// Thought.Unmarshal should error on malformed input
func TestThought_Unmarshal_Malformed(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	type testCase struct {
		In []byte
	}
	testCases := []testCase{

		// Empty
		testCase{
			In: []byte(``),
		},

		// Frontmatter absent
		testCase{
			In: []byte(`# i'm a markdown document

blah blah blah`),
		},

		// Frontmatter not valid YAML
		testCase{
			In: []byte(`---
class-thought!
pending_review: true
tags:
  - foo
  - bar
---
# blah blah

some text`),
		},
	}

	for i, tc := range testCases {
		t.Logf("test case %d", i)
		obj := NewThought()
		err := obj.Unmarshal(tc.In)
		assert.Error(err)
	}
}

// Thought should marshal to correct YAML/Markdown
func TestThought_Marshal(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	type testCase struct {
		In  *Thought
		Out []byte
	}
	testCases := []testCase{

		// Most basic
		testCase{
			In: &Thought{
				Body:          "# blah blah",
				PendingReview: false,
				Meta: &Meta{
					Class: "thought",
					Tags:  []string{},
				},
			},
			Out: []byte(`---
class: thought
tags: []
pending_review: false
---
# blah blah`),
		},

		// Body empty
		testCase{
			In: &Thought{
				Body: "",
				Meta: &Meta{
					Class: "thought",
					Tags:  []string{},
				},
			},
			Out: []byte(`---
class: thought
tags: []
pending_review: false
---
`),
		},

		// Tags present
		testCase{
			In: &Thought{
				Body: "# blah blah\n\nsome text\n",
				Meta: &Meta{
					Class: "thought",
					Tags:  []string{"foo", "bar"},
				},
			},
			Out: []byte(`---
class: thought
tags:
- foo
- bar
pending_review: false
---
# blah blah

some text
`),
		},
	}

	for i, tc := range testCases {
		t.Logf("test case %d", i)
		obj := tc.In
		b, err := obj.Marshal()
		assert.Nil(err)
		assert.Equal(string(tc.Out), string(b))
	}
}

// Thought should handle mutations via file modification
func TestThought_Mutate(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	type testCase struct {
		In       *Thought
		MutateFn func(string) error
		Match    func(*Thought)
	}

	testCases := []testCase{
		testCase{
			In: &Thought{
				Body: "# hello\n\ni am some markdown",
				Meta: &Meta{
					Class: "thought",
					Tags:  []string{"foo", "bar"},
				},
			},
			MutateFn: func(path string) error {
				return ioutil.WriteFile(
					path,
					[]byte(`---
class: thought
tags:
- foo
- bar
- baz
pending_review: true
---
# new body

different from the old body
`),
					0644,
				)
			},
			Match: func(obj *Thought) {
				assert.Equal("# new body\n\ndifferent from the old body", obj.Body)
				assert.Equal("thought", obj.Meta.Class)
				assert.Equal([]string{"foo", "bar", "baz"}, obj.Meta.Tags)
			},
		},
	}

	for i, tc := range testCases {
		t.Logf("test case %d", i)
		err := tc.In.Mutate(tc.MutateFn)
		assert.Nil(err)
		tc.Match(tc.In)
	}
}
