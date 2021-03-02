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
				assert.Equal([]byte("# blah blah"), obj.Body)
			},
		},

		// Body empty
		testCase{
			In: []byte(`---
class: thought
tags:
---
`),
			Match: func(obj *Thought) {
				assert.Equal("thought", obj.Meta.Class)
				assert.Equal(0, len(obj.Meta.Tags))
				assert.Equal([]byte{}, obj.Body)
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
				assert.Equal([]string{"foo", "bar"}, obj.Meta.Tags)
				assert.Equal([]byte("# blah blah\n\nsome text"), obj.Body)
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
				Body:          []byte("# blah blah"),
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
# blah blah
`),
		},

		// Body empty
		testCase{
			In: &Thought{
				Body: []byte{},
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
				Body: []byte("# blah blah\n\nsome text"),
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

		// User modifies tags and body
		testCase{
			In: &Thought{
				Body: []byte("# hello\n\ni am some markdown"),
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
				assert.Equal([]byte("# new body\n\ndifferent from the old body"), obj.Body)
				assert.Equal("thought", obj.Meta.Class)
				assert.Equal([]string{"foo", "bar", "baz"}, obj.Meta.Tags)
			},
		},

		// User removes all tags and deletes pending_review field
		testCase{
			In: &Thought{
				Body: []byte("# boopty bewpty spoot\n"),
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
---
# boopty bewpty spoot
`),
					0644,
				)
			},
			Match: func(obj *Thought) {
				assert.Equal([]byte("# boopty bewpty spoot"), obj.Body)
				assert.Equal("thought", obj.Meta.Class)
				assert.Equal([]string{}, obj.Meta.Tags)
			},
		},

		// User deletes body and doesn't terminate with a newline
		testCase{
			In: &Thought{
				Body: []byte("# tomorrow\n\nand tomorrow"),
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
---
`),
					0644,
				)
			},
			Match: func(obj *Thought) {
				assert.Equal([]byte{}, obj.Body)
				assert.Equal("thought", obj.Meta.Class)
				assert.Equal([]string{"foo", "bar"}, obj.Meta.Tags)
			},
		},

		// User attempts to change class; this should be ignored.
		testCase{
			In: &Thought{
				Body: []byte("# did you ever wear pants\n\nthat were bigger than you meant to?"),
				Meta: &Meta{
					Class: "thought",
					Tags:  []string{"foo", "bar"},
				},
			},
			MutateFn: func(path string) error {
				return ioutil.WriteFile(
					path,
					[]byte(`---
class: journal_entry
tags:
- foo
- bar
---
# did you ever wear pants

that were bigger than you meant to?


`),
					0644,
				)
			},
			Match: func(obj *Thought) {
				assert.Equal([]byte("# did you ever wear pants\n\nthat were bigger than you meant to?"), obj.Body)
				assert.Equal("thought", obj.Meta.Class)
				assert.Equal([]string{"foo", "bar"}, obj.Meta.Tags)
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

// Thought should return error if Mutate goes pear-shaped
func TestThought_Mutate_Error(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	type testCase struct {
		In       *Thought
		MutateFn func(string) error
	}

	testCases := []testCase{

		// User's modification results in unparseable YAML
		testCase{
			In: &Thought{
				Body: []byte("# hello\n\ni am some markdown"),
				Meta: &Meta{
					Class: "thought",
					Tags:  []string{"foo", "bar"},
				},
			},
			MutateFn: func(path string) error {
				return ioutil.WriteFile(
					path,
					[]byte(`---
# hello

i am some markdown
`),
					0644,
				)
			},
		},

		/*
					// User removes all tags and deletes pending_review field
					testCase{
						In: &Thought{
							Body: "# boopty bewpty spoot\n",
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
			---
			# boopty bewpty spoot
			`),
								0644,
							)
						},
						Match: func(obj *Thought) {
							assert.Equal("# boopty bewpty spoot", obj.Body)
							assert.Equal("thought", obj.Meta.Class)
							assert.Equal([]string{}, obj.Meta.Tags)
						},
					},

					// User deletes body and doesn't terminate with a newline
					testCase{
						In: &Thought{
							Body: "# tomorrow\n\nand tomorrow\n",
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
			---`),
								0644,
							)
						},
						Match: func(obj *Thought) {
							assert.Equal("", obj.Body)
							assert.Equal("thought", obj.Meta.Class)
							assert.Equal([]string{"foo", "bar"}, obj.Meta.Tags)
						},
					},

					// User attempts to change class; this should be ignored.
					testCase{
						In: &Thought{
							Body: "# did you ever wear pants\n\nthat were bigger than you meant to?",
							Meta: &Meta{
								Class: "thought",
								Tags:  []string{"foo", "bar"},
							},
						},
						MutateFn: func(path string) error {
							return ioutil.WriteFile(
								path,
								[]byte(`---
			class: journal_entry
			tags:
			- foo
			- bar
			---
			# did you ever wear pants

			that were bigger than you meant to?


			`),
								0644,
							)
						},
						Match: func(obj *Thought) {
							assert.Equal("# did you ever wear pants\n\nthat were bigger than you meant to?", obj.Body)
							assert.Equal("thought", obj.Meta.Class)
							assert.Equal([]string{"foo", "bar"}, obj.Meta.Tags)
						},
					},
		*/
	}

	for i, tc := range testCases {
		t.Logf("test case %d", i)
		err := tc.In.Mutate(tc.MutateFn)
		assert.Error(err)
	}
}
