package utils

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

type RateLimiter struct {
	baseDelay  time.Duration
	maxDelay   time.Duration
	maxRetries int
}

// NewRateLimiter creates a new RateLimiter with default values
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		baseDelay:  2 * time.Second,
		maxDelay:   60 * time.Second,
		maxRetries: 5,
	}
}

// ExecuteWithBackoff retries API call with exponential backoff on 429
func (r *RateLimiter) ExecuteWithBackoff(ctx context.Context, fn func() error) error {
	var lastErr error
	delay := r.baseDelay

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(delay):
				// Continue to retry
			case <-ctx.Done():
				return fmt.Errorf("backoff cancelled: %w", ctx.Err())
			}

			// Exponential backoff with cap
			delay *= 2
			if delay > r.maxDelay {
				delay = r.maxDelay
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if it's a rate limit error (429)
		if !isRateLimitError(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		log.Printf("Rate limit hit (attempt %d/%d), retrying after %v", attempt+1, r.maxRetries, delay)
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// isRateLimitError checks if error is a rate limit error
func isRateLimitError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "429") || strings.Contains(errStr, "rate limit")
}
