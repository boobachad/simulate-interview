package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type statsService struct {
	db     *gorm.DB
	client *http.Client
}

// NewStatsService creates a new StatsService instance
func NewStatsService(db *gorm.DB, client *http.Client) *statsService {
	return &statsService{db: db, client: client}
}

// CombinedStats represents stats from both platforms
type CombinedStats struct {
	LeetCode   *LeetCodeStats   `json:"leetcode"`
	Codeforces *CodeforcesStats `json:"codeforces"`
	CachedAt   time.Time        `json:"cached_at"`
}

// SyncStats fetches stats from both platforms in parallel
func (s *statsService) SyncStats(ctx context.Context, userID uuid.UUID) error {
	var profiles []models.CodingProfile
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&profiles).Error; err != nil {
		return fmt.Errorf("fetch coding profiles: %w", err)
	}

	var leetcodeUsername, codeforcesUsername string
	for _, p := range profiles {
		if p.Platform == "leetcode" {
			leetcodeUsername = p.Username
		} else if p.Platform == "codeforces" {
			codeforcesUsername = p.Username
		}
	}

	g, ctx := errgroup.WithContext(ctx)

	var leetcodeStats *LeetCodeStats
	var codeforcesStats *CodeforcesStats
	var leetcodeErr, codeforcesErr error

	// Fetch LeetCode stats in parallel
	g.Go(func() error {
		if leetcodeUsername == "" {
			return nil
		}
		stats, err := FetchLeetCodeStats(leetcodeUsername)
		if err != nil {
			leetcodeErr = err
			log.Printf("LeetCode fetch failed for user %s: %v", userID, err)
			return nil
		}
		leetcodeStats = stats
		return nil
	})

	// Fetch Codeforces stats in parallel
	g.Go(func() error {
		if codeforcesUsername == "" {
			return nil
		}
		stats, err := FetchCodeforcesStats(codeforcesUsername)
		if err != nil {
			codeforcesErr = err
			log.Printf("Codeforces fetch failed for user %s: %v", userID, err)
			return nil
		}
		codeforcesStats = stats
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("stats sync failed: %w", err)
	}

	// Check if both platforms failed
	if leetcodeErr != nil && codeforcesErr != nil {
		return fmt.Errorf("both platforms failed: leetcode=%w, codeforces=%v", leetcodeErr, codeforcesErr)
	}

	return s.storeStats(ctx, userID, leetcodeStats, codeforcesStats)
}

// storeStats saves stats to database
func (s *statsService) storeStats(ctx context.Context, userID uuid.UUID, leetcodeStats *LeetCodeStats, codeforcesStats *CodeforcesStats) error {
	now := time.Now()

	if leetcodeStats != nil {
		statsJSON, err := json.Marshal(leetcodeStats)
		if err != nil {
			return fmt.Errorf("marshal leetcode stats: %w", err)
		}

		userStats := models.UserStats{
			UserID:    userID,
			Platform:  "leetcode",
			StatsJSON: models.StatsJSON(statsJSON),
			SyncedAt:  now,
		}

		if err := s.db.WithContext(ctx).
			Where("user_id = ? AND platform = ?", userID, "leetcode").
			Assign(userStats).
			FirstOrCreate(&userStats).Error; err != nil {
			return fmt.Errorf("store leetcode stats: %w", err)
		}

		// Update last_synced in coding_profiles
		if err := s.db.WithContext(ctx).
			Model(&models.CodingProfile{}).
			Where("user_id = ? AND platform = ?", userID, "leetcode").
			Update("last_synced", now).Error; err != nil {
			return fmt.Errorf("update leetcode last_synced: %w", err)
		}
	}

	if codeforcesStats != nil {
		statsJSON, err := json.Marshal(codeforcesStats)
		if err != nil {
			return fmt.Errorf("marshal codeforces stats: %w", err)
		}

		userStats := models.UserStats{
			UserID:    userID,
			Platform:  "codeforces",
			StatsJSON: models.StatsJSON(statsJSON),
			SyncedAt:  now,
		}

		if err := s.db.WithContext(ctx).
			Where("user_id = ? AND platform = ?", userID, "codeforces").
			Assign(userStats).
			FirstOrCreate(&userStats).Error; err != nil {
			return fmt.Errorf("store codeforces stats: %w", err)
		}

		// Update last_synced in coding_profiles
		if err := s.db.WithContext(ctx).
			Model(&models.CodingProfile{}).
			Where("user_id = ? AND platform = ?", userID, "codeforces").
			Update("last_synced", now).Error; err != nil {
			return fmt.Errorf("update codeforces last_synced: %w", err)
		}
	}

	return nil
}

