package testutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
Tests that TempDir gets created and cleaned up correctly.
*/
func TestTempDir(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tempDir, err := NewTempDir("tempdir_test_")
	assert.Nil(err, "NewTempDir() should not return an error; err = %s", err)
	assert.DirExists(tempDir.Path, "Directory should exist after NewTempDir(); tempDir = %v", tempDir)

	tempDir.Cleanup()
	_, err = os.Lstat(tempDir.Path)
	assert.True(os.IsNotExist(err), "Directory should no longer exist after Cleanup(); tempDir = %v", tempDir)
}
