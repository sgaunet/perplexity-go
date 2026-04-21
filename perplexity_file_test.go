package perplexity_test

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/sgaunet/perplexity-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileProcessor(t *testing.T) {
	t.Run("creates a new FileProcessor instance", func(t *testing.T) {
		processor := perplexity.NewFileProcessor()
		assert.NotNil(t, processor)
	})
}

func TestValidateFileFormat(t *testing.T) {
	processor := perplexity.NewFileProcessor()

	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"pdf format", "pdf", false},
		{"doc format", "doc", false},
		{"docx format", "docx", false},
		{"txt format", "txt", false},
		{"rtf format", "rtf", false},
		{"PDF uppercase", "PDF", false},
		{"DOCX uppercase", "DOCX", false},
		{"invalid format", "invalid", true},
		{"jpg format", "jpg", true},
		{"png format", "png", true},
		{"empty format", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.ValidateFileFormat(tt.format)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, perplexity.ErrFileFormatNotSupported)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFileSize(t *testing.T) {
	processor := perplexity.NewFileProcessor()

	tests := []struct {
		name    string
		size    int64
		wantErr bool
	}{
		{"zero size", 0, false},
		{"1KB file", 1024, false},
		{"1MB file", 1024 * 1024, false},
		{"10MB file", 10 * 1024 * 1024, false},
		{"exactly 50MB", perplexity.MaxFileSizeBytes, false},
		{"50MB + 1 byte", perplexity.MaxFileSizeBytes + 1, true},
		{"100MB file", 100 * 1024 * 1024, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.ValidateFileSize(tt.size)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, perplexity.ErrFileTooLarge)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFileURL(t *testing.T) {
	processor := perplexity.NewFileProcessor()

	payload := base64.StdEncoding.EncodeToString([]byte("file bytes"))

	tests := []struct {
		name    string
		url     string
		wantErr bool
		errType error
	}{
		{"valid HTTPS URL", "https://example.com/document.pdf", false, nil},
		{"valid HTTPS with path", "https://example.com/docs/report.pdf", false, nil},
		{"HTTP not allowed", "http://example.com/document.pdf", true, perplexity.ErrFileURLNotHTTPS},
		{"empty URL", "", true, perplexity.ErrFileURLInvalid},
		{"invalid URL", "not-a-url", true, perplexity.ErrFileURLNotHTTPS},
		{"no host", "https://", true, perplexity.ErrFileURLInvalid},
		{"ftp protocol", "ftp://example.com/file.pdf", true, perplexity.ErrFileURLNotHTTPS},

		// Valid data URIs for each supported format.
		{"valid data URI pdf", "data:application/pdf;base64," + payload, false, nil},
		{"valid data URI doc", "data:application/msword;base64," + payload, false, nil},
		{"valid data URI docx",
			"data:application/vnd.openxmlformats-officedocument.wordprocessingml.document;base64," + payload,
			false, nil},
		{"valid data URI txt", "data:text/plain;base64," + payload, false, nil},
		{"valid data URI rtf (application/rtf)", "data:application/rtf;base64," + payload, false, nil},
		{"valid data URI rtf (text/rtf)", "data:text/rtf;base64," + payload, false, nil},
		{"valid data URI mixed case mime", "data:APPLICATION/PDF;base64," + payload, false, nil},

		// Unsupported data URI mimes.
		{"data URI unsupported mime", "data:application/json;base64," + payload, true, perplexity.ErrFileFormatNotSupported},
		{"data URI image mime", "data:image/png;base64," + payload, true, perplexity.ErrFileFormatNotSupported},

		// Malformed data URIs.
		{"data URI empty payload", "data:application/pdf;base64,", true, perplexity.ErrFileDataURIMalformed},
		{"data URI empty mime", "data:;base64,YWJj", true, perplexity.ErrFileDataURIMalformed},
		{"data URI missing separator", "data:application/pdf,YWJj", true, perplexity.ErrFileDataURIMalformed},
		{"data URI invalid base64", "data:application/pdf;base64,!!!not-base64!!!", true, perplexity.ErrFileDataURIMalformed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.ValidateFileURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFileURLOversizedDataURI(t *testing.T) {
	processor := perplexity.NewFileProcessor()

	oversized := make([]byte, perplexity.MaxFileSizeBytes+1)
	uri := "data:application/pdf;base64," + base64.StdEncoding.EncodeToString(oversized)
	err := processor.ValidateFileURL(uri)
	assert.Error(t, err)
	assert.ErrorIs(t, err, perplexity.ErrFileTooLarge)
}

func TestEncodeFileFromPath(t *testing.T) {
	processor := perplexity.NewFileProcessor()

	t.Run("encodes valid txt file", func(t *testing.T) {
		// Create temp txt file
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		testContent := []byte("Hello, World!")
		err := os.WriteFile(testFile, testContent, 0600)
		require.NoError(t, err)

		// Encode file
		base64Data, fileName, err := processor.EncodeFileFromPath(testFile)
		assert.NoError(t, err)
		assert.NotEmpty(t, base64Data)
		assert.Equal(t, "test.txt", fileName)

		// Verify it's valid base64 (no data URI prefix)
		assert.NotContains(t, base64Data, "data:")
		assert.NotContains(t, base64Data, ";base64,")
	})

	t.Run("encodes valid pdf file", func(t *testing.T) {
		// Create temp pdf file (minimal valid PDF)
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.pdf")
		testContent := []byte("%PDF-1.4\n")
		err := os.WriteFile(testFile, testContent, 0600)
		require.NoError(t, err)

		// Encode file
		base64Data, fileName, err := processor.EncodeFileFromPath(testFile)
		assert.NoError(t, err)
		assert.NotEmpty(t, base64Data)
		assert.Equal(t, "test.pdf", fileName)
	})

	t.Run("fails with non-existent file", func(t *testing.T) {
		_, _, err := processor.EncodeFileFromPath("/nonexistent/file.txt")
		assert.Error(t, err)
		assert.ErrorIs(t, err, perplexity.ErrFileNotFound)
	})

	t.Run("fails with unsupported format", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.jpg")
		err := os.WriteFile(testFile, []byte("test"), 0600)
		require.NoError(t, err)

		_, _, err = processor.EncodeFileFromPath(testFile)
		assert.Error(t, err)
		assert.ErrorIs(t, err, perplexity.ErrFileFormatNotSupported)
	})

	t.Run("fails with file exceeding size limit", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "large.txt")

		// Create a file larger than 50MB
		largeContent := make([]byte, perplexity.MaxFileSizeBytes+1)
		err := os.WriteFile(testFile, largeContent, 0600)
		require.NoError(t, err)

		_, _, err = processor.EncodeFileFromPath(testFile)
		assert.Error(t, err)
		assert.ErrorIs(t, err, perplexity.ErrFileTooLarge)
	})

	t.Run("extracts filename from path", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "my-report.pdf")
		testContent := []byte("%PDF-1.4\n")
		err := os.WriteFile(testFile, testContent, 0600)
		require.NoError(t, err)

		_, fileName, err := processor.EncodeFileFromPath(testFile)
		assert.NoError(t, err)
		assert.Equal(t, "my-report.pdf", fileName)
	})
}

