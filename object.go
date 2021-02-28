package main

import (
	"bytes"
	"fmt"
	"reflect"

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
			return fmt.Errorf("field 'class' has wrong type '%s'", reflect.TypeOf(vi).Name())
		}
	}
	if vi, ok := frontMatter["tags"]; ok {
		if v, ok := vi.([]interface{}); ok {
			tags := make([]string, 0)
			for _, vvi := range v {
				if vv, ok := vvi.(string); ok {
					tags = append(tags, vv)
				} else {
					return fmt.Errorf("element of field 'tags' has wrong type '%v'", reflect.TypeOf(vvi))
				}
			}
			m.Tags = tags
		} else if vi == nil {
			m.Tags = []string{}
		} else {
			return fmt.Errorf("field 'tags' has wrong type '%v'", reflect.TypeOf(vi))
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
	if len(b) < 3 {
		return fmt.Errorf("frontmatter missing")
	}

	m := front.NewMatter()
	m.Handle("---", front.YAMLHandler)
	frontMatter, body, err := m.Parse(bytes.NewReader(b))
	if err == front.ErrIsEmpty {
		log.WithField("object_id", obj.id).Warn("Markdown section of file is empty")
	} else if err != nil {
		return err
	}

	if vi, ok := frontMatter["pending_review"]; ok {
		if v, ok := vi.(bool); ok {
			obj.PendingReview = v
		} else {
			return fmt.Errorf("field 'pending_review' has wrong type '%s'", reflect.TypeOf(vi).Name())
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
