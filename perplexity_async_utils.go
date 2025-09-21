package perplexity

import (
	"context"
	"errors"
	"time"
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
		InitialInterval:   2 * time.Second,
		MaxInterval:       30 * time.Second,
		BackoffMultiplier: 1.5,
		MaxWaitTime:       30 * time.Minute,
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
	if jobID == "" {
		return nil, errors.New("job ID cannot be empty")
	}

	if opts == nil {
		opts = DefaultAsyncPollingOptions()
	}

	// Create timeout context if MaxWaitTime is set
	if opts.MaxWaitTime > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.MaxWaitTime)
		defer cancel()
	}

	interval := opts.InitialInterval

	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return nil, ErrAsyncPollingTimeout
			}
			return nil, ctx.Err()
		default:
		}

		// Get job status
		job, err := s.GetAsyncJobWithContext(ctx, jobID)
		if err != nil {
			return nil, err
		}

		// Check if job has expired first
		if job.IsExpired() {
			return job, ErrAsyncJobExpired
		}

		// Check if job is completed
		if job.IsCompleted() {
			return job, nil
		}

		// Wait before next poll
		time.Sleep(interval)

		// Calculate next interval with exponential backoff
		interval = s.calculateNextInterval(interval, opts)
	}
}

// WaitForAsyncJobWithProgress polls an async job and sends progress updates via channel.
// The progress channel will receive job responses on each poll until completion.
func (s *Client) WaitForAsyncJobWithProgress(ctx context.Context, jobID string, opts *AsyncPollingOptions, progressChan chan<- *AsyncJobResponse) (*AsyncJobResponse, error) {
	if jobID == "" {
		return nil, errors.New("job ID cannot be empty")
	}

	if opts == nil {
		opts = DefaultAsyncPollingOptions()
	}

	// Create timeout context if MaxWaitTime is set
	if opts.MaxWaitTime > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.MaxWaitTime)
		defer cancel()
	}

	interval := opts.InitialInterval

	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return nil, ErrAsyncPollingTimeout
			}
			return nil, ctx.Err()
		default:
		}

		// Get job status
		job, err := s.GetAsyncJobWithContext(ctx, jobID)
		if err != nil {
			return nil, err
		}

		// Send progress update if channel is provided
		if progressChan != nil {
			select {
			case progressChan <- job:
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				// Don't block if channel is full
			}
		}

		// Check if job has expired first
		if job.IsExpired() {
			return job, ErrAsyncJobExpired
		}

		// Check if job is completed
		if job.IsCompleted() {
			return job, nil
		}

		// Wait before next poll
		time.Sleep(interval)

		// Calculate next interval with exponential backoff
		interval = s.calculateNextInterval(interval, opts)
	}
}

// calculateNextInterval calculates the next polling interval using exponential backoff.
func (s *Client) calculateNextInterval(currentInterval time.Duration, opts *AsyncPollingOptions) time.Duration {
	// Apply exponential backoff
	nextInterval := time.Duration(float64(currentInterval) * opts.BackoffMultiplier)

	// Cap at maximum interval
	if nextInterval > opts.MaxInterval {
		nextInterval = opts.MaxInterval
	}

	// Add jitter if enabled
	if opts.JitterEnabled {
		jitter := time.Duration(float64(nextInterval) * 0.1 * (0.5 - 0.5)) // ±10% jitter
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