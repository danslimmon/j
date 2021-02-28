package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"github.com/gernest/front"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
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
//
// Update will return an error if any fields are of the wrong type.
func (m *Meta) Update(frontMatter map[string]interface{}) error {
	if vi, ok := frontMatter["class"]; ok {
		if v, ok := vi.(string); ok {
			if v != m.Class {
				log.WithFields(log.Fields{
					"from": m.Class,
					"to":   v,
				}).Warn("Cannot change class of object")
			}
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

	// Body is the Markdown part of the file. It will not end in a newline, even when unmarshaling
	// from a file that does.
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

	b, err := yaml.Marshal(frontMatter)
	if err != nil {
		return []byte{}, nil
	}

	// Append a newline, but only if there is a body. If the body is empty, then the data is already
	// newline-terminated (from the YAML frontmatter part), so we don't need an extra one.
	var suffix string
	if obj.Body != "" {
		suffix = "\n"
	}

	return []byte(fmt.Sprintf("---\n%s---\n%s%s", b, obj.Body, suffix)), nil
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
