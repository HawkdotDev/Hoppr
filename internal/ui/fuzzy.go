package ui

import (
	"strings"
)

// Levenshtein calculates the edit distance between two strings.
func Levenshtein(a, b string) int {
	a, b = strings.ToLower(a), strings.ToLower(b)
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			curr[j] = min(
				curr[j-1]+1,     // insertion
				prev[j]+1,       // deletion
				prev[j-1]+cost,  // substitution
			)
		}
		copy(prev, curr)
	}

	return curr[lb]
}

// FindClosestMatch finds the most similar string from a list of candidates.
// Returns empty string if no candidate is within a reasonable distance (maxDistance).
func FindClosestMatch(input string, candidates []string, maxDistance int) string {
	bestMatch := ""
	lowestDistance := maxDistance + 1

	inputLower := strings.ToLower(input)
	for _, candidate := range candidates {
		candLower := strings.ToLower(candidate)
		if candLower == inputLower {
			return candidate
		}

		dist := Levenshtein(inputLower, candLower)
		if dist < lowestDistance {
			lowestDistance = dist
			bestMatch = candidate
		}
	}

	if lowestDistance <= maxDistance {
		return bestMatch
	}
	return ""
}
