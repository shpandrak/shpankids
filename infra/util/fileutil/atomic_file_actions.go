package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// WriteFileAtomicallyWithChmod writes data to filename+some suffix, then renames it into filename.
// The perm argument is ignored on Windows. If the target filename already
// exists but is not a regular file, returns an error.
func WriteFileAtomicallyWithChmod(filename string, data []byte, perm os.FileMode) (err error) {
	fi, err := os.Stat(filename)
	if err == nil && !fi.Mode().IsRegular() {
		return fmt.Errorf("%s already exists and is not a regular file", filename)
	}
	f, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".tmp")
	if err != nil {
		return err
	}
	tmpName := f.Name()
	defer func() {
		if err != nil {
			// Ignoring errors here, no much we can do...
			_ = f.Close()
			_ =os.Remove(tmpName)
		}
	}()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if err := f.Chmod(perm); err != nil {
			return err
		}
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, filename)
}

// WriteFileAtomically writes data to filename+some suffix, then renames it into filename.
//If the target filename already exists but is not a regular file, returns an error.
func WriteFileAtomically(filename string, data []byte) (err error) {
	fi, err := os.Stat(filename)
	if err == nil && !fi.Mode().IsRegular() {
		return fmt.Errorf("%s already exists and is not a regular file", filename)
	}
	f, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".tmp")
	if err != nil {
		return err
	}
	tmpName := f.Name()
	defer func() {
		if err != nil {
			// Ignoring errors here, no much we can do...
			_ = f.Close()
			_ =os.Remove(tmpName)
		}
	}()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, filename)
}