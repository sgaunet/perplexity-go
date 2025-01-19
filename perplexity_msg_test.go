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
