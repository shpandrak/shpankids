package kvstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSafeFilename_ok(t *testing.T) {
	s := FileSystemRawJsonStore{rootDir: "/test", filenamePrefix: "key-"}

	path, err := s.getSafeFilePath("namespace", "my_key")
	assert.Nil(t, err)
	assert.Equal(t, "/test/namespace/key-my_key", path)

	path, err = s.getSafeFilePath("namespace", "......")
	assert.Nil(t, err)
	assert.Equal(t, "/test/namespace/key-......", path)

	path, err = s.getSafeFilePath("namespace", "my.key")
	assert.Nil(t, err)
	assert.Equal(t, "/test/namespace/key-my.key", path)

	path, err = s.getSafeFilePath("namespace", "abc123A-_")
	assert.Nil(t, err)
	assert.Equal(t, "/test/namespace/key-abc123A-_", path)
}

func TestGetSafeFilePath_rejects_traversal(t *testing.T) {
	s := FileSystemRawJsonStore{rootDir: "/test", filenamePrefix: "key-"}

	_, err := s.getSafeFilePath("namespace", "/../../secret-stuff")
	assert.Error(t, err)

	_, err = s.getSafeFilePath("namespace", "/./anotherfile")
	assert.Error(t, err)

	_, err = s.getSafeFilePath("namespace", "/../key-/anotherfile")
	assert.Error(t, err)
}

func TestGetSafeFilename_rejects_empty_key(t *testing.T) {
	s := FileSystemRawJsonStore{rootDir: "/test", filenamePrefix: "key-"}

	_, err := s.getSafeFilePath("namespace", "")
	assert.Error(t, err)
}
