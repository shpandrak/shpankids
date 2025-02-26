package kvstore

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"shpankids/infra/util/functional"
	"strings"
	"sync"
)

type filesystemJsonStreamer struct {
	rootDir        string
	dir            string
	fileChan       chan os.DirEntry
	fMu            *sync.RWMutex
	filenamePrefix string
}

func newFilesystemJsonStreamer(
	rootDir string,
	filenamePrefix string,
	fMu *sync.RWMutex,
	namespace string,
) *filesystemJsonStreamer {
	return &filesystemJsonStreamer{
		rootDir: rootDir,
		dir:     filepath.Join(rootDir, namespace),

		fMu:            fMu,
		fileChan:       make(chan os.DirEntry),
		filenamePrefix: filenamePrefix,
	}
}

func (fsp *filesystemJsonStreamer) Open(ctx context.Context) error {
	fsp.fMu.RLock()
	defer fsp.fMu.RUnlock()

	files, err := os.ReadDir(fsp.dir)
	if err != nil {
		close(fsp.fileChan)
		if os.IsNotExist(err) {
			// Nothing there, all good
			return nil
		}
		return err
	}

	go func() {
		defer close(fsp.fileChan)
		for _, file := range files {
			if !file.IsDir() {
				select {
				case <-ctx.Done():
					return
				case fsp.fileChan <- file:
				}
			}
		}
	}()
	return nil
}

func (fsp *filesystemJsonStreamer) Close() {
	// Nothing to do, the channel is closed when the stream is done as part of the Open method
}

func (fsp *filesystemJsonStreamer) Emit(ctx context.Context) (*functional.Entry[string, json.RawMessage], error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case file, ok := <-fsp.fileChan:
		if !ok {
			return nil, io.EOF
		}

		fileName := file.Name()
		key, isValid := strings.CutPrefix(fileName, fsp.filenamePrefix)
		if isValid && filepath.Ext(fileName) == "" {
			fsp.fMu.RLock()
			defer fsp.fMu.RUnlock()
			data, err := os.ReadFile(filepath.Join(fsp.dir, fileName))
			if err != nil {
				return nil, err
			}
			return &functional.Entry[string, json.RawMessage]{Key: key, Value: data}, nil
		}
		return nil, nil
	}
}
