package perplexity_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sgaunet/perplexity-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultAsyncPollingOptions(t *testing.T) {
	opts := perplexity.DefaultAsyncPollingOptions()
	assert.Equal(t, 2*time.Second, opts.InitialInterval)
	assert.Equal(t, 30*time.Second, opts.MaxInterval)
	assert.Equal(t, 1.5, opts.BackoffMultiplier)
	assert.Equal(t, 30*time.Minute, opts.MaxWaitTime)
	assert.True(t, opts.JitterEnabled)
}

func TestWaitForAsyncJob(t *testing.T) {
	t.Run("job completes successfully", func(t *testing.T) {
		callCount := 0
		mockResponses := []*perplexity.AsyncJobResponse{
			{
				ID:        "job_123",
				Status:    perplexity.StatusPending,
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			},
			{
				ID:        "job_123",
				Status:    perplexity.StatusProcessing,
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				Progress: &perplexity.AsyncJobProgress{
					Stage:      "analyzing",
					Percentage: intPtr(30),
				},
			},
			{
				ID:        "job_123",
				Status:    perplexity.StatusCompleted,
				CreatedAt: time.Now(),
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
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/async/chat/completions/job_123", r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponses[callCount])
			callCount++
		}))
		defer server.Close()

		client := perplexity.NewClient("test-key")
		client.SetAsyncEndpoint(server.URL + "/async/chat/completions")

		opts := &perplexity.AsyncPollingOptions{
			InitialInterval:   50 * time.Millisecond, // Fast polling for tests
			MaxInterval:       100 * time.Millisecond,
			BackoffMultiplier: 1.2,
			MaxWaitTime:       5 * time.Second,
			JitterEnabled:     false, // Disable jitter for predictable tests
		}

		result, err := client.WaitForAsyncJob("job_123", opts)
		require.NoError(t, err)
		assert.Equal(t, "job_123", result.ID)
		assert.Equal(t, perplexity.StatusCompleted, result.Status)
		assert.NotNil(t, result.Result)
		assert.Equal(t, 3, callCount) // Should have made 3 calls
	})

	t.Run("job fails", func(t *testing.T) {
		errorMsg := "Job failed due to internal error"
		mockResponse := &perplexity.AsyncJobResponse{
			ID:        "job_123",
			Status:    perplexity.StatusFailed,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			Error:     &errorMsg,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client := perplexity.NewClient("test-key")
		client.SetAsyncEndpoint(server.URL + "/async/chat/completions")

		opts := &perplexity.AsyncPollingOptions{
			InitialInterval: 10 * time.Millisecond,
			MaxWaitTime:     1 * time.Second,
		}

		result, err := client.WaitForAsyncJob("job_123", opts)
		require.NoError(t, err) // Should return the failed job, not error
		assert.Equal(t, perplexity.StatusFailed, result.Status)
		assert.NotNil(t, result.Error)
	})

	t.Run("job expires", func(t *testing.T) {
		mockResponse := &perplexity.AsyncJobResponse{
			ID:        "job_123",
			Status:    perplexity.StatusExpired,
			CreatedAt: time.Now().Add(-8 * 24 * time.Hour),
			ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired yesterday
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client := perplexity.NewClient("test-key")
		client.SetAsyncEndpoint(server.URL + "/async/chat/completions")

		result, err := client.WaitForAsyncJob("job_123", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "async job has expired")
		assert.NotNil(t, result)
		assert.Equal(t, perplexity.StatusExpired, result.Status)
	})

	t.Run("timeout", func(t *testing.T) {
		mockResponse := &perplexity.AsyncJobResponse{
			ID:        "job_123",
			Status:    perplexity.StatusProcessing,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client := perplexity.NewClient("test-key")
		client.SetAsyncEndpoint(server.URL + "/async/chat/completions")

		opts := &perplexity.AsyncPollingOptions{
			InitialInterval: 10 * time.Millisecond,
			MaxWaitTime:     50 * time.Millisecond, // Very short timeout
		}

		_, err := client.WaitForAsyncJob("job_123", opts)
		assert.Error(t, err)
		// The error could be either polling timeout or context deadline exceeded
		assert.True(t, strings.Contains(err.Error(), "async job polling timeout") ||
			strings.Contains(err.Error(), "context deadline exceeded"))
	})

	t.Run("context cancellation", func(t *testing.T) {
		mockResponse := &perplexity.AsyncJobResponse{
			ID:        "job_123",
			Status:    perplexity.StatusProcessing,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client := perplexity.NewClient("test-key")
		client.SetAsyncEndpoint(server.URL + "/async/chat/completions")

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		_, err := client.WaitForAsyncJobWithContext(ctx, "job_123", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("empty job ID", func(t *testing.T) {
		client := perplexity.NewClient("test-key")
		_, err := client.WaitForAsyncJob("", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job ID cannot be empty")
	})
}

func TestWaitForAsyncJobWithProgress(t *testing.T) {
	t.Run("progress updates", func(t *testing.T) {
		callCount := 0
		mockResponses := []*perplexity.AsyncJobResponse{
			{
				ID:        "job_123",
				Status:    perplexity.StatusPending,
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			},
			{
				ID:        "job_123",
				Status:    perplexity.StatusProcessing,
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				Progress: &perplexity.AsyncJobProgress{
					Stage:      "analyzing",
					Percentage: intPtr(50),
				},
			},
			{
				ID:        "job_123",
				Status:    perplexity.StatusCompleted,
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				Result: &perplexity.CompletionResponse{
					ID: "completion_456",
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponses[callCount])
			callCount++
		}))
		defer server.Close()

		client := perplexity.NewClient("test-key")
		client.SetAsyncEndpoint(server.URL + "/async/chat/completions")

		progressChan := make(chan *perplexity.AsyncJobResponse, 10)
		opts := &perplexity.AsyncPollingOptions{
			InitialInterval: 10 * time.Millisecond,
			MaxWaitTime:     1 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		go func() {
			result, err := client.WaitForAsyncJobWithProgress(ctx, "job_123", opts, progressChan)
			require.NoError(t, err)
			assert.Equal(t, perplexity.StatusCompleted, result.Status)
		}()

		// Collect progress updates
		var progressUpdates []*perplexity.AsyncJobResponse
		for update := range progressChan {
			progressUpdates = append(progressUpdates, update)
			if update.Status == perplexity.StatusCompleted {
				break
			}
		}

		assert.Len(t, progressUpdates, 3)
		assert.Equal(t, perplexity.StatusPending, progressUpdates[0].Status)
		assert.Equal(t, perplexity.StatusProcessing, progressUpdates[1].Status)
		assert.Equal(t, perplexity.StatusCompleted, progressUpdates[2].Status)
	})
}

func TestAsyncJobMonitor(t *testing.T) {
	t.Run("monitor creation and basic operations", func(t *testing.T) {
		client := perplexity.NewClient("test-key")
		monitor := client.NewAsyncJobMonitor("job_123", nil)

		assert.Equal(t, "job_123", monitor.GetJobID())
		assert.True(t, monitor.GetElapsedTime() >= 0)
		assert.True(t, monitor.GetLastPollTime().IsZero()) // Not polled yet
	})

	t.Run("poll operation", func(t *testing.T) {
		mockResponse := &perplexity.AsyncJobResponse{
			ID:        "job_123",
			Status:    perplexity.StatusProcessing,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client := perplexity.NewClient("test-key")
		client.SetAsyncEndpoint(server.URL + "/async/chat/completions")
		monitor := client.NewAsyncJobMonitor("job_123", nil)

		result, err := monitor.Poll(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "job_123", result.ID)
		assert.Equal(t, perplexity.StatusProcessing, result.Status)
		assert.False(t, monitor.GetLastPollTime().IsZero()) // Should be set after poll
	})
}

func TestEstimateRemainingTime(t *testing.T) {
	t.Run("with progress information", func(t *testing.T) {
		createdAt := time.Now().Add(-10 * time.Minute) // Started 10 minutes ago
		job := &perplexity.AsyncJobResponse{
			CreatedAt: createdAt,
			Progress: &perplexity.AsyncJobProgress{
				Percentage: intPtr(25), // 25% complete
			},
		}

		remaining := perplexity.EstimateRemainingTime(job)

		// Should estimate about 30 minutes remaining (10 min elapsed / 25% = 40 min total - 10 min elapsed)
		// Allow some tolerance for timing variations in tests
		assert.True(t, remaining > 25*time.Minute && remaining < 35*time.Minute,
			"Expected remaining time around 30 minutes, got %v", remaining)
	})

	t.Run("without progress information", func(t *testing.T) {
		job := &perplexity.AsyncJobResponse{
			CreatedAt: time.Now().Add(-10 * time.Minute),
			Progress:  nil,
		}

		remaining := perplexity.EstimateRemainingTime(job)
		assert.Equal(t, time.Duration(0), remaining)
	})

	t.Run("with 0% progress", func(t *testing.T) {
		job := &perplexity.AsyncJobResponse{
			CreatedAt: time.Now().Add(-5 * time.Minute),
			Progress: &perplexity.AsyncJobProgress{
				Percentage: intPtr(0),
			},
		}

		remaining := perplexity.EstimateRemainingTime(job)
		assert.Equal(t, time.Duration(0), remaining)
	})

	t.Run("with 100% progress", func(t *testing.T) {
		job := &perplexity.AsyncJobResponse{
			CreatedAt: time.Now().Add(-5 * time.Minute),
			Progress: &perplexity.AsyncJobProgress{
				Percentage: intPtr(100),
			},
		}

		remaining := perplexity.EstimateRemainingTime(job)
		assert.Equal(t, time.Duration(0), remaining)
	})
}

func TestIsAsyncJobComplete(t *testing.T) {
	tests := []struct {
		status   perplexity.AsyncJobStatus
		expected bool
	}{
		{perplexity.StatusPending, false},
		{perplexity.StatusProcessing, false},
		{perplexity.StatusCompleted, true},
		{perplexity.StatusFailed, true},
		{perplexity.StatusExpired, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := perplexity.IsAsyncJobComplete(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}
