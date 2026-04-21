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

// Image processing constants and errors.
const (
	// MaxImageSizeBytes is the maximum allowed image size (50MB) according to Perplexity API.
	MaxImageSizeBytes = 50 * 1024 * 1024
	// TokenEstimationDivisor is used to estimate token usage from image pixels
	// Based on empirical measurements of Perplexity API token consumption.
	TokenEstimationDivisor = 750
	// HTTPTimeoutSeconds is the timeout for HTTP requests to check image URL accessibility.
	HTTPTimeoutSeconds = 10
)

// Error definitions for image processing.
var (
	// ErrImageTooLarge is returned when an image exceeds the 50MB size limit.
	ErrImageTooLarge = errors.New("image size exceeds 50MB limit")

	// ErrImageFormatNotSupported is returned when an image format is not supported.
	ErrImageFormatNotSupported = errors.New("image format not supported (use PNG, JPEG, WEBP, or GIF)")

	// ErrImageURLNotHTTPS is returned when an image URL does not use HTTPS protocol.
	ErrImageURLNotHTTPS = errors.New("image URL must use HTTPS protocol")

	// ErrImageFileNotFound is returned when an image file cannot be found.
	ErrImageFileNotFound = errors.New("image file not found")

	// ErrImageReadFailed is returned when an image file cannot be read.
	ErrImageReadFailed = errors.New("failed to read image file")

	// ErrImageURLInvalid is returned when an image URL is malformed.
	ErrImageURLInvalid = errors.New("invalid image URL format")

	// ErrImageDataURIMalformed is returned when an image data URI is malformed.
	ErrImageDataURIMalformed = errors.New("invalid image data URI")

	// ErrImageURLBadStatus is returned when image URL returns a non-2xx status.
	ErrImageURLBadStatus = errors.New("image URL returned bad status")

	// ErrInvalidContentType is returned when image content type is invalid.
	ErrInvalidContentType = errors.New("invalid content type")
)

// SupportedImageFormats contains the image formats supported by the Perplexity API.
var SupportedImageFormats = []string{"png", "jpeg", "jpg", "webp", "gif"}

// ImageProcessor handles image encoding and validation for the Perplexity API.
type ImageProcessor struct{}

// NewImageProcessor creates a new ImageProcessor instance.
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
}

// EncodeImageFromFile reads an image file, validates it, and returns a base64 data URI.
// The file path should point to a valid image file in one of the supported formats.
// Returns a data URI in the format: "data:image/[format];base64,[encoded_data]".
func (p *ImageProcessor) EncodeImageFromFile(filepath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return "", ErrImageFileNotFound
	}

	// Get file info for size validation
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrImageReadFailed, err)
	}

	// Validate file size
	if err := p.ValidateImageSize(fileInfo.Size()); err != nil {
		return "", err
	}

	// Validate file format based on extension
	format := p.getImageFormatFromPath(filepath)
	if err := p.ValidateImageFormat(format); err != nil {
		return "", err
	}

	// Read file content
	fileData, err := os.ReadFile(filepath) //nolint:gosec // G304: File path comes from validated user input for image processing
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrImageReadFailed, err)
	}

	// Encode to base64 and create data URI
	base64Data := base64.StdEncoding.EncodeToString(fileData)
	mimeType := p.getMimeType(format)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

	return dataURI, nil
}

// ValidateImageURL validates that a URL is HTTPS (or a base64 data URI) and has a valid format.
// Accepts either an https:// URL or a data:image/<fmt>;base64,<payload> URI, matching what the
// Perplexity API accepts and what EncodeImageFromFile produces.
func (p *ImageProcessor) ValidateImageURL(imageURL string) error {
	if imageURL == "" {
		return ErrImageURLInvalid
	}

	if strings.HasPrefix(imageURL, "data:") {
		return p.validateImageDataURI(imageURL)
	}

	// Parse the URL
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrImageURLInvalid, err)
	}

	// Validate HTTPS scheme
	if parsedURL.Scheme != "https" {
		return ErrImageURLNotHTTPS
	}

	// Validate host is present
	if parsedURL.Host == "" {
		return ErrImageURLInvalid
	}

	return nil
}

// ValidateImageFormat checks if the image format is supported by the Perplexity API.
func (p *ImageProcessor) ValidateImageFormat(format string) error {
	format = strings.ToLower(format)
	if slices.Contains(SupportedImageFormats, format) {
		return nil
	}
	return ErrImageFormatNotSupported
}

