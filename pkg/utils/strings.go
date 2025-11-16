package utils

import (
	"crypto/rand"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// IsEmpty checks if a string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty checks if a string is not empty and contains non-whitespace characters
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// DefaultIfEmpty returns the default value if the string is empty
func DefaultIfEmpty(s, defaultValue string) string {
	if IsEmpty(s) {
		return defaultValue
	}
	return s
}

// Truncate truncates a string to the specified length with optional ellipsis
func Truncate(s string, maxLength int, ellipsis ...string) string {
	if len(s) <= maxLength {
		return s
	}

	suffix := "..."
	if len(ellipsis) > 0 {
		suffix = ellipsis[0]
	}

	return s[:maxLength-len(suffix)] + suffix
}

// TruncateWords truncates a string to the specified number of words
func TruncateWords(s string, maxWords int, ellipsis ...string) string {
	words := strings.Fields(s)
	if len(words) <= maxWords {
		return s
	}

	suffix := "..."
	if len(ellipsis) > 0 {
		suffix = ellipsis[0]
	}

	return strings.Join(words[:maxWords], " ") + suffix
}

// Capitalize capitalizes the first letter of a string
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// Title converts a string to title case (first letter of each word capitalized)
func Title(s string) string {
	return strings.Title(strings.ToLower(s))
}

// CamelCase converts a string to camelCase
func CamelCase(s string) string {
	words := strings.Fields(regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(s, " "))
	if len(words) == 0 {
		return ""
	}

	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		result += Capitalize(strings.ToLower(words[i]))
	}
	return result
}

// PascalCase converts a string to PascalCase
func PascalCase(s string) string {
	words := strings.Fields(regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(s, " "))
	var result strings.Builder

	for _, word := range words {
		result.WriteString(Capitalize(strings.ToLower(word)))
	}
	return result.String()
}

// SnakeCase converts a string to snake_case
func SnakeCase(s string) string {
	// Insert underscores before uppercase letters (for camelCase/PascalCase)
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	s = re.ReplaceAllString(s, `${1}_${2}`)

	// Replace non-alphanumeric characters with underscores
	re = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	s = re.ReplaceAllString(s, "_")

	// Remove leading and trailing underscores
	s = strings.Trim(s, "_")

	return strings.ToLower(s)
}

// KebabCase converts a string to kebab-case
func KebabCase(s string) string {
	return strings.ReplaceAll(SnakeCase(s), "_", "-")
}

// Slugify creates a URL-friendly slug from a string
func Slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and special characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")

	// Remove leading and trailing hyphens
	s = strings.Trim(s, "-")

	return s
}

// RandomString generates a random string of specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	for i := range b {
		randomByte := make([]byte, 1)
		rand.Read(randomByte)
		b[i] = charset[randomByte[0]%byte(len(charset))]
	}

	return string(b)
}

// RandomAlphanumeric generates a random alphanumeric string
func RandomAlphanumeric(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return randomStringWithCharset(length, charset)
}

// RandomAlphabetic generates a random alphabetic string
func RandomAlphabetic(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return randomStringWithCharset(length, charset)
}

// RandomNumeric generates a random numeric string
func RandomNumeric(length int) string {
	const charset = "0123456789"
	return randomStringWithCharset(length, charset)
}

func randomStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		randomByte := make([]byte, 1)
		rand.Read(randomByte)
		b[i] = charset[randomByte[0]%byte(len(charset))]
	}
	return string(b)
}

// Contains checks if a string contains a substring (case-sensitive)
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// ContainsIgnoreCase checks if a string contains a substring (case-insensitive)
func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// ContainsAny checks if a string contains any of the given substrings
func ContainsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// ContainsAll checks if a string contains all of the given substrings
func ContainsAll(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if !strings.Contains(s, substr) {
			return false
		}
	}
	return true
}

// Reverse reverses a string
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// CountWords counts the number of words in a string
func CountWords(s string) int {
	return len(strings.Fields(s))
}

// CountLines counts the number of lines in a string
func CountLines(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n"))
}

// WrapText wraps text to the specified width
func WrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	currentLine := ""

	for _, word := range words {
		if len(currentLine) == 0 {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

// PadLeft pads a string with the specified character on the left
func PadLeft(s string, length int, pad rune) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(string(pad), length-len(s)) + s
}

// PadRight pads a string with the specified character on the right
func PadRight(s string, length int, pad rune) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(string(pad), length-len(s))
}

// PadCenter pads a string with the specified character on both sides
func PadCenter(s string, length int, pad rune) string {
	if len(s) >= length {
		return s
	}

	totalPad := length - len(s)
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad

	return strings.Repeat(string(pad), leftPad) + s + strings.Repeat(string(pad), rightPad)
}

// RemoveDuplicateSpaces removes duplicate spaces from a string
func RemoveDuplicateSpaces(s string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(strings.TrimSpace(s), " ")
}

// IsValidEmail checks if a string is a valid email address
func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(url string) bool {
	re := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return re.MatchString(url)
}

// IsAlphanumeric checks if a string contains only alphanumeric characters
func IsAlphanumeric(s string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	return re.MatchString(s)
}

// IsNumeric checks if a string contains only numeric characters
func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// IsInteger checks if a string represents a valid integer
func IsInteger(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// ExtractNumbers extracts all numbers from a string
func ExtractNumbers(s string) []string {
	re := regexp.MustCompile(`\d+`)
	return re.FindAllString(s, -1)
}

// StripHTML removes HTML tags from a string
func StripHTML(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}

// Similarity calculates the similarity between two strings using Levenshtein distance
func Similarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	maxLen := math.Max(float64(len(s1)), float64(len(s2)))
	if maxLen == 0 {
		return 1.0
	}

	distance := levenshteinDistance(s1, s2)
	return 1.0 - (float64(distance) / maxLen)
}

func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}

	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min3(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// FormatBytes formats bytes into a human-readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
