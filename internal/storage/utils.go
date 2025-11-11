package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/adrg/xdg"
)

// GetDataDir returns the cross-platform data directory
func GetDataDir() string {
	return filepath.Join(xdg.DataHome, "syntax", "projects")
}

// GetConfigDir returns the cross-platform config directory
func GetConfigDir() string {
	return filepath.Join(xdg.ConfigHome, "syntax")
}

// GenerateCharacterID generates a unique character ID
func GenerateCharacterID() string {
	bytes := make([]byte, 8) // 8 bytes = 16 hex chars
	rand.Read(bytes)
	return "char_" + hex.EncodeToString(bytes)
}

// GenerateLocationID generates a unique location ID
func GenerateLocationID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "loc_" + hex.EncodeToString(bytes)
}

// GenerateProjectID generates a unique project ID
func GenerateProjectID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "proj_" + hex.EncodeToString(bytes)
}

// GenerateSceneID generates a scene ID from chapter and scene number
func GenerateSceneID(chapter, scene int) string {
	return fmt.Sprintf("ch%02d_sc%02d", chapter, scene)
}

// SanitizeFilename sanitizes user input for use as filename
func SanitizeFilename(input string) string {
	// Remove path separators
	safe := strings.ReplaceAll(input, "/", "-")
	safe = strings.ReplaceAll(safe, "\\", "-")
	safe = strings.ReplaceAll(safe, "..", "")

	// Remove special characters
	safe = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '_' || r == ' ' {
			return r
		}
		return -1
	}, safe)

	// Limit length
	if len(safe) > 255 {
		safe = safe[:255]
	}

	return filepath.Clean(safe)
}

// AtomicWriteFile writes a file atomically to prevent corruption
func AtomicWriteFile(path string, data []byte, perm os.FileMode) error {
	// Write to temporary file first
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, perm); err != nil {
		return err
	}

	// Atomic rename (overwrites existing file)
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath) // Clean up temp file
		return err
	}

	return nil
}

// EnsureDir ensures a directory exists, creating it if necessary
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0700)
}
