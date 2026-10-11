package main

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// maxSuggestions is the largest number of issues that the error for an unknown name suggests.
const maxSuggestions = 3

// suggestIssues returns the text that the error for an unknown issue name adds: up to 3 existing
// issues whose names and titles share the most words with name, for example
// " Did you mean: update-docs-site (Update the docs site theme)?". It returns "" when no issue shares
// a word. Models guess names, and the list lets them correct the call.
func suggestIssues(x *issueIndex, name string) string {
	wanted := meaningfulWords(name)
	if len(wanted) == 0 {
		return ""
	}
	type candidate struct {
		issue issueFile
		title string
		score float64
	}
	issues := x.all()
	x.preload(issues, false)
	var candidates []candidate
	for _, issue := range issues {
		title := ""
		if doc, err := x.header(issue); err == nil {
			title = doc.get("title")
		}
		have := meaningfulWords(issue.name + " " + title)
		score := 0.0
		for _, word := range wanted {
			score += wordMatch(word, have)
		}
		if score >= 1 {
			candidates = append(candidates, candidate{issue, title, score})
		}
	}
	slices.SortStableFunc(candidates, func(a, b candidate) int {
		if a.score != b.score {
			if a.score > b.score {
				return -1
			}
			return 1
		}
		// An issue that is not closed is the more likely target.
		if closedA, closedB := a.issue.state == "closed", b.issue.state == "closed"; closedA != closedB {
			if closedB {
				return -1
			}
			return 1
		}
		return strings.Compare(a.issue.name, b.issue.name)
	})
	if len(candidates) == 0 {
		return ""
	}
	var names []string
	for _, c := range candidates[:min(maxSuggestions, len(candidates))] {
		label := c.title
		if c.issue.state == "closed" {
			label = "closed: " + label
		}
		names = append(names, fmt.Sprintf("%s (%s)", c.issue.name, label))
	}
	return " Did you mean: " + strings.Join(names, ", ") + "?"
}

// meaningfulWords returns the lowercase words of text, without the filler words of names.
func meaningfulWords(text string) []string {
	var words []string
	for _, word := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if !fillerWords[word] && !slices.Contains(words, word) {
			words = append(words, word)
		}
	}
	return words
}

// wordMatch scores how well word matches the best of words: 1 for the same word, and 0.75 for a
// near word, for example "convention" and "conventions" or "docs" and "doc".
func wordMatch(word string, words []string) float64 {
	best := 0.0
	for _, other := range words {
		switch {
		case other == word:
			return 1
		case len(word) >= 3 && len(other) >= 3 && (strings.HasPrefix(other, word) || strings.HasPrefix(word, other)),
			len(word) >= 5 && editDistance(word, other) <= 1:
			best = 0.75
		}
	}
	return best
}
