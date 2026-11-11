package filter

import (
	"strings"
	"unicode"
)

// Deduplicator checks for semantic redundancy using token n-gram Jaccard similarity.
type Deduplicator struct {
	threshold  float64
	seenNGrams []map[string]struct{}
}

// NewDeduplicator creates a deduplicator with a given similarity cutoff (e.g. 0.85).
func NewDeduplicator(threshold float64) *Deduplicator {
	if threshold <= 0 || threshold > 1.0 {
		threshold = 0.85
	}
	return &Deduplicator{
		threshold:  threshold,
		seenNGrams: make([]map[string]struct{}, 0),
	}
}

// IsDuplicate checks whether the text has similarity higher than threshold against previously seen texts.
// If it is unique, it records the text n-grams and returns false.
func (d *Deduplicator) IsDuplicate(text string) bool {
	ngrams := extract3Grams(text)
	if len(ngrams) == 0 {
		return false
	}

	for _, seen := range d.seenNGrams {
		sim := jaccardSimilarity(ngrams, seen)
		if sim >= d.threshold {
			return true
		}
	}

	d.seenNGrams = append(d.seenNGrams, ngrams)
	return false
}

func extract3Grams(s string) map[string]struct{} {
	words := tokenize(s)
	grams := make(map[string]struct{})
	if len(words) < 3 {
		for _, w := range words {
			grams[w] = struct{}{}
		}
		return grams
	}

	for i := 0; i <= len(words)-3; i++ {
		gram := words[i] + " " + words[i+1] + " " + words[i+2]
		grams[gram] = struct{}{}
	}
	return grams
}

func tokenize(s string) []string {
	var words []string
	var current strings.Builder

	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			words = append(words, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}

func jaccardSimilarity(a, b map[string]struct{}) float64 {
	intersection := 0
	for k := range a {
		if _, exists := b[k]; exists {
			intersection++
		}
	}

	union := len(a) + len(b) - intersection
	if union == 0 {
		return 1.0
	}
	return float64(intersection) / float64(union)
}
