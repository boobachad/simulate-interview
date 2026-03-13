package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
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

	g, gCtx := errgroup.WithContext(ctx)

	var leetcodeStats *LeetCodeStats
	var codeforcesStats *CodeforcesStats
	var leetcodeErr, codeforcesErr error

	// Fetch LeetCode stats in parallel
	g.Go(func() error {
		if leetcodeUsername == "" {
			return nil
		}
		stats, err := FetchLeetCodeStats(gCtx, leetcodeUsername)
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
		stats, err := FetchCodeforcesStats(gCtx, codeforcesUsername)
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

	log.Printf("SyncStats: Both fetches completed, checking context before storeStats")
	select {
	case <-ctx.Done():
		log.Printf("SyncStats: Context was canceled before storeStats: %v", ctx.Err())
		return fmt.Errorf("context canceled before storing stats: %w", ctx.Err())
	default:
		log.Printf("SyncStats: Context still valid, proceeding to storeStats")
	}

	// Check if both platforms failed
	if leetcodeErr != nil && codeforcesErr != nil {
		return fmt.Errorf("both platforms failed: leetcode=%w, codeforces=%v", leetcodeErr, codeforcesErr)
	}

	if err := s.storeStats(ctx, userID, leetcodeStats, codeforcesStats); err != nil {
		return fmt.Errorf("store stats: %w", err)
	}

	if err := s.updateFocusAreas(ctx, leetcodeStats, codeforcesStats); err != nil {
		log.Printf("Failed to update focus areas: %v", err)
	}

	return nil
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

// SkillStats represents skill/tag statistics with problem count
type SkillStats struct {
	Name         string
	ProblemCount int
	Level        string
}

// BuildPersonalizationContext creates context string for AI prompts
func (s *statsService) BuildPersonalizationContext(ctx context.Context, userID uuid.UUID, focusMode string, focusTopic string, focusTopics []string) (string, error) {
	stats, err := s.GetStats(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get stats: %w", err)
	}

	var contextParts []string

	// Global stats - always include
	if stats.LeetCode != nil {
		contextParts = append(contextParts, fmt.Sprintf(
			"LeetCode Profile: %d total solved (Easy: %d, Medium: %d, Hard: %d)",
			stats.LeetCode.TotalSolved,
			stats.LeetCode.EasySolved,
			stats.LeetCode.MediumSolved,
			stats.LeetCode.HardSolved,
		))
	}

	if stats.Codeforces != nil {
		contextParts = append(contextParts, fmt.Sprintf(
			"Codeforces Profile: Rating %d (Max: %d), %d problems solved, %d contests participated",
			stats.Codeforces.Rating,
			stats.Codeforces.MaxRating,
			stats.Codeforces.ProblemsSolved,
			stats.Codeforces.ContestCount,
		))
	}

	// Mode-specific context
	if focusMode == "single" && focusTopic != "" {
		contextParts = append(contextParts, fmt.Sprintf("\n--- FOCUS MODE: Single Topic ---"))
		contextParts = append(contextParts, fmt.Sprintf("Selected Topic: %s", focusTopic))
		
		// Add per-topic stats for the selected topic
		topicStats := s.getTopicStats(stats, focusTopic)
		if topicStats != "" {
			contextParts = append(contextParts, topicStats)
		}
		
		contextParts = append(contextParts, "\nINSTRUCTION: Generate a problem STRICTLY focused on this topic. Tailor difficulty based on user's solve count for this topic.")
		
	} else if focusMode == "multiple" && len(focusTopics) > 0 {
		contextParts = append(contextParts, fmt.Sprintf("\n--- FOCUS MODE: Multiple Topics ---"))
		contextParts = append(contextParts, fmt.Sprintf("Selected Topics: %s", strings.Join(focusTopics, ", ")))
		
		// Add per-topic stats for each selected topic
		contextParts = append(contextParts, "\nPer-Topic Statistics:")
		for _, topic := range focusTopics {
			topicStats := s.getTopicStats(stats, topic)
			if topicStats != "" {
				contextParts = append(contextParts, topicStats)
			}
		}
		
		contextParts = append(contextParts, "\nINSTRUCTION: Generate a problem that combines these topics. Apply the configured strategy (rotate/combine/mix). Tailor difficulty based on user's weakest topic.")
		
	} else {
		// All/Random mode - identify weak areas
		contextParts = append(contextParts, fmt.Sprintf("\n--- FOCUS MODE: All/Random (Weakness-Based) ---"))
		
		weakAreas := s.identifyWeakAreas(stats)
		if len(weakAreas) > 0 {
			contextParts = append(contextParts, "User's WEAK AREAS (topics with fewest problems solved):")
			for i, area := range weakAreas {
				if i >= 10 {
					break
				}
				contextParts = append(contextParts, fmt.Sprintf("  %d. %s: %d problems solved", i+1, area.Name, area.ProblemCount))
			}
			contextParts = append(contextParts, "\nINSTRUCTION: Generate a problem targeting one of these weak areas. Pick topics where user has LEAST experience.")
		} else {
			contextParts = append(contextParts, "INSTRUCTION: Generate a problem based on user's overall skill level.")
		}
	}

	contextStr := ""
	for _, part := range contextParts {
		contextStr += part + "\n"
	}

	return contextStr, nil
}

// getTopicStats returns statistics for a specific topic from user's profile
func (s *statsService) getTopicStats(stats *CombinedStats, topic string) string {
	var parts []string
	
	topicLower := strings.ToLower(topic)
	
	// Check LeetCode skills
	if stats.LeetCode != nil && len(stats.LeetCode.Skills) > 0 {
		for tag, skill := range stats.LeetCode.Skills {
			if strings.ToLower(tag) == topicLower {
				parts = append(parts, fmt.Sprintf("  - LeetCode '%s': %d problems solved (%s level)", tag, skill.ProblemCount, skill.Level))
			}
		}
	}
	
	// Check Codeforces tags
	if stats.Codeforces != nil && len(stats.Codeforces.Tags) > 0 {
		for tag, count := range stats.Codeforces.Tags {
			if strings.ToLower(tag) == topicLower {
				parts = append(parts, fmt.Sprintf("  - Codeforces '%s': %d problems solved", tag, count))
			}
		}
	}
	
	if len(parts) == 0 {
		return fmt.Sprintf("  - Topic '%s': No problems solved yet (NEW TOPIC for user)", topic)
	}
	
	return strings.Join(parts, "\n")
}

// identifyWeakAreas returns topics sorted by problem count (ascending) - weak areas first
func (s *statsService) identifyWeakAreas(stats *CombinedStats) []SkillStats {
	skillMap := make(map[string]*SkillStats)
	
	// Aggregate from LeetCode
	if stats.LeetCode != nil && len(stats.LeetCode.Skills) > 0 {
		for tag, skill := range stats.LeetCode.Skills {
			tagNormalized := strings.ToLower(tag)
			if existing, ok := skillMap[tagNormalized]; ok {
				existing.ProblemCount += skill.ProblemCount
			} else {
				skillMap[tagNormalized] = &SkillStats{
					Name:         tag,
					ProblemCount: skill.ProblemCount,
					Level:        skill.Level,
				}
			}
		}
	}
	
	// Aggregate from Codeforces
	if stats.Codeforces != nil && len(stats.Codeforces.Tags) > 0 {
		for tag, count := range stats.Codeforces.Tags {
			tagNormalized := strings.ToLower(tag)
			if existing, ok := skillMap[tagNormalized]; ok {
				existing.ProblemCount += count
			} else {
				skillMap[tagNormalized] = &SkillStats{
					Name:         tag,
					ProblemCount: count,
					Level:        "unknown",
				}
			}
		}
	}
	
	// Convert map to slice
	var skills []SkillStats
	for _, skill := range skillMap {
		skills = append(skills, *skill)
	}
	
	// Sort by problem count ascending (weak areas first)
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].ProblemCount < skills[j].ProblemCount
	})
	
	return skills
}