// GetStats retrieves cached stats with 1-hour TTL check
func (s *statsService) GetStats(ctx context.Context, userID uuid.UUID) (*CombinedStats, error) {
	var userStats []models.UserStats
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&userStats).Error; err != nil {
		return nil, fmt.Errorf("query user stats: %w", err)
	}

	combined := &CombinedStats{}
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	for _, stats := range userStats {
		// Check if cache is stale
		if stats.SyncedAt.Before(oneHourAgo) {
			log.Printf("Stats cache stale for user %s platform %s", userID, stats.Platform)
		}

		if stats.Platform == "leetcode" {
			var leetcodeStats LeetCodeStats
			if err := json.Unmarshal(stats.StatsJSON, &leetcodeStats); err != nil {
				log.Printf("Failed to unmarshal leetcode stats: %v", err)
				continue
			}
			combined.LeetCode = &leetcodeStats
			combined.CachedAt = stats.SyncedAt
		} else if stats.Platform == "codeforces" {
			var codeforcesStats CodeforcesStats
			if err := json.Unmarshal(stats.StatsJSON, &codeforcesStats); err != nil {
				log.Printf("Failed to unmarshal codeforces stats: %v", err)
				continue
			}
			combined.Codeforces = &codeforcesStats
			if combined.CachedAt.IsZero() || stats.SyncedAt.Before(combined.CachedAt) {
				combined.CachedAt = stats.SyncedAt
			}
		}
	}

	return combined, nil
}

// BuildPersonalizationContext creates context string for AI prompts
func (s *statsService) BuildPersonalizationContext(ctx context.Context, userID uuid.UUID, focusMode string, focusTopic string) (string, error) {
	stats, err := s.GetStats(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get stats: %w", err)
	}

	var contextParts []string

	if stats.LeetCode != nil {
		contextParts = append(contextParts, fmt.Sprintf(
			"LeetCode: %d total solved (Easy: %d, Medium: %d, Hard: %d)",
			stats.LeetCode.TotalSolved,
			stats.LeetCode.EasySolved,
			stats.LeetCode.MediumSolved,
			stats.LeetCode.HardSolved,
		))

		// Add top skills
		if len(stats.LeetCode.Skills) > 0 {
			contextParts = append(contextParts, "Top LeetCode skills:")
			count := 0
			for tag, skill := range stats.LeetCode.Skills {
				if count >= 5 {
					break
				}
				contextParts = append(contextParts, fmt.Sprintf("  - %s: %d problems (%s)", tag, skill.ProblemCount, skill.Level))
				count++
			}
		}
	}

	if stats.Codeforces != nil {
		contextParts = append(contextParts, fmt.Sprintf(
			"Codeforces: Rating %d (Max: %d), %d problems solved, %d contests",
			stats.Codeforces.Rating,
			stats.Codeforces.MaxRating,
			stats.Codeforces.ProblemsSolved,
			stats.Codeforces.ContestCount,
		))

		// Add top tags
		if len(stats.Codeforces.Tags) > 0 {
			contextParts = append(contextParts, "Top Codeforces tags:")
			count := 0
			for tag, solvedCount := range stats.Codeforces.Tags {
				if count >= 5 {
					break
				}
				contextParts = append(contextParts, fmt.Sprintf("  - %s: %d problems", tag, solvedCount))
				count++
			}
		}
	}

	// Add focus mode context
	if focusMode == "single" && focusTopic != "" {
		contextParts = append(contextParts, fmt.Sprintf("\nFocus: Generate problem specifically for topic '%s'", focusTopic))
	} else {
		contextParts = append(contextParts, "\nFocus: Generate problem targeting user's weak areas")
	}

	contextStr := ""
	for _, part := range contextParts {
		contextStr += part + "\n"
	}

	return contextStr, nil
}