func TestCheckFileURLAccessibility(t *testing.T) {
	processor := perplexity.NewFileProcessor()
	ctx := context.Background()

	t.Run("rejects invalid URL", func(t *testing.T) {
		err := processor.CheckFileURLAccessibility(ctx, "http://example.com/file.pdf")
		assert.Error(t, err)
		assert.ErrorIs(t, err, perplexity.ErrFileURLNotHTTPS)
	})

	t.Run("rejects empty URL", func(t *testing.T) {
		err := processor.CheckFileURLAccessibility(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, perplexity.ErrFileURLInvalid)
	})
}

func TestSupportedFileFormats(t *testing.T) {
	t.Run("contains expected formats", func(t *testing.T) {
		formats := perplexity.SupportedFileFormats
		assert.Contains(t, formats, "pdf")
		assert.Contains(t, formats, "doc")
		assert.Contains(t, formats, "docx")
		assert.Contains(t, formats, "txt")
		assert.Contains(t, formats, "rtf")
		assert.Len(t, formats, 5)
	})
}

// TestValidateRequestWithLocalImageFile reproduces the bug report:
// NewImageFileContent produces a data URI, and req.Validate() must accept it.
// SearchRecencyFilter is explicitly cleared because NewCompletionRequest defaults it to "month",
// which is (separately) rejected by validateImageCompatibility — unrelated to this fix.
func TestValidateRequestWithLocalImageFile(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "local.jpg")
	require.NoError(t, os.WriteFile(imgPath, []byte("fake-jpeg-bytes"), 0600))

	msgs := perplexity.NewMessages()
	require.NoError(t, msgs.AddUserMessageWithImageFile("describe this", imgPath))

	req := perplexity.NewCompletionRequest(
		perplexity.WithMessagesFromMessages(&msgs),
		perplexity.WithModel("sonar-pro"),
		perplexity.WithSearchRecencyFilter(""),
	)
	assert.NoError(t, req.Validate())
}

// TestValidateRequestWithLocalFile reproduces the analogous file-path bug:
// NewFileFileContent produces a data URI and req.Validate() must accept it.
func TestValidateRequestWithLocalFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "report.pdf")
	require.NoError(t, os.WriteFile(filePath, []byte("%PDF-1.4\nfake"), 0600))

	msgs := perplexity.NewMessages()
	require.NoError(t, msgs.AddUserMessageWithFileFromPath("summarize", filePath))

	req := perplexity.NewCompletionRequest(
		perplexity.WithMessagesFromMessages(&msgs),
		perplexity.WithModel("sonar-pro"),
	)
	assert.NoError(t, req.Validate())
}