// ValidateImageSize checks if the image size is within the 50MB limit.
func (p *ImageProcessor) ValidateImageSize(size int64) error {
	if size > MaxImageSizeBytes {
		return ErrImageTooLarge
	}
	return nil
}

// CheckImageURLAccessibility performs a HEAD request to verify the image URL is accessible.
// This is an optional validation that can be used to verify URLs before sending to the API.
func (p *ImageProcessor) CheckImageURLAccessibility(ctx context.Context, imageURL string) error {
	if err := p.ValidateImageURL(imageURL); err != nil {
		return err
	}

	// Perform HEAD request to check if URL is accessible
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, imageURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req) //nolint:gosec // URL is validated as HTTPS before this call
	if err != nil {
		return fmt.Errorf("image URL not accessible: %w", err)
	}
	defer resp.Body.Close()

	// Check if response is successful
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ErrImageURLBadStatus
	}

	// Optionally validate content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !p.isValidImageContentType(contentType) {
		return ErrInvalidContentType
	}

	return nil
}

// EstimateTokenUsage estimates the token usage for an image based on its dimensions.
// According to Perplexity documentation: tokens = (width px × height px) / 750
// This function requires image dimensions to be provided separately.
func (p *ImageProcessor) EstimateTokenUsage(width, height int) int {
	if width <= 0 || height <= 0 {
		return 0
	}
	return (width * height) / TokenEstimationDivisor
}

// validateImageDataURI validates a data:image/<fmt>;base64,<payload> URI.
func (p *ImageProcessor) validateImageDataURI(uri string) error {
	mime, decoded, err := parseBase64DataURI(uri)
	if err != nil {
		return ErrImageDataURIMalformed
	}

	// Mime must start with "image/" and the suffix must be a supported format.
	format, ok := strings.CutPrefix(mime, "image/")
	if !ok {
		return ErrImageFormatNotSupported
	}
	if err := p.ValidateImageFormat(format); err != nil {
		return err
	}

	return p.ValidateImageSize(int64(len(decoded)))
}

// getImageFormatFromPath extracts the image format from a file path.
func (p *ImageProcessor) getImageFormatFromPath(path string) string {
	ext := filepath.Ext(path)
	if ext != "" {
		return strings.ToLower(ext[1:]) // Remove the dot and convert to lowercase
	}
	return ""
}

// getMimeType returns the MIME type for a given image format.
func (p *ImageProcessor) getMimeType(format string) string {
	switch strings.ToLower(format) {
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "webp":
		return "image/webp"
	case "gif":
		return "image/gif"
	default:
		return "image/" + format
	}
}

// errDataURIMalformed is an internal sentinel for parseBase64DataURI failures.
// Callers map this to their module-specific public error (ErrImageDataURIMalformed, ErrFileDataURIMalformed).
var errDataURIMalformed = errors.New("malformed data URI")

// parseBase64DataURI parses a "data:<mime>;base64,<payload>" URI.
// Returns the lowercased mime and decoded bytes, or errDataURIMalformed on any parse failure
// (missing prefix, missing ";base64," separator, empty mime, empty payload, invalid base64).
func parseBase64DataURI(uri string) (string, []byte, error) {
	const prefix = "data:"
	const sep = ";base64,"

	if !strings.HasPrefix(uri, prefix) {
		return "", nil, errDataURIMalformed
	}
	rest := uri[len(prefix):]

	rawMime, payload, ok := strings.Cut(rest, sep)
	if !ok {
		return "", nil, errDataURIMalformed
	}
	mime := strings.ToLower(rawMime)
	if mime == "" || payload == "" {
		return "", nil, errDataURIMalformed
	}

	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", nil, errDataURIMalformed
	}
	return mime, decoded, nil
}

// isValidImageContentType checks if a content type corresponds to a supported image format.
func (p *ImageProcessor) isValidImageContentType(contentType string) bool {
	validTypes := []string{
		"image/png",
		"image/jpeg",
		"image/webp",
		"image/gif",
	}

	contentType = strings.ToLower(contentType)
	for _, validType := range validTypes {
		if strings.HasPrefix(contentType, validType) {
			return true
		}
	}
	return false
}
