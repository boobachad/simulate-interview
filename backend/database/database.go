package database

import (
	"fmt"
	"log"
	"os"

	"github.com/boobachad/simulate-interview/backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect establishes database connection
func Connect() error {
	database_url := os.Getenv("DATABASE_URL")
	if database_url == "" {
		return fmt.Errorf("DATABASE_URL environment variable not set")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(database_url), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established")
	return nil
}

// Migrate runs database migrations
func Migrate() error {
	log.Println("Running database migrations...")

	// Enable UUID extension if not already enabled
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("Warning: Could not create uuid-ossp extension: %v", err)
	}

	// Run auto-migration
	err := DB.AutoMigrate(
		&models.FocusArea{},
		&models.Problem{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Seed initial focus areas
	seedFocusAreas()

	log.Println("Database migrations completed")
	return nil
}

// seedFocusAreas seeds initial focus areas if they don't exist
func seedFocusAreas() {
	focus_areas := []models.FocusArea{
		{
			Name:           "Dynamic Programming",
			Slug:           "dynamic-programming",
			PromptGuidance: "- Focus on optimal substructure and overlapping subproblems\n- Include state transition equations in the solution\n- Examples: Knapsack, LCS, LIS, path finding in grid",
		},
		{
			Name:           "Greedy Algorithms",
			Slug:           "greedy",
			PromptGuidance: "- Focus on making locally optimal choices at each step\n- Prove or explain why the greedy choice works\n- Examples: Interval scheduling, Huffman coding, Dijkstra",
		},
		{
			Name:           "Graph Algorithms",
			Slug:           "graphs",
			PromptGuidance: "- Focus on nodes, edges, and traversals (BFS/DFS)\n- Can involve shortest path, topological sort, or connectivity\n- Examples: Number of islands, course schedule, network delay",
		},
		{
			Name:           "Tree Algorithms",
			Slug:           "trees",
			PromptGuidance: "- Focus on hierarchical data structures, binary trees, BSTs\n- Heavy use of recursion and traversal orders (inorder, preorder, postorder)\n- Examples: Max depth, lowest common ancestor, symmetric tree",
		},
		{
			Name:           "Array and String Manipulation",
			Slug:           "arrays-strings",
			PromptGuidance: "- Focus on efficient indexing, two pointers, or sliding window\n- Usually O(n) or O(n log n) solutions required\n- Examples: Rotated sorted array, anagrams, palindrome",
		},
		{
			Name:           "Sorting and Searching",
			Slug:           "sorting-searching",
			PromptGuidance: "- Focus on custom comparators, merge intervals, or binary search\n- Binary search on answer space is a common pattern\n- Examples: Merge intervals, search in rotated array, Kth largest element",
		},
		{
			Name:           "Backtracking",
			Slug:           "backtracking",
			PromptGuidance: "- Focus on exploring all potential solutions (brute force with pruning)\n- Recursion with state reset\n- Examples: N-Queens, Sudoku solver, combination sum, permutations",
		},
		{
			Name:           "Bit Manipulation",
			Slug:           "bit-manipulation",
			PromptGuidance: "- Focus on XOR, AND, OR, shifting operations\n- O(1) space complexity constraints often apply\n- Examples: Single number, counting bits, reverse bits",
		},
		{
			Name:           "Mathematics and Number Theory",
			Slug:           "mathematics",
			PromptGuidance: "- Focus on prime factors, GCD/LCM, modulo arithmetic\n- efficient calculation avoiding overflow\n- Examples: Count primes, power(x, n), ugly numbers",
		},
		{
			Name:           "Sliding Window",
			Slug:           "sliding-window",
			PromptGuidance: "- Focus on maintaining a window of elements satisfying a condition\n- Two pointers moving in same direction\n- Examples: Longest substring without repeating characters, minimum window substring",
		},
	}

	for _, fa := range focus_areas {
		var existing models.FocusArea
		result := DB.Where("slug = ?", fa.Slug).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			DB.Create(&fa)
			log.Printf("Created focus area: %s", fa.Name)
		} else {
			// Update existing record if prompt guidance is missing or different
			if existing.PromptGuidance != fa.PromptGuidance {
				existing.PromptGuidance = fa.PromptGuidance
				DB.Save(&existing)
				log.Printf("Updated focus area guidance: %s", fa.Name)
			}
		}
	}
}
