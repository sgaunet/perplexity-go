package perplexity_test

import (
	"testing"

	"github.com/sgaunet/perplexity-go/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewMessages(t *testing.T) {
	t.Run("creates a new Messages object", func(t *testing.T) {
		m := perplexity.NewMessages()
		assert.NotNil(t, m)
	})
}

func TestWithSystemMessage(t *testing.T) {
	t.Run("sets the system message for the Messages object", func(t *testing.T) {
		m := perplexity.NewMessages(perplexity.WithSystemMessage("system message"))
		sysMsg := m.GetSystemMessage()
		assert.Equal(t, sysMsg, "system message")
	})
}

func TestAddUserMessage(t *testing.T) {
	t.Run("adds a user message to the Messages object", func(t *testing.T) {
		m := perplexity.NewMessages()
		err := m.AddUserMessage("hello")
		assert.Nil(t, err)
		msgs := m.GetMessages()
		assert.Equal(t, len(msgs), 1)
		assert.Equal(t, msgs[0].Role, "user")
		assert.Equal(t, msgs[0].Content, "hello")
	})
}

func TestAddAgentMessage(t *testing.T) {
	t.Run("adds an assistant message to the Messages object", func(t *testing.T) {
		m := perplexity.NewMessages()
		m.AddUserMessage("hello")
		err := m.AddAgentMessage("hello")
		assert.Nil(t, err)
		msgs := m.GetMessages()
		assert.Equal(t, len(msgs), 2)
		assert.Equal(t, msgs[1].Role, "assistant")
		assert.Equal(t, msgs[1].Content, "hello")
	})
}

func TestAddTwiceUserMessage(t *testing.T) {
	t.Run("adds a user message to the Messages object", func(t *testing.T) {
		m := perplexity.NewMessages()
		err := m.AddUserMessage("hello")
		assert.Nil(t, err)
		err = m.AddUserMessage("hello")
		assert.NotNil(t, err)
	})
}

func TestAddTwiceAgentMessage(t *testing.T) {
	t.Run("adds an assistant message to the Messages object", func(t *testing.T) {
		m := perplexity.NewMessages()
		m.AddUserMessage("hello")
		err := m.AddAgentMessage("hello")
		assert.Nil(t, err)
		err = m.AddAgentMessage("hello")
		assert.NotNil(t, err)
	})
}

func TestNewTextContent(t *testing.T) {
	t.Run("creates text content", func(t *testing.T) {
		content := perplexity.NewTextContent("Hello world")
		assert.Equal(t, perplexity.ContentTypeText, content.Type)
		assert.NotNil(t, content.Text)
		assert.Equal(t, "Hello world", *content.Text)
		assert.Nil(t, content.ImageURL)
	})
}

func TestNewImageURLContent(t *testing.T) {
	t.Run("creates image URL content", func(t *testing.T) {
		url := "https://example.com/image.jpg"
		content := perplexity.NewImageURLContent(url)
		assert.Equal(t, perplexity.ContentTypeImageURL, content.Type)
		assert.Nil(t, content.Text)
		assert.NotNil(t, content.ImageURL)
		assert.Equal(t, url, content.ImageURL.URL)
	})
}

func TestAddMultimodalUserMessage(t *testing.T) {
	t.Run("adds multimodal user message with text and image", func(t *testing.T) {
		m := perplexity.NewMessages()
		contents := []perplexity.Content{
			perplexity.NewTextContent("What's in this image?"),
			perplexity.NewImageURLContent("https://example.com/image.jpg"),
		}

		err := m.AddMultimodalUserMessage(contents)
		assert.NoError(t, err)
		assert.True(t, m.IsMultimodal())
		assert.True(t, m.HasImages())

		multimodalMsgs := m.GetMultimodalMessages()
		assert.Len(t, multimodalMsgs, 1)
		assert.Equal(t, "user", multimodalMsgs[0].Role)
		assert.Len(t, multimodalMsgs[0].Content, 2)
	})

	t.Run("rejects empty content", func(t *testing.T) {
		m := perplexity.NewMessages()
		err := m.AddMultimodalUserMessage([]perplexity.Content{})
		assert.Error(t, err)
		assert.ErrorIs(t, err, perplexity.ErrMultimodalContentEmpty)
	})

	t.Run("validates content types", func(t *testing.T) {
		m := perplexity.NewMessages()
		invalidContent := perplexity.Content{
			Type: "invalid_type",
		}
		err := m.AddMultimodalUserMessage([]perplexity.Content{invalidContent})
		assert.Error(t, err)
		assert.ErrorIs(t, err, perplexity.ErrMultimodalInvalidContentType)
	})
}

