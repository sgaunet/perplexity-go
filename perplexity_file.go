package perplexity

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// File processing constants and errors.
const (
	// MaxFileSizeBytes is the maximum allowed file size (50MB) according to Perplexity API.
	MaxFileSizeBytes = 50 * 1024 * 1024
	// HTTPFileTimeoutSeconds is the timeout for HTTP requests to check file URL accessibility.
	HTTPFileTimeoutSeconds = 10
)

// Error definitions for file processing.
var (
	// ErrFileTooLarge is returned when a file exceeds the 50MB size limit.
	ErrFileTooLarge = errors.New("file size exceeds 50MB limit")

	// ErrFileFormatNotSupported is returned when a file format is not supported.
	ErrFileFormatNotSupported = errors.New("file format not supported (use PDF, DOC, DOCX, TXT, or RTF)")

	// ErrFileURLNotHTTPS is returned when a file URL does not use HTTPS protocol.
	ErrFileURLNotHTTPS = errors.New("file URL must use HTTPS protocol")

	// ErrFileNotFound is returned when a file cannot be found.
	ErrFileNotFound = errors.New("file not found")

	// ErrFileReadFailed is returned when a file cannot be read.
	ErrFileReadFailed = errors.New("failed to read file")

	// ErrFileURLInvalid is returned when a file URL is malformed.
	ErrFileURLInvalid = errors.New("invalid file URL format")

	// ErrFileURLBadStatus is returned when file URL returns a non-2xx status.
	ErrFileURLBadStatus = errors.New("file URL returned bad status")

	// ErrInvalidFileContentType is returned when file content type is invalid.
	ErrInvalidFileContentType = errors.New("invalid file content type")
)

// SupportedFileFormats contains the file formats supported by the Perplexity API.
var SupportedFileFormats = []string{"pdf", "doc", "docx", "txt", "rtf"}

// FileProcessor handles file encoding and validation for the Perplexity API.
type FileProcessor struct{}

// NewFileProcessor creates a new FileProcessor instance.
func NewFileProcessor() *FileProcessor {
	return &FileProcessor{}
}

// EncodeFileFromPath reads a file, validates it, and returns base64 encoded data and filename.
// The file path should point to a valid file in one of the supported formats.
// Returns (base64EncodedData, fileName, error).
// IMPORTANT: Returns raw base64 data WITHOUT data URI prefix (unlike image encoding).
func (p *FileProcessor) EncodeFileFromPath(filePath string) (string, string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", "", ErrFileNotFound
	}

	// Get file info for size validation
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", "", fmt.Errorf("%w: %w", ErrFileReadFailed, err)
	}

	// Validate file size
	if err := p.ValidateFileSize(fileInfo.Size()); err != nil {
		return "", "", err
	}

	// Validate file format based on extension
	format := p.getFileFormatFromPath(filePath)
	if err := p.ValidateFileFormat(format); err != nil {
		return "", "", err
	}

	// Read file content
	fileData, err := os.ReadFile(filePath) //nolint:gosec // G304: File path comes from validated user input for file processing
	if err != nil {
		return "", "", fmt.Errorf("%w: %w", ErrFileReadFailed, err)
	}

	// Encode to base64 (raw base64, no data URI prefix)
	base64Data := base64.StdEncoding.EncodeToString(fileData)

	// Extract filename from path
	fileName := filepath.Base(filePath)

	return base64Data, fileName, nil
}

// ValidateFileURL validates that a URL is HTTPS and has a valid format.
// The URL must use HTTPS protocol to be accepted by the Perplexity API.
func (p *FileProcessor) ValidateFileURL(fileURL string) error {
	if fileURL == "" {
		return ErrFileURLInvalid
	}

	// Parse the URL
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFileURLInvalid, err)
	}

	// Validate HTTPS scheme
	if parsedURL.Scheme != "https" {
		return ErrFileURLNotHTTPS
	}

	// Validate host is present
	if parsedURL.Host == "" {
		return ErrFileURLInvalid
	}

	return nil
}

// ValidateFileFormat checks if the file format is supported by the Perplexity API.
func (p *FileProcessor) ValidateFileFormat(format string) error {
	format = strings.ToLower(format)
	if slices.Contains(SupportedFileFormats, format) {
		return nil
	}
	return ErrFileFormatNotSupported
}

// ValidateFileSize checks if the file size is within the 50MB limit.
func (p *FileProcessor) ValidateFileSize(size int64) error {
	if size > MaxFileSizeBytes {
		return ErrFileTooLarge
	}
	return nil
}

// CheckFileURLAccessibility performs a HEAD request to verify the file URL is accessible.
// This is an optional validation that can be used to verify URLs before sending to the API.
func (p *FileProcessor) CheckFileURLAccessibility(ctx context.Context, fileURL string) error {
	if err := p.ValidateFileURL(fileURL); err != nil {
		return err
	}

	// Perform HEAD request to check if URL is accessible
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, fileURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("file URL not accessible: %w", err)
	}
	defer resp.Body.Close()

	// Check if response is successful
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ErrFileURLBadStatus
	}

	// Optionally validate content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !p.isValidFileContentType(contentType) {
		return ErrInvalidFileContentType
	}

	return nil
}

// getFileFormatFromPath extracts the file format from a file path.
func (p *FileProcessor) getFileFormatFromPath(path string) string {
	ext := filepath.Ext(path)
	if ext != "" {
		return strings.ToLower(ext[1:]) // Remove the dot and convert to lowercase
	}
	return ""
}

// isValidFileContentType checks if a content type corresponds to a supported file format.
func (p *FileProcessor) isValidFileContentType(contentType string) bool {
	validTypes := []string{
		"application/pdf",
		"application/msword",                                                       // .doc
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document", // .docx
		"text/plain",                                                               // .txt
		"text/rtf",                                                                 // .rtf
		"application/rtf",                                                          // .rtf alternative
	}

	contentType = strings.ToLower(contentType)
	for _, validType := range validTypes {
		if strings.HasPrefix(contentType, validType) {
			return true
		}
	}
	return false
}
