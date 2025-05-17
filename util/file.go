package util

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func Exists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false // File doesn't exist
		}
		// For other errors like permission issues, log but still return false
		logrus.Errorf("Error checking if file %s exists: %v", path, err)
		return false
	}
	return true
}

func RemoveFile(fileName string) error {
	if fileName == "" {
		return errors.New("empty file name provided")
	}

	if Exists(fileName) {
		if err := os.Remove(fileName); err != nil {
			logrus.Errorf("Failed to delete file %s: %v", fileName, err)
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}
	return nil
}

// ReadLinesFromBottom Read the file contents sequentially from bottom to top and return the specified number of lines
func ReadLinesFromBottom(filePath string, numLines int) ([]string, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	// Read the file contents line by line and reverse the order of the lines
	total := 0
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		total++
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}

	// Reverse row order
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}

	// Returns the specified number of rows
	if len(lines) < numLines {
		numLines = len(lines)
	}
	return lines[:numLines], total, nil
}

func FindFile(dir, filename string) (string, error) {
	var result string
	var foundErr = errors.New("file found")
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == filename {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			result = absPath
			return foundErr
		}
		return nil
	})
	if err != nil && err != foundErr {
		return "", err
	}
	if result == "" {
		return "", fmt.Errorf("file %s not found in directory %s", filename, dir)
	}
	return result, nil
}
