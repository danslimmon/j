package testutil

import (
	"io/ioutil"
	"os"
)

/*
NewTempDir creates a temporary directory and returns a corresponding TempDir object.
*/
func NewTempDir(pattern string) (TempDir, error) {
	path, err := ioutil.TempDir("", pattern)
	if err != nil {
		return TempDir{}, err
	}
	return TempDir{Path: path}, nil
}

/*
TempDir represents a temporary directory in the filesystem.

You use it like so:

    tempDir, err := NewTempDir()
	if err != nil {...}
	defer tempDir.Cleanup()
	os.Chdir(tempDir)
	... use the temporary directory as you see fit...

It's safe to call Cleanup() on the TempDir returned by NewTempDir(), even if
NewTempDir() returned an error.
*/
type TempDir struct {
	Path string
}

/*
Cleanup() removes the temporary directory and anything in it.

It's safe to call Cleanup on the return value of NewT
*/
func (tempDir TempDir) Cleanup() {
	if tempDir.Path == "" {
		// tempDir hasn't been initialized; presumably there was some problem creating
		// the directory in NewTempDir().
		return
	}

	// ignore error because this function gets called in a defer(), and if we can't
	// remove the directory, oh well.
	_ = os.RemoveAll(tempDir.Path)
}
