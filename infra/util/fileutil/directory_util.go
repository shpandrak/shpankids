package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// MoveDirectoryIfExists moves a directory if it exists, doing nothing if it doesn't exist.
func MoveDirectoryIfExists(src, dest string) error {
	// Check if the source directory exists
	info, err := os.Stat(src)
	if os.IsNotExist(err) {
		// If the directory does not exist, do nothing, all good
		return nil
	}
	if err != nil {
		return fmt.Errorf("error checking source directory %s with the intention of moving it to %s: %w", src, dest, err)
	}

	// Ensure it is a directory
	if !info.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}

	// Move the directory
	err = os.Rename(src, dest)
	if err != nil {
		return fmt.Errorf("failed to move directory from %s to %s : %w", src, dest, err)
	}

	return nil
}

// MoveFileIfExists moves a file if it exists, doing nothing if it doesn't exist.
func MoveFileIfExists(src, dest string) error {
	// Check if the source file exists
	info, err := os.Stat(src)
	if os.IsNotExist(err) {
		// If the file does not exist, do nothing
		return nil
	}
	if err != nil {
		return fmt.Errorf("error checking source file %s for moving it into %s: %w", src, dest, err)
	}

	// Ensure it is a file, not a directory
	if info.IsDir() {
		return fmt.Errorf("source is a directory, not a file: %s", src)
	}

	// Ensure the destination directory exists, create it if necessary
	destDir := filepath.Dir(dest)
	err = os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}


	// Move the file
	err = os.Rename(src, dest)
	if err != nil {
		return fmt.Errorf("failed to move file %s to %s: %w", src,dest, err)
	}

	return nil
}
