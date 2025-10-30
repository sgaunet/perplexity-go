package perplexity_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sgaunet/perplexity-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAsyncJobRequest(t *testing.T) {
	t.Run("creates request with default model", func(t *testing.T) {
		req := perplexity.NewAsyncJobRequest()
		assert.Equal(t, perplexity.ModelSonarDeepResearch, req.Model)
	})

	t.Run("applies functional options", func(t *testing.T) {
		messages := []perplexity.Message{
			{Role: "user", Content: "Test message"},
		}
		req := perplexity.NewAsyncJobRequest(
			perplexity.WithAsyncMessages(messages),
			perplexity.WithAsyncReasoningEffort("high"),
			perplexity.WithAsyncMaxTokens(1000),
		)
		assert.Equal(t, messages, req.Messages)
		assert.Equal(t, "high", req.ReasoningEffort)
		assert.Equal(t, 1000, req.MaxTokens)
	})
}

func TestAsyncJobRequestValidation(t *testing.T) {
	tests := []struct {
		name        string
		setupReq    func() *perplexity.AsyncJobRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			setupReq: func() *perplexity.AsyncJobRequest {
				return perplexity.NewAsyncJobRequest(
					perplexity.WithAsyncMessages([]perplexity.Message{
						{Role: "user", Content: "Test"},
					}),
				)
			},
			expectError: false,
		},
		{
			name: "nil request",
			setupReq: func() *perplexity.AsyncJobRequest {
				return nil
			},
			expectError: true,
			errorMsg:    "request cannot be nil",
		},
		{
			name: "wrong model",
			setupReq: func() *perplexity.AsyncJobRequest {
				req := perplexity.NewAsyncJobRequest()
				req.Model = "sonar"
				return req
			},
			expectError: true,
			errorMsg:    "async jobs require sonar-deep-research model",
		},
		{
			name: "no messages",
			setupReq: func() *perplexity.AsyncJobRequest {
				return perplexity.NewAsyncJobRequest()
			},
			expectError: true,
			errorMsg:    "at least one message is required",
		},
		{
			name: "invalid reasoning effort",
			setupReq: func() *perplexity.AsyncJobRequest {
				return perplexity.NewAsyncJobRequest(
					perplexity.WithAsyncMessages([]perplexity.Message{
						{Role: "user", Content: "Test"},
					}),
					perplexity.WithAsyncReasoningEffort("invalid"),
				)
			},
			expectError: true,
			errorMsg:    "reasoning effort must be 'low', 'medium', or 'high'",
		},
		{
			name: "invalid temperature",
			setupReq: func() *perplexity.AsyncJobRequest {
				return perplexity.NewAsyncJobRequest(
					perplexity.WithAsyncMessages([]perplexity.Message{
						{Role: "user", Content: "Test"},
					}),
					perplexity.WithAsyncTemperature(3.0),
				)
			},
			expectError: true,
			errorMsg:    "temperature must be between 0 and 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			err := perplexity.ValidateAsyncJobRequest(req)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateAsyncJob(t *testing.T) {
	t.Run("successful job creation", func(t *testing.T) {
		mockResponse := &perplexity.AsyncJobResponse{
			ID:        "job_123",
			Status:    perplexity.StatusPending,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/async/chat/completions", r.URL.Path)
			assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client := perplexity.NewClient("test-key")
		client.SetAsyncEndpoint(server.URL + "/async/chat/completions")

		req := perplexity.NewAsyncJobRequest(
			perplexity.WithAsyncMessages([]perplexity.Message{
				{Role: "user", Content: "Test async request"},
			}),
		)

		result, err := client.CreateAsyncJob(req)
		require.NoError(t, err)
		assert.Equal(t, "job_123", result.ID)
		assert.Equal(t, perplexity.StatusPending, result.Status)
	})

	t.Run("validation error for wrong model", func(t *testing.T) {
		client := perplexity.NewClient("test-key")
		req := perplexity.NewAsyncJobRequest()
		req.Model = "sonar" // Wrong model

		_, err := client.CreateAsyncJob(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "async jobs require sonar-deep-research model")
	})

	t.Run("nil request error", func(t *testing.T) {
		client := perplexity.NewClient("test-key")
		_, err := client.CreateAsyncJob(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request must not be nil")
	})
}

func TestGetAsyncJob(t *testing.T) {
	t.Run("successful job retrieval", func(t *testing.T) {
		mockResponse := &perplexity.AsyncJobResponse{
			ID:        "job_123",
			Status:    perplexity.StatusCompleted,
			CreatedAt: time.Now().Add(-5 * time.Minute),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			Result: &perplexity.CompletionResponse{
				ID: "completion_456",
				Choices: []perplexity.Choice{
					{
						Index:        0,
						Message:      perplexity.Message{Role: "assistant", Content: "Test response"},
						FinishReason: "stop",
					},
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/async/chat/completions/job_123", r.URL.Path)
			assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client := perplexity.NewClient("test-key")
		client.SetAsyncEndpoint(server.URL + "/async/chat/completions")

		result, err := client.GetAsyncJob("job_123")
		require.NoError(t, err)
		assert.Equal(t, "job_123", result.ID)
		assert.Equal(t, perplexity.StatusCompleted, result.Status)
		assert.NotNil(t, result.Result)
		assert.Equal(t, "completion_456", result.Result.ID)
	})

	t.Run("job not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": {"message": "Job not found"}}`))
		}))
		defer server.Close()

		client := perplexity.NewClient("test-key")
		client.SetAsyncEndpoint(server.URL + "/async/chat/completions")

		_, err := client.GetAsyncJob("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "async job not found")
	})

	t.Run("empty job ID", func(t *testing.T) {
		client := perplexity.NewClient("test-key")
		_, err := client.GetAsyncJob("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job ID cannot be empty")
	})
}

func TestListAsyncJobs(t *testing.T) {
	t.Run("successful job listing", func(t *testing.T) {
		mockResponse := &perplexity.AsyncJobListResponse{
			Jobs: []perplexity.AsyncJobResponse{
				{
					ID:        "job_1",
					Status:    perplexity.StatusCompleted,
					CreatedAt: time.Now().Add(-10 * time.Minute),
				},
				{
					ID:        "job_2",
					Status:    perplexity.StatusProcessing,
					CreatedAt: time.Now().Add(-5 * time.Minute),
				},
			},
			Total:   2,
			Limit:   20,
			Offset:  0,
			HasMore: false,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/async/chat/completions", r.URL.Path)
			assert.Equal(t, "limit=20&offset=0", r.URL.RawQuery)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client := perplexity.NewClient("test-key")
		client.SetAsyncEndpoint(server.URL + "/async/chat/completions")

		result, err := client.ListAsyncJobs(20, 0)
		require.NoError(t, err)
		assert.Len(t, result.Jobs, 2)
		assert.Equal(t, 2, result.Total)
		assert.False(t, result.HasMore)
	})

	t.Run("default parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "limit=20&offset=0", r.URL.RawQuery)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(&perplexity.AsyncJobListResponse{
				Jobs: []perplexity.AsyncJobResponse{},
			})
		}))
		defer server.Close()

		client := perplexity.NewClient("test-key")
		client.SetAsyncEndpoint(server.URL + "/async/chat/completions")

		_, err := client.ListAsyncJobs(0, -1) // Should use defaults
		assert.NoError(t, err)
	})
}

func TestAsyncJobResponse_Methods(t *testing.T) {
	t.Run("IsCompleted method", func(t *testing.T) {
		tests := []struct {
			status   perplexity.AsyncJobStatus
			expected bool
		}{
			{perplexity.StatusPending, false},
			{perplexity.StatusProcessing, false},
			{perplexity.StatusCompleted, true},
			{perplexity.StatusFailed, true},
			{perplexity.StatusExpired, false}, // Changed: expired is not considered "completed"
		}

		for _, tt := range tests {
			t.Run(string(tt.status), func(t *testing.T) {
				job := &perplexity.AsyncJobResponse{Status: tt.status}
				assert.Equal(t, tt.expected, job.IsCompleted())
			})
		}
	})

	t.Run("IsExpired method", func(t *testing.T) {
		now := time.Now()

		t.Run("expired status", func(t *testing.T) {
			job := &perplexity.AsyncJobResponse{
				Status:    perplexity.StatusExpired,
				ExpiresAt: now.Add(time.Hour), // Future expiry but status is expired
			}
			assert.True(t, job.IsExpired())
		})

		t.Run("past expiry time", func(t *testing.T) {
			job := &perplexity.AsyncJobResponse{
				Status:    perplexity.StatusProcessing,
				ExpiresAt: now.Add(-time.Hour), // Past expiry
			}
			assert.True(t, job.IsExpired())
		})

		t.Run("not expired", func(t *testing.T) {
			job := &perplexity.AsyncJobResponse{
				Status:    perplexity.StatusProcessing,
				ExpiresAt: now.Add(time.Hour), // Future expiry
			}
			assert.False(t, job.IsExpired())
		})
	})

	t.Run("GetResult method", func(t *testing.T) {
		t.Run("completed with result", func(t *testing.T) {
			expectedResult := &perplexity.CompletionResponse{ID: "test"}
			job := &perplexity.AsyncJobResponse{
				Status: perplexity.StatusCompleted,
				Result: expectedResult,
			}
			result, err := job.GetResult()
			assert.NoError(t, err)
			assert.Equal(t, expectedResult, result)
		})

		t.Run("completed without result", func(t *testing.T) {
			job := &perplexity.AsyncJobResponse{
				Status: perplexity.StatusCompleted,
				Result: nil,
			}
			_, err := job.GetResult()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "no result available")
		})

		t.Run("failed status", func(t *testing.T) {
			errorMsg := "Job failed due to timeout"
			job := &perplexity.AsyncJobResponse{
				ID:     "job_123",
				Status: perplexity.StatusFailed,
				Error:  &errorMsg,
			}
			_, err := job.GetResult()
			assert.Error(t, err)

			var asyncErr *perplexity.AsyncJobError
			assert.ErrorAs(t, err, &asyncErr)
			assert.Equal(t, "job_123", asyncErr.JobID)
			assert.Equal(t, perplexity.StatusFailed, asyncErr.Status)
		})

		t.Run("still processing", func(t *testing.T) {
			job := &perplexity.AsyncJobResponse{
				Status: perplexity.StatusProcessing,
			}
			_, err := job.GetResult()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "still processing")
		})
	})
}

func TestAsyncJobRequest_JSON_Marshaling(t *testing.T) {
	t.Run("marshals regular messages", func(t *testing.T) {
		req := perplexity.NewAsyncJobRequest(
			perplexity.WithAsyncMessages([]perplexity.Message{
				{Role: "user", Content: "Test message"},
			}),
		)

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		// The async API wraps the request in a "request" field
		request, ok := parsed["request"].(map[string]interface{})
		require.True(t, ok)

		messages, ok := request["messages"].([]interface{})
		require.True(t, ok)
		assert.Len(t, messages, 1)
	})

	t.Run("marshals multimodal messages", func(t *testing.T) {
		req := perplexity.NewAsyncJobRequest(
			perplexity.WithAsyncMultimodalMessages([]perplexity.MultimodalMessage{
				{
					Role: "user",
					Content: []perplexity.Content{
						perplexity.NewTextContent("Test message"),
						perplexity.NewImageURLContent("https://example.com/image.jpg"),
					},
				},
			}),
		)

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		// The async API wraps the request in a "request" field
		request, ok := parsed["request"].(map[string]interface{})
		require.True(t, ok)

		messages, ok := request["messages"].([]interface{})
		require.True(t, ok)
		assert.Len(t, messages, 1)

		firstMsg := messages[0].(map[string]interface{})
		content := firstMsg["content"].([]interface{})
		assert.Len(t, content, 2)
	})

	t.Run("HasImages method", func(t *testing.T) {
		t.Run("with images", func(t *testing.T) {
			req := perplexity.NewAsyncJobRequest(
				perplexity.WithAsyncMultimodalMessages([]perplexity.MultimodalMessage{
					{
						Role: "user",
						Content: []perplexity.Content{
							perplexity.NewTextContent("Test"),
							perplexity.NewImageURLContent("https://example.com/image.jpg"),
						},
					},
				}),
			)
			assert.True(t, req.HasImages())
		})

		t.Run("without images", func(t *testing.T) {
			req := perplexity.NewAsyncJobRequest(
				perplexity.WithAsyncMessages([]perplexity.Message{
					{Role: "user", Content: "Test"},
				}),
			)
			assert.False(t, req.HasImages())
		})
	})
}
