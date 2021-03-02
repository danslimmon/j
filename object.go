package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// It may be a thought, or a journal_entry, or whatever other class of object we may have.
type Object interface {
	ID() string
	Bucket() string
	Mutate(func(string) error) error
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

// FrontMatter is the struct into which we unmarshal a document's YAML section.
//
// It is not used on marshal.
type FrontMatter struct {
	// Attributes shared by all objects
	Class string
	Tags  []string

	// Thought-specific attributes
	PendingReview bool
}

// UnmarshalYAML unmarshal's YAML into the FrontMatter.
//
// UnmarshalYAML is part of the yaml.Unmarshaler interface. We implement this interface in order to
// ensure that an empty list in the YAML unmarshals to an empty slice rather than a nil slice.
func (fm *FrontMatter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	buf := new(struct {
		Class string   `yaml:"class"`
		Tags  []string `yaml:"tags"`

		// Thought-specific attributes
		PendingReview bool `yaml:"pending_review"`
	})
	err := unmarshal(buf)
	if err != nil {
		return err
	}

	fm.Class = buf.Class
	if buf.Tags == nil {
		fm.Tags = []string{}
	} else {
		fm.Tags = buf.Tags[:]
	}

	fm.PendingReview = buf.PendingReview

	return nil
}

type Meta struct {
	Class string
	Tags  []string
}

// Update sets the Meta's attributes according to the given FrontMatter.
func (m *Meta) Update(fm *FrontMatter) error {
	if fm.Class != m.Class {
		log.WithFields(log.Fields{
			"from": m.Class,
			"to":   fm.Class,
		}).Warn("Cannot change class of object")
	}

	m.Tags = fm.Tags[:]
	return nil
}

// Thought is an object of class 'thought'. It implements Object.
type Thought struct {
	id string

	// Body is the Markdown part of the file. It will not end in a newline, even when unmarshaling
	// from a file that does.
	Body          []byte
	PendingReview bool
	Meta          *Meta
}

// ID returns the object's UUID
func (obj *Thought) ID() string {
	return obj.id
}

// Bucket returns the bucket that the object should be stored in.
//
// If it should go in the default bucket for objects of its class, then an empty string is returned.
func (obj *Thought) Bucket() string {
	if obj.PendingReview {
		return "to_review"
	}
	return ""
}

// Mutate calls the given function to modify the object.
//
// The function passed to Mutate will be passed the path to a temporary file into which the object
// has been marshaled. The function can make whatever changes it wants to that file. When the
// function returns, the new contents of the temp file will be unmarshaled back into the object.
//
// Mutate returns an error if there's any problem marshaling or unmarshaling, or if the function
// passed to Mutate returns an error. If Mutate returns an error, the object remains unchanged.
func (obj *Thought) Mutate(fn func(string) error) error {
	f, err := ioutil.TempFile("", "jt_*")
	if err != nil {
		return err
	}

	path, err := filepath.Abs(f.Name())
	if err != nil {
		return err
	}
	defer os.Remove(path)

	b, err := obj.Marshal()
	if err != nil {
		return err
	}

	n, err := f.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return fmt.Errorf("failed to write entire object to temp file; wrote %d/%d bytes", n, len(b))
	}
	f.Close()

	// Do the mutation
	if err := fn(path); err != nil {
		return err
	}

	ff, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open temp file after mutation: %w", err)
	}

	bb, err := ioutil.ReadAll(ff)
	if err != nil {
		return fmt.Errorf("failed to read temp file after mutation: %w", err)
	}

	return obj.Unmarshal(bb)
}

// Marshal produces a Markdown-with-YAML-frontmatter document based on the Thought.
//
// The marshaled data will end with a newline.
func (obj *Thought) Marshal() ([]byte, error) {
	frontMatter := yaml.MapSlice{
		yaml.MapItem{Key: "class", Value: obj.Meta.Class},
		yaml.MapItem{Key: "tags", Value: obj.Meta.Tags},
		yaml.MapItem{Key: "pending_review", Value: obj.PendingReview},
	}

	y, err := yaml.Marshal(frontMatter)
	if err != nil {
		return []byte{}, nil
	}

	// Append a newline, but only if there is a body. If the body is empty, then the data is already
	// newline-terminated (from the YAML frontmatter part), so we don't need an extra one.
	body := obj.Body[:]
	if len(body) != 0 {
		body = append(body, []byte("\n")...)
	}

	return bytes.Join(
		[][]byte{
			[]byte{},
			y,
			body,
		},
		[]byte("---\n"),
	), nil
}

// Unmarshal updates the object to match the Markdown it's passed.
func (obj *Thought) Unmarshal(b []byte) error {
	var parts [][]byte = bytes.SplitN(b, []byte("---\n"), 3)
	if len(parts) != 3 {
		return fmt.Errorf(`File must have 3 sections separated by "---\n"`)
	}

	// parts[0]
	if len(parts[0]) != 0 {
		return fmt.Errorf(`File must start with "---\n"`)
	}

	// parts[1] (the YAML part)
	fm := new(FrontMatter)
	err := yaml.Unmarshal(parts[1], fm)
	if err != nil {
		return err
	}

	// parts[2] (the body Markdown)
	body := bytes.TrimRight(
		parts[2][:],
		"\n",
	)

	obj.PendingReview = fm.PendingReview
	obj.Body = body
	if err := obj.Meta.Update(fm); err != nil {
		return err
	}

	return nil
}

func NewThought() *Thought {
	return &Thought{
		Meta: &Meta{
			Class: "thought",
			Tags:  []string{},
		},
	}
}