func TestAddUserMessageWithImage(t *testing.T) {
	t.Run("adds user message with text and image URL", func(t *testing.T) {
		m := perplexity.NewMessages()
		err := m.AddUserMessageWithImage("Describe this image", "https://example.com/image.jpg")
		assert.NoError(t, err)
		assert.True(t, m.IsMultimodal())
		assert.True(t, m.HasImages())

		multimodalMsgs := m.GetMultimodalMessages()
		assert.Len(t, multimodalMsgs, 1)
		assert.Len(t, multimodalMsgs[0].Content, 2)
	})

	t.Run("rejects invalid image URL", func(t *testing.T) {
		m := perplexity.NewMessages()
		err := m.AddUserMessageWithImage("Describe this image", "http://example.com/image.jpg") // HTTP not HTTPS
		assert.Error(t, err)
	})
}

func TestAddMultimodalAgentMessage(t *testing.T) {
	t.Run("adds assistant message after user message", func(t *testing.T) {
		m := perplexity.NewMessages()

		// First add user message
		userContents := []perplexity.Content{
			perplexity.NewTextContent("What's in this image?"),
			perplexity.NewImageURLContent("https://example.com/image.jpg"),
		}
		err := m.AddMultimodalUserMessage(userContents)
		assert.NoError(t, err)

		// Then add assistant message
		assistantContents := []perplexity.Content{
			perplexity.NewTextContent("I can see a beautiful landscape in this image."),
		}
		err = m.AddMultimodalAgentMessage(assistantContents)
		assert.NoError(t, err)

		multimodalMsgs := m.GetMultimodalMessages()
		assert.Len(t, multimodalMsgs, 2)
		assert.Equal(t, "user", multimodalMsgs[0].Role)
		assert.Equal(t, "assistant", multimodalMsgs[1].Role)
	})

	t.Run("rejects assistant message as first message", func(t *testing.T) {
		m := perplexity.NewMessages()
		assistantContents := []perplexity.Content{
			perplexity.NewTextContent("Hello!"),
		}
		err := m.AddMultimodalAgentMessage(assistantContents)
		assert.Error(t, err)
		assert.ErrorIs(t, err, perplexity.ErrFirstMessageShouldBeUser)
	})
}

func TestGetMultimodalMessages(t *testing.T) {
	t.Run("includes system message when present", func(t *testing.T) {
		m := perplexity.NewMessages(perplexity.WithSystemMessage("You are a helpful assistant."))

		userContents := []perplexity.Content{
			perplexity.NewTextContent("Hello!"),
		}
		err := m.AddMultimodalUserMessage(userContents)
		assert.NoError(t, err)

		multimodalMsgs := m.GetMultimodalMessages()
		assert.Len(t, multimodalMsgs, 2)
		assert.Equal(t, "system", multimodalMsgs[0].Role)
		assert.Equal(t, "user", multimodalMsgs[1].Role)

		// Check system message content
		assert.Len(t, multimodalMsgs[0].Content, 1)
		assert.Equal(t, perplexity.ContentTypeText, multimodalMsgs[0].Content[0].Type)
		assert.Equal(t, "You are a helpful assistant.", *multimodalMsgs[0].Content[0].Text)
	})
}

func TestNewFileURLContent(t *testing.T) {
	t.Run("creates file URL content without filename", func(t *testing.T) {
		url := "https://example.com/document.pdf"
		content := perplexity.NewFileURLContent(url, "")
		assert.Equal(t, perplexity.ContentTypeFileURL, content.Type)
		assert.Nil(t, content.Text)
		assert.Nil(t, content.ImageURL)
		assert.NotNil(t, content.FileURL)
		assert.Equal(t, url, content.FileURL.URL)
		assert.Nil(t, content.FileName)
	})

	t.Run("creates file URL content with filename", func(t *testing.T) {
		url := "https://example.com/document.pdf"
		fileName := "report.pdf"
		content := perplexity.NewFileURLContent(url, fileName)
		assert.Equal(t, perplexity.ContentTypeFileURL, content.Type)
		assert.NotNil(t, content.FileURL)
		assert.Equal(t, url, content.FileURL.URL)
		assert.NotNil(t, content.FileName)
		assert.Equal(t, fileName, *content.FileName)
	})
}

