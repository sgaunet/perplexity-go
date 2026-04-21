package perplexity

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// Constants for async polling configuration.
const (
	// DefaultInitialPollingInterval is the default initial polling interval.
	DefaultInitialPollingInterval = 2 * time.Second
	// DefaultMaxPollingInterval is the default maximum polling interval.
	DefaultMaxPollingInterval = 30 * time.Second
	// DefaultBackoffMultiplier is the default exponential backoff multiplier.
	DefaultBackoffMultiplier = 1.5
	// DefaultMaxWaitTime is the default maximum wait time for job completion.
	DefaultMaxWaitTime = 30 * time.Minute
	// JitterPercentage is the percentage of jitter to apply to polling intervals.
	JitterPercentage = 0.1
	// JitterOffset centers the jitter around zero.
	JitterOffset = 0.5
)

// AsyncPollingOptions configures the polling behavior for async jobs.
type AsyncPollingOptions struct {
	// InitialInterval is the initial polling interval.
	InitialInterval time.Duration

	// MaxInterval is the maximum polling interval.
	MaxInterval time.Duration

	// BackoffMultiplier is the multiplier for exponential backoff.
	BackoffMultiplier float64

	// MaxWaitTime is the maximum time to wait for job completion.
	MaxWaitTime time.Duration

	// JitterEnabled adds random jitter to polling intervals.
	JitterEnabled bool
}

// DefaultAsyncPollingOptions returns default polling options with exponential backoff.
func DefaultAsyncPollingOptions() *AsyncPollingOptions {
	return &AsyncPollingOptions{
		InitialInterval:   DefaultInitialPollingInterval,
		MaxInterval:       DefaultMaxPollingInterval,
		BackoffMultiplier: DefaultBackoffMultiplier,
		MaxWaitTime:       DefaultMaxWaitTime,
		JitterEnabled:     true,
	}
}

// WaitForAsyncJob polls an async job until completion or timeout.
// Returns the final job response when completed successfully.
func (s *Client) WaitForAsyncJob(jobID string, opts *AsyncPollingOptions) (*AsyncJobResponse, error) {
	return s.WaitForAsyncJobWithContext(context.Background(), jobID, opts)
}

// WaitForAsyncJobWithContext polls an async job with context until completion or timeout.
func (s *Client) WaitForAsyncJobWithContext(ctx context.Context, jobID string, opts *AsyncPollingOptions) (*AsyncJobResponse, error) {
	if err := s.validateJobIDAndOptions(jobID, &opts); err != nil {
		return nil, err
	}

	ctx, cancel := s.setupTimeoutContext(ctx, opts)
	defer cancel()

	return s.pollJobUntilCompletion(ctx, jobID, opts)
}

// pollJobUntilCompletion polls the job until completion.
func (s *Client) pollJobUntilCompletion(ctx context.Context, jobID string, opts *AsyncPollingOptions) (*AsyncJobResponse, error) {
	interval := opts.InitialInterval

	for {
		if err := s.checkContextDone(ctx); err != nil {
			return nil, err
		}

		job, err := s.GetAsyncJobWithContext(ctx, jobID)
		if err != nil {
			return nil, err
		}

		if result, err := s.checkJobCompletion(job); !errors.Is(err, ErrJobNotComplete) {
			return result, err
		}

		time.Sleep(interval)
		interval = s.calculateNextInterval(interval, opts)
	}
}

// WaitForAsyncJobWithProgress polls an async job and sends progress updates via channel.
// The progress channel will receive job responses on each poll until completion.
func (s *Client) WaitForAsyncJobWithProgress(ctx context.Context, jobID string, opts *AsyncPollingOptions, progressChan chan<- *AsyncJobResponse) (*AsyncJobResponse, error) {
	if err := s.validateJobIDAndOptions(jobID, &opts); err != nil {
		return nil, err
	}

	ctx, cancel := s.setupTimeoutContext(ctx, opts)
	defer cancel()
	interval := opts.InitialInterval

	for {
		if err := s.checkContextDone(ctx); err != nil {
			return nil, err
		}

		job, err := s.pollJobStatus(ctx, jobID)
		if err != nil {
			return nil, err
		}

		if err := s.sendProgressUpdate(ctx, progressChan, job); err != nil {
			return nil, err
		}

		if result, err := s.checkJobCompletion(job); !errors.Is(err, ErrJobNotComplete) {
			return result, err
		}

		time.Sleep(interval)
		interval = s.calculateNextInterval(interval, opts)
	}
}

// validateJobIDAndOptions validates the job ID and sets default options if needed.
func (s *Client) validateJobIDAndOptions(jobID string, opts **AsyncPollingOptions) error {
	if jobID == "" {
		return ErrJobIDEmpty
	}
	if *opts == nil {
		*opts = DefaultAsyncPollingOptions()
	}
	return nil
}

// setupTimeoutContext creates a timeout context if MaxWaitTime is set.
// The returned cancel function is always non-nil; the caller is responsible for calling it.
func (s *Client) setupTimeoutContext(ctx context.Context, opts *AsyncPollingOptions) (context.Context, context.CancelFunc) {
	if opts.MaxWaitTime > 0 {
		return context.WithTimeout(ctx, opts.MaxWaitTime) //nolint:gosec // caller defers cancel()
	}
	return context.WithCancel(ctx) //nolint:gosec // caller defers cancel()
}

