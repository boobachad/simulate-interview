package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// LeetCodeStats represents user statistics from LeetCode
type LeetCodeStats struct {
	Username     string                `json:"username"`
	Ranking      int                   `json:"ranking"`
	TotalSolved  int                   `json:"total_solved"`
	EasySolved   int                   `json:"easy_solved"`
	MediumSolved int                   `json:"medium_solved"`
	HardSolved   int                   `json:"hard_solved"`
	Skills       map[string]SkillLevel `json:"skills"`
}

// SkillLevel represents skill proficiency with problem count
type SkillLevel struct {
	Level        string `json:"level"` // "Fundamental", "Intermediate", "Advanced"
	ProblemCount int    `json:"problem_count"`
	Tag          string `json:"tag"`
}

// CodeforcesStats represents user statistics from Codeforces
type CodeforcesStats struct {
	Username       string         `json:"username"`
	Rating         int            `json:"rating"`
	MaxRating      int            `json:"max_rating"`
	Rank           string         `json:"rank"`
	MaxRank        string         `json:"max_rank"`
	ProblemsSolved int            `json:"problems_solved"`
	Tags           map[string]int `json:"tags"`
}

// UserProfile combines stats from both platforms
type UserProfile struct {
	Name            string           `json:"name"`
	LeetCodeStats   *LeetCodeStats   `json:"leetcode_stats,omitempty"`
	CodeforcesStats *CodeforcesStats `json:"codeforces_stats,omitempty"`
	SuggestedAreas  []string         `json:"suggested_areas"`
}

