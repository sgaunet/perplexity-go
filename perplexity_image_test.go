package perplexity_test

import (
	"testing"

	"github.com/sgaunet/perplexity-go/v2"
	"github.com/stretchr/testify/assert"
)

func TestImageProcessor(t *testing.T) {
	processor := perplexity.NewImageProcessor()

	t.Run("ValidateImageFormat with supported formats", func(t *testing.T) {
		supportedFormats := []string{"png", "jpeg", "jpg", "webp", "gif"}
		for _, format := range supportedFormats {
			err := processor.ValidateImageFormat(format)
			assert.NoError(t, err, "Format %s should be supported", format)
		}
	})

	t.Run("ValidateImageFormat with unsupported formats", func(t *testing.T) {
		unsupportedFormats := []string{"bmp", "tiff", "svg", "ico"}
		for _, format := range unsupportedFormats {
			err := processor.ValidateImageFormat(format)
			assert.Error(t, err, "Format %s should not be supported", format)
			assert.ErrorIs(t, err, perplexity.ErrImageFormatNotSupported)
		}
	})

	t.Run("ValidateImageSize within limits", func(t *testing.T) {
		err := processor.ValidateImageSize(1024 * 1024) // 1MB
		assert.NoError(t, err)

		err = processor.ValidateImageSize(perplexity.MaxImageSizeBytes) // Exactly 50MB
		assert.NoError(t, err)
	})

	t.Run("ValidateImageSize exceeds limits", func(t *testing.T) {
		err := processor.ValidateImageSize(perplexity.MaxImageSizeBytes + 1) // Over 50MB
		assert.Error(t, err)
		assert.ErrorIs(t, err, perplexity.ErrImageTooLarge)
	})

	t.Run("ValidateImageURL with HTTPS", func(t *testing.T) {
		validURLs := []string{
			"https://example.com/image.jpg",
			"https://cdn.example.com/path/to/image.png",
		}
		for _, url := range validURLs {
			err := processor.ValidateImageURL(url)
			assert.NoError(t, err, "URL %s should be valid", url)
		}
	})

	t.Run("ValidateImageURL with invalid URLs", func(t *testing.T) {
		invalidURLs := []string{
			"http://example.com/image.jpg", // HTTP not HTTPS
			"ftp://example.com/image.jpg",  // Wrong protocol
			"not-a-url",                    // Not a URL
			"",                             // Empty
		}
		for _, url := range invalidURLs {
			err := processor.ValidateImageURL(url)
			assert.Error(t, err, "URL %s should be invalid", url)
		}
	})

	t.Run("EstimateTokenUsage", func(t *testing.T) {
		// Test token calculation: (width × height) / 750
		tokens := processor.EstimateTokenUsage(1500, 1000)
		expected := (1500 * 1000) / 750
		assert.Equal(t, expected, tokens)

		// Test zero dimensions
		tokens = processor.EstimateTokenUsage(0, 1000)
		assert.Equal(t, 0, tokens)

		tokens = processor.EstimateTokenUsage(1000, 0)
		assert.Equal(t, 0, tokens)
	})
}
