package main

import (
	"bytes"

	"github.com/gernest/front"
	log "github.com/sirupsen/logrus"
)

// Object is something that can be stored in the J workspace.
//
// It may be a thought, or a journal_entry, or whatever other class of object we may have.
type Object interface {
	ID() string
	Bucket() string
	Mutate(func(string) error) error
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type Meta struct {
	Class string
	Tags  []string
}

// Update sets the Meta's attributes according to the given frontmatter map.
func (m *Meta) Update(frontMatter map[string]interface{}) error {
	if vi, ok := frontMatter["class"]; ok {
		if v, ok := vi.(string); ok {
			m.Class = v
		} else {
			log.WithField("class", vi).Warn("class field has wrong type; ignoring")
		}
	}
	if vi, ok := frontMatter["tags"]; ok {
		if v, ok := vi.([]string); ok {
			m.Tags = v
		} else {
			log.WithField("tags", vi).Warn("tags field has wrong type; ignoring")
		}
	}
	return nil
}

// Thought is an object of class 'thought'. It implements Object.
type Thought struct {
	id string

	Body          string
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
func (obj *Thought) Mutate(func(string) error) error {
	return nil
}

func (obj *Thought) Marshal() ([]byte, error) {
	return []byte{}, nil
}

// Unmarshal updates the object to match the Markdown it's passed.
func (obj *Thought) Unmarshal(b []byte) error {
	m := front.NewMatter()
	m.Handle("---", front.YAMLHandler)
	frontMatter, body, err := m.Parse(bytes.NewReader(b))
	if err != nil {
		return err
	}

	if vi, ok := frontMatter["pending_review"]; ok {
		if v, ok := vi.(bool); ok {
			obj.PendingReview = v
		}
	}
	obj.Body = body
	if err := obj.Meta.Update(frontMatter); err != nil {
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