func TestAddMultimodalUserMessageWithFile(t *testing.T) {
	t.Run("adds multimodal user message with text and file", func(t *testing.T) {
		m := perplexity.NewMessages()
		contents := []perplexity.Content{
			perplexity.NewTextContent("Summarize this document"),
			perplexity.NewFileURLContent("https://example.com/doc.pdf", ""),
		}

		err := m.AddMultimodalUserMessage(contents)
		assert.NoError(t, err)
		assert.True(t, m.IsMultimodal())
		assert.True(t, m.HasFiles())

		multimodalMsgs := m.GetMultimodalMessages()
		assert.Len(t, multimodalMsgs, 1)
		assert.Equal(t, "user", multimodalMsgs[0].Role)
		assert.Len(t, multimodalMsgs[0].Content, 2)
	})

	t.Run("validates file content type", func(t *testing.T) {
		m := perplexity.NewMessages()
		invalidContent := perplexity.Content{
			Type: "invalid_file_type",
		}
		err := m.AddMultimodalUserMessage([]perplexity.Content{invalidContent})
		assert.Error(t, err)
		assert.ErrorIs(t, err, perplexity.ErrMultimodalInvalidContentType)
	})
}

func TestAddUserMessageWithFile(t *testing.T) {
	t.Run("adds user message with text and file URL", func(t *testing.T) {
		m := perplexity.NewMessages()
		err := m.AddUserMessageWithFile("Analyze this document", "https://example.com/report.pdf", "report.pdf")
		assert.NoError(t, err)
		assert.True(t, m.IsMultimodal())
		assert.True(t, m.HasFiles())

		multimodalMsgs := m.GetMultimodalMessages()
		assert.Len(t, multimodalMsgs, 1)
		assert.Len(t, multimodalMsgs[0].Content, 2)
	})

	t.Run("rejects invalid file URL", func(t *testing.T) {
		m := perplexity.NewMessages()
		err := m.AddUserMessageWithFile("Analyze this document", "http://example.com/report.pdf", "") // HTTP not HTTPS
		assert.Error(t, err)
	})
}

func TestHasFiles(t *testing.T) {
	t.Run("returns false for non-multimodal messages", func(t *testing.T) {
		m := perplexity.NewMessages()
		err := m.AddUserMessage("Hello")
		assert.NoError(t, err)
		assert.False(t, m.HasFiles())
	})

	t.Run("returns false when no files present", func(t *testing.T) {
		m := perplexity.NewMessages()
		contents := []perplexity.Content{
			perplexity.NewTextContent("Hello"),
		}
		err := m.AddMultimodalUserMessage(contents)
		assert.NoError(t, err)
		assert.False(t, m.HasFiles())
	})

	t.Run("returns true when files present", func(t *testing.T) {
		m := perplexity.NewMessages()
		contents := []perplexity.Content{
			perplexity.NewTextContent("Analyze this"),
			perplexity.NewFileURLContent("https://example.com/doc.pdf", ""),
		}
		err := m.AddMultimodalUserMessage(contents)
		assert.NoError(t, err)
		assert.True(t, m.HasFiles())
	})
}

func TestAddMultimodalUserMessageMixedContent(t *testing.T) {
	t.Run("supports text, image, and file in same message", func(t *testing.T) {
		m := perplexity.NewMessages()
		contents := []perplexity.Content{
			perplexity.NewTextContent("Analyze this document and image"),
			perplexity.NewImageURLContent("https://example.com/chart.png"),
			perplexity.NewFileURLContent("https://example.com/report.pdf", "report.pdf"),
		}

		err := m.AddMultimodalUserMessage(contents)
		assert.NoError(t, err)
		assert.True(t, m.IsMultimodal())
		assert.True(t, m.HasImages())
		assert.True(t, m.HasFiles())

		multimodalMsgs := m.GetMultimodalMessages()
		assert.Len(t, multimodalMsgs, 1)
		assert.Len(t, multimodalMsgs[0].Content, 3)
		assert.Equal(t, perplexity.ContentTypeText, multimodalMsgs[0].Content[0].Type)
		assert.Equal(t, perplexity.ContentTypeImageURL, multimodalMsgs[0].Content[1].Type)
		assert.Equal(t, perplexity.ContentTypeFileURL, multimodalMsgs[0].Content[2].Type)
	})
}