// updateFocusAreas populates focus_area_dynamics from fetched stats
func (s *statsService) updateFocusAreas(ctx context.Context, leetcodeStats *LeetCodeStats, codeforcesStats *CodeforcesStats) error {
	type FocusAreaUpdate struct {
		Platform     string
		Topic        string
		ProblemCount int
	}

	var updates []FocusAreaUpdate

	if leetcodeStats != nil {
		for topic, skill := range leetcodeStats.Skills {
			if skill.ProblemCount > 0 {
				updates = append(updates, FocusAreaUpdate{
					Platform:     "leetcode",
					Topic:        topic,
					ProblemCount: skill.ProblemCount,
				})
			}
		}
	}

	if codeforcesStats != nil {
		for tag, count := range codeforcesStats.Tags {
			if count > 0 {
				updates = append(updates, FocusAreaUpdate{
					Platform:     "codeforces",
					Topic:        tag,
					ProblemCount: count,
				})
			}
		}
	}

	for _, update := range updates {
		if err := s.db.WithContext(ctx).Exec(`
			INSERT INTO focus_area_dynamics (platform, topic, problem_count, updated_at)
			VALUES (?, ?, ?, ?)
			ON CONFLICT (platform, topic)
			DO UPDATE SET problem_count = EXCLUDED.problem_count, updated_at = EXCLUDED.updated_at
		`, update.Platform, update.Topic, update.ProblemCount, time.Now()).Error; err != nil {
			return fmt.Errorf("upsert focus area %s:%s: %w", update.Platform, update.Topic, err)
		}
	}

	log.Printf("Updated %d focus areas", len(updates))
	return nil
}