// checkContextDone checks if the context is done and returns appropriate error.
func (s *Client) checkContextDone(ctx context.Context) error {
	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ErrAsyncPollingTimeout
		}
		return fmt.Errorf("context error: %w", ctx.Err())
	default:
		return nil
	}
}

// pollJobStatus retrieves the current job status.
func (s *Client) pollJobStatus(ctx context.Context, jobID string) (*AsyncJobResponse, error) {
	return s.GetAsyncJobWithContext(ctx, jobID)
}

// sendProgressUpdate sends a progress update to the channel if provided.
func (s *Client) sendProgressUpdate(ctx context.Context, progressChan chan<- *AsyncJobResponse, job *AsyncJobResponse) error {
	if progressChan == nil {
		return nil
	}

	select {
	case progressChan <- job:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context error during progress update: %w", ctx.Err())
	default:
		// Don't block if channel is full
		return nil
	}
}

// checkJobCompletion checks if the job is expired or completed.
func (s *Client) checkJobCompletion(job *AsyncJobResponse) (*AsyncJobResponse, error) {
	if job.IsExpired() {
		return job, ErrAsyncJobExpired
	}
	if job.IsCompleted() {
		return job, nil
	}
	return nil, ErrJobNotComplete
}

// calculateNextInterval calculates the next polling interval using exponential backoff.
func (s *Client) calculateNextInterval(currentInterval time.Duration, opts *AsyncPollingOptions) time.Duration {
	// Apply exponential backoff
	nextInterval := time.Duration(float64(currentInterval) * opts.BackoffMultiplier)

	// Cap at maximum interval
	nextInterval = min(nextInterval, opts.MaxInterval)

	// Add jitter if enabled
	if opts.JitterEnabled {
		jitterRange := float64(nextInterval) * JitterPercentage
		jitter := time.Duration(jitterRange * (rand.Float64() - JitterOffset)) //nolint:gosec // G404: Math rand is acceptable for jitter timing
		nextInterval += jitter
	}

	return nextInterval
}

// IsAsyncJobComplete checks if an async job is in a terminal state.
func IsAsyncJobComplete(status AsyncJobStatus) bool {
	return status == StatusCompleted || status == StatusFailed || status == StatusExpired
}

// EstimateRemainingTime estimates remaining time based on job progress.
// Returns zero if progress information is unavailable.
func EstimateRemainingTime(job *AsyncJobResponse) time.Duration {
	if job.Progress == nil || job.Progress.Percentage == nil {
		return 0
	}

	percentage := *job.Progress.Percentage
	if percentage <= 0 || percentage >= 100 {
		return 0
	}

	// Calculate elapsed time
	elapsed := time.Since(job.CreatedAt)

	// Estimate total time based on current progress
	estimatedTotal := time.Duration(float64(elapsed) * 100.0 / float64(percentage))

	// Calculate remaining time
	remaining := estimatedTotal - elapsed
	if remaining < 0 {
		return 0
	}

	return remaining
}

// AsyncJobMonitor provides advanced monitoring capabilities for async jobs.
type AsyncJobMonitor struct {
	client   *Client
	jobID    string
	options  *AsyncPollingOptions
	started  time.Time
	lastPoll time.Time
}

// NewAsyncJobMonitor creates a new job monitor.
func (s *Client) NewAsyncJobMonitor(jobID string, opts *AsyncPollingOptions) *AsyncJobMonitor {
	if opts == nil {
		opts = DefaultAsyncPollingOptions()
	}

	return &AsyncJobMonitor{
		client:  s,
		jobID:   jobID,
		options: opts,
		started: time.Now(),
	}
}

// GetJobID returns the job ID being monitored.
func (m *AsyncJobMonitor) GetJobID() string {
	return m.jobID
}

// GetElapsedTime returns the time elapsed since monitoring started.
func (m *AsyncJobMonitor) GetElapsedTime() time.Duration {
	return time.Since(m.started)
}

// GetLastPollTime returns the time of the last poll attempt.
func (m *AsyncJobMonitor) GetLastPollTime() time.Time {
	return m.lastPoll
}

// Poll performs a single poll operation and returns the current job status.
func (m *AsyncJobMonitor) Poll(ctx context.Context) (*AsyncJobResponse, error) {
	m.lastPoll = time.Now()
	return m.client.GetAsyncJobWithContext(ctx, m.jobID)
}

// Wait waits for job completion using the configured polling options.
func (m *AsyncJobMonitor) Wait(ctx context.Context) (*AsyncJobResponse, error) {
	return m.client.WaitForAsyncJobWithContext(ctx, m.jobID, m.options)
}

// WaitWithProgress waits for job completion and sends progress updates.
func (m *AsyncJobMonitor) WaitWithProgress(ctx context.Context, progressChan chan<- *AsyncJobResponse) (*AsyncJobResponse, error) {
	return m.client.WaitForAsyncJobWithProgress(ctx, m.jobID, m.options, progressChan)
}
