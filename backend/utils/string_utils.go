package utils

import (
	"log"
	"strings"
)

// Slugify converts a string to a slug
func Slugify(s string) string {
	// Simple slugify - convert to lowercase and replace spaces with hyphens
	result := ""
	for _, char := range s {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' || char == '_' {
			result += string(char)
		} else if char >= 'A' && char <= 'Z' {
			result += string(char + 32) // Convert to lowercase
		} else if char == ' ' {
			result += "-"
		}
	}
	return result
}

// ExtractJSON extracts a JSON object from a string that might contain other text
// Helpful for extracting JSON from LLM responses that include markdown code blocks
func ExtractJSON(content string) string {
	trimmed := strings.TrimSpace(content)

	// If it looks like raw JSON, don't try to strip markdown blocks (which might be inside the JSON strings)
	if !strings.HasPrefix(trimmed, "{") {
		// Check if content is wrapped in markdown code blocks
		if strings.Contains(content, "```json") {
			parts := strings.Split(content, "```json")
			if len(parts) > 1 {
				content = parts[1]
			}
			// Remove closing backticks if present
			if strings.Contains(content, "```") {
				parts = strings.Split(content, "```")
				content = parts[0]
			}
		} else if strings.Contains(content, "```") {
			// Generic code block
			parts := strings.Split(content, "```")
			if len(parts) > 1 {
				content = parts[1]
			}
		}
	}

	// Basic trim
	content = strings.TrimSpace(content)

	// Attempt to find the first '{' and last '}' to isolate the JSON object
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")

	if start != -1 && end != -1 && end > start {
		content = content[start : end+1]
	} else {
		log.Printf("Warning: Could not find JSON object boundaries in content: %s...", content[:min(len(content), 50)])
	}

	return content
}

// FormatMarkdownDescription attempts to auto-format the problem description
// It ensures that section headers like "Input Format", "Constraints", etc., are properly prefixed with ##
func FormatMarkdownDescription(description string) string {
	lines := strings.Split(description, "\n")
	formattedLines := make([]string, 0, len(lines))

	// Headers we want to ensure have ## prefix and are clean
	headers := []string{
		"Problem Description",
		"Input Format",
		"Output Format",
		"Constraints",
		"Example",
		"Explanation",
		"Sample Input",
		"Sample Output",
		"Solution Hints",
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Strip common markdown formatting characters from start to check content
		cleanLine := strings.TrimLeft(trimmed, "*_# ")

		isHeader := false
		for _, header := range headers {
			// Check if clean line starts with the header (case insensitive)
			if StringStartsWithIgnoreCase(cleanLine, header) {
				// Avoid false positives like "For the Input Format, we use..." by checking
				// if the original line didn't look like a sentence (too long)
				// But allow if it was just bolded like "**Input Format**" (len matches roughly)

				// If it's already a header with #, let it be (or enforce ## if we want)
				if strings.HasPrefix(trimmed, "#") {
					formattedLines = append(formattedLines, line)
					isHeader = true
					break
				}

				// If it looks like a header (short enough, not a sentence)
				if len(cleanLine) < len(header)+3 || strings.HasSuffix(cleanLine, ":") {
					// Use the clean standard header name
					formattedLines = append(formattedLines, "## "+header)
					isHeader = true
					break
				}
			}
		}

		if !isHeader {
			formattedLines = append(formattedLines, line)
		}
	}

	return strings.Join(formattedLines, "\n")
}

// StringStartsWithIgnoreCase checks if s starts with prefix, ignoring case
func StringStartsWithIgnoreCase(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return strings.EqualFold(s[:len(prefix)], prefix)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
