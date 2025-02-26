package kvstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/fileutil"
	"shpankids/infra/util/functional"
	"strings"
	"sync"
)

// FileSystemRawJsonStore represents a file system raw JSON store
type FileSystemRawJsonStore struct {
	rootDir        string
	filenamePrefix string
	mu             sync.RWMutex
}

func (s *FileSystemRawJsonStore) StreamAllNamespaces(_ context.Context) shpanstream.Stream[string] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	//todo:amit: do proper streaming, can do with scanner much nicer
	files, err := os.ReadDir(s.rootDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return shpanstream.EmptyStream[string]()
		}
		return shpanstream.NewErrorStream[string](err)
	}
	return shpanstream.Just(functional.MapSliceNoErr(files, func(file os.DirEntry) string {
		return file.Name()
	})...)

}

func (s *FileSystemRawJsonStore) StreamAllJson(_ context.Context, namespace string) shpanstream.Stream[functional.Entry[string, json.RawMessage]] {
	return shpanstream.NewStream[functional.Entry[string, json.RawMessage]](
		newFilesystemJsonStreamer(s.rootDir, s.filenamePrefix, &s.mu, namespace),
	)
}

func (s *FileSystemRawJsonStore) CreateSpaceStore(_ context.Context, spaceHierarchy []string) (RawJsonStore, error) {
	return NewFileSystemRawJsonStore(filepath.Join(append([]string{s.rootDir}, spaceHierarchy...)...))
}

func (s *FileSystemRawJsonStore) UnSetJSON(_ context.Context, namespace, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename, err := s.getSafeFilePath(namespace, key)
	if err != nil {
		return err
	}
	err = os.Remove(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("key %s:%s not found %w", namespace, key, err)
		}
		return err
	}
	return nil
}

func (s *FileSystemRawJsonStore) UnSetJSONIfExist(_ context.Context, namespace, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename, err := s.getSafeFilePath(namespace, key)
	if err != nil {
		return err
	}
	err = os.Remove(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return nil
}

// NewFileSystemRawJsonStore creates a new FileSystemRawJsonStore with the specified root directory
func NewFileSystemRawJsonStore(rootDir string) (*FileSystemRawJsonStore, error) {
	err := os.MkdirAll(rootDir, 0755) // 0755 is the permission mode for the directory
	if err != nil {
		return nil, err
	}
	return &FileSystemRawJsonStore{rootDir: rootDir, filenamePrefix: "key-"}, nil
}

// SetJSON stores JSON data with a given key and namespace
func (s *FileSystemRawJsonStore) SetJSON(_ context.Context, namespace, key string, json json.RawMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Join(s.rootDir, namespace)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	filename, err := s.getSafeFilePath(namespace, key)
	if err != nil {
		return err
	}

	return fileutil.WriteFileAtomically(filename, json)

	//file, err := os.Create(filename)
	//if err != nil {
	//	return err
	//}
	//defer file.Close()
	//
	//_, err = file.Write(json)
	//
	//// Make sure we flush
	//if err = file.Sync(); err != nil {
	//	return err
	//}

	//return err
}

func (s *FileSystemRawJsonStore) GetJSONIfExist(_ context.Context, namespace, key string) (*json.RawMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename, err := s.getSafeFilePath(namespace, key)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	bytes := json.RawMessage(data)
	return &bytes, nil

}

// GetJSON retrieves JSON data for a given key and namespace
func (s *FileSystemRawJsonStore) GetJSON(_ context.Context, namespace, key string) (json.RawMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename, err := s.getSafeFilePath(namespace, key)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("key %s:%s not found %w", namespace, key, err)
		}
		return nil, err
	}
	return data, nil
}

// ListAllJSON returns a slice of all JSON objects stored in the store for a given namespace
func (s *FileSystemRawJsonStore) ListAllJSON(_ context.Context, namespace string) (map[string]json.RawMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Join(s.rootDir, namespace)
	files, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]json.RawMessage{}, nil
		}
		return nil, err
	}

	var allJSON = map[string]json.RawMessage{}
	for _, file := range files {
		fileName := file.Name()
		key, isValid := strings.CutPrefix(fileName, s.filenamePrefix)
		if !file.IsDir() && filepath.Ext(fileName) == "" && isValid {
			data, err := os.ReadFile(filepath.Join(dir, fileName))
			if err != nil {
				return nil, err
			}
			allJSON[key] = data
		}
	}
	return allJSON, nil
}

func (s *FileSystemRawJsonStore) getSafeFilePath(namespace string, key string) (string, error) {
	// Ideally we'd just strip out invalid characters here (:?* etc. on windows, plus \/ on linux)
	// but there's code that relies on the key (filename) being the same as the id within the file ..
	// which isn't true if we modify it (e.g. filename = x, contents {"id":"x/"})

	if key == "" {
		return "", fmt.Errorf("[filesystem kvs] Cannot have an empty key for namespace %s", namespace)
	}

	filename := fmt.Sprintf("%s%s", s.filenamePrefix, key)

	if filepath.Clean(filename) != filename {
		slog.Warn("[filesystem kvs] Key has characters which modify the filepath", "namespace", namespace, "key", key)
		return "", fmt.Errorf("key contains invalid characters: %s", key)
	}

	if !filepath.IsLocal(filename) {
		slog.Warn("[filesystem kvs] Key is not a valid filename or causes directory traversal", "namespace", namespace, "key", key)
		return "", fmt.Errorf("key is not a valid filename: %s", key)
	}

	return filepath.Join(s.rootDir, namespace, filename), nil
}