// FetchLeetCodeStats fetches user statistics from LeetCode
func FetchLeetCodeStats(username string) (*LeetCodeStats, error) {
	if username == "" {
		return nil, nil
	}

	// LeetCode GraphQL API endpoint
	url := "https://leetcode.com/graphql"

	query := fmt.Sprintf(`{
		matchedUser(username: "%s") {
			username
			profile {
				ranking
			}
			submitStats {
				acSubmissionNum {
					difficulty
					count
				}
			}
			tagProblemCounts {
				advanced {
					tagName
					problemsSolved
				}
				intermediate {
					tagName
					problemsSolved
				}
				fundamental {
					tagName
					problemsSolved
				}
			}
		}
	}`, username)

	requestBody := map[string]string{
		"query": query,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch LeetCode stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LeetCode API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			MatchedUser struct {
				Username string `json:"username"`
				Profile  struct {
					Ranking int `json:"ranking"`
				} `json:"profile"`
				SubmitStats struct {
					AcSubmissionNum []struct {
						Difficulty string `json:"difficulty"`
						Count      int    `json:"count"`
					} `json:"acSubmissionNum"`
				} `json:"submitStats"`
				TagProblemCounts struct {
					Advanced []struct {
						TagName        string `json:"tagName"`
						ProblemsSolved int    `json:"problemsSolved"`
					} `json:"advanced"`
					Intermediate []struct {
						TagName        string `json:"tagName"`
						ProblemsSolved int    `json:"problemsSolved"`
					} `json:"intermediate"`
					Fundamental []struct {
						TagName        string `json:"tagName"`
						ProblemsSolved int    `json:"problemsSolved"`
					} `json:"fundamental"`
				} `json:"tagProblemCounts"`
			} `json:"matchedUser"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode LeetCode response: %w", err)
	}

	stats := &LeetCodeStats{
		Username: result.Data.MatchedUser.Username,
		Ranking:  result.Data.MatchedUser.Profile.Ranking,
		Skills:   make(map[string]SkillLevel),
	}

	// Parse submission stats
	for _, sub := range result.Data.MatchedUser.SubmitStats.AcSubmissionNum {
		switch sub.Difficulty {
		case "All":
			stats.TotalSolved = sub.Count
		case "Easy":
			stats.EasySolved = sub.Count
		case "Medium":
			stats.MediumSolved = sub.Count
		case "Hard":
			stats.HardSolved = sub.Count
		}
	}

	// Parse skills
	for _, tag := range result.Data.MatchedUser.TagProblemCounts.Advanced {
		stats.Skills[tag.TagName] = SkillLevel{
			Level:        "Advanced",
			ProblemCount: tag.ProblemsSolved,
			Tag:          tag.TagName,
		}
	}
	for _, tag := range result.Data.MatchedUser.TagProblemCounts.Intermediate {
		stats.Skills[tag.TagName] = SkillLevel{
			Level:        "Intermediate",
			ProblemCount: tag.ProblemsSolved,
			Tag:          tag.TagName,
		}
	}
	for _, tag := range result.Data.MatchedUser.TagProblemCounts.Fundamental {
		stats.Skills[tag.TagName] = SkillLevel{
			Level:        "Fundamental",
			ProblemCount: tag.ProblemsSolved,
			Tag:          tag.TagName,
		}
	}

	return stats, nil
}

// FetchCodeforcesStats fetches user statistics from Codeforces
func FetchCodeforcesStats(username string) (*CodeforcesStats, error) {
	if username == "" {
		return nil, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Fetch user info
	userInfoURL := fmt.Sprintf("https://codeforces.com/api/user.info?handles=%s", username)
	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Codeforces user info: %w", err)
	}
	defer resp.Body.Close()

	var userInfoResult struct {
		Status string `json:"status"`
		Result []struct {
			Handle    string `json:"handle"`
			Rating    int    `json:"rating"`
			MaxRating int    `json:"maxRating"`
			Rank      string `json:"rank"`
			MaxRank   string `json:"maxRank"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfoResult); err != nil {
		return nil, fmt.Errorf("failed to decode Codeforces user info: %w", err)
	}

	if userInfoResult.Status != "OK" || len(userInfoResult.Result) == 0 {
		return nil, fmt.Errorf("user not found on Codeforces")
	}

	stats := &CodeforcesStats{
		Username:  userInfoResult.Result[0].Handle,
		Rating:    userInfoResult.Result[0].Rating,
		MaxRating: userInfoResult.Result[0].MaxRating,
		Rank:      userInfoResult.Result[0].Rank,
		MaxRank:   userInfoResult.Result[0].MaxRank,
		Tags:      make(map[string]int),
	}

	// Fetch user submissions to count problems by tags
	submissionsURL := fmt.Sprintf("https://codeforces.com/api/user.status?handle=%s", username)
	resp, err = client.Get(submissionsURL)
	if err != nil {
		return stats, nil // Return basic stats even if submissions fail
	}
	defer resp.Body.Close()

	var submissionsResult struct {
		Status string `json:"status"`
		Result []struct {
			Problem struct {
				ContestID int      `json:"contestId"`
				Index     string   `json:"index"`
				Tags      []string `json:"tags"`
			} `json:"problem"`
			Verdict string `json:"verdict"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&submissionsResult); err != nil {
		return stats, nil // Return basic stats
	}

	// Count solved problems by tags
	solvedProblems := make(map[string]bool)
	for _, submission := range submissionsResult.Result {
		if submission.Verdict == "OK" {
			problemKey := fmt.Sprintf("%d-%s", submission.Problem.ContestID, submission.Problem.Index)
			if !solvedProblems[problemKey] {
				solvedProblems[problemKey] = true
				stats.ProblemsSolved++
				for _, tag := range submission.Problem.Tags {
					stats.Tags[tag]++
				}
			}
		}
	}

	return stats, nil
}

// GenerateSuggestedAreas analyzes user stats and suggests focus areas
func GenerateSuggestedAreas(leetcodeStats *LeetCodeStats, codeforcesStats *CodeforcesStats) []string {
	tagCounts := make(map[string]int)

	// Combine tags from both platforms
	if leetcodeStats != nil {
		for tag, skill := range leetcodeStats.Skills {
			normalizedTag := normalizeTag(tag)
			tagCounts[normalizedTag] += skill.ProblemCount
		}
	}

	if codeforcesStats != nil {
		for tag, count := range codeforcesStats.Tags {
			normalizedTag := normalizeTag(tag)
			tagCounts[normalizedTag] += count
		}
	}

	// Map tags to focus areas
	focusAreaMapping := map[string]string{
		"dynamic-programming": "dynamic-programming",
		"dp":                  "dynamic-programming",
		"greedy":              "greedy",
		"graph":               "graphs",
		"graphs":              "graphs",
		"tree":                "trees",
		"trees":               "trees",
		"array":               "arrays-strings",
		"string":              "arrays-strings",
		"sorting":             "sorting-searching",
		"binary-search":       "sorting-searching",
		"backtracking":        "backtracking",
		"bit-manipulation":    "bit-manipulation",
		"math":                "mathematics",
		"number-theory":       "mathematics",
		"sliding-window":      "sliding-window",
	}

	areaScores := make(map[string]int)
	for tag, count := range tagCounts {
		if area, ok := focusAreaMapping[tag]; ok {
			areaScores[area] += count
		}
	}

	// Get areas where user has less practice (to improve weak areas)
	var suggestions []string
	allAreas := []string{"dynamic-programming", "greedy", "graphs", "trees", "arrays-strings",
		"sorting-searching", "backtracking", "bit-manipulation", "mathematics", "sliding-window"}

	for _, area := range allAreas {
		score := areaScores[area]
		// Suggest areas with low to medium practice
		if score < 50 { // Adjust threshold as needed
			suggestions = append(suggestions, area)
			if len(suggestions) >= 3 {
				break
			}
		}
	}

	return suggestions
}

func normalizeTag(tag string) string {
	return strings.ToLower(strings.ReplaceAll(tag, " ", "-"))
}
