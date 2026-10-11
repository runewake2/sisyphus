package main

import (
	"math/rand"
	"regexp"
	"strings"
	"testing"
)

// fieldReference is the regular expression that parseField replaced. The test keeps it as the
// definition of the expected result.
var fieldReference = regexp.MustCompile(`^([A-Za-z_-]+):\s*("(?:[^"\\]|\\.)*"|'[^']*'|[^#]*?)\s*(#.*)?$`)

func checkParseField(t *testing.T, line string) {
	t.Helper()
	got, ok := parseField(line)
	match := fieldReference.FindStringSubmatch(line)
	if (match != nil) != ok {
		t.Fatalf("%q: parseField ok=%v, regexp matched=%v", line, ok, match != nil)
	}
	if match != nil && (got.key != match[1] || got.value != match[2] || got.comment != match[3]) {
		t.Fatalf("%q: parseField %#v, regexp key=%q value=%q comment=%q", line, got, match[1], match[2], match[3])
	}
}

func TestParseFieldMatchesTheRegexpOnTheTemplate(t *testing.T) {
	doc, err := parseDocument(kitTemplate(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range doc.front {
		checkParseField(t, line)
	}
}

func TestParseFieldMatchesTheRegexpOnRandomLines(t *testing.T) {
	pieces := []string{"title", "depends-on", "Key_9", "x", ":", " ", "\t", "\f", "\"", "'", "\\", "\\\"", "#", " # c",
		"[[a]]", "[", "]", "{", "}", "a, b", "é", "日本", "\"x\"", "'y'", "\"a\\\"b\"", "value", "  ", "\r", "-"}
	random := rand.New(rand.NewSource(1))
	for i := 0; i < 200000; i++ {
		var b strings.Builder
		if random.Intn(4) > 0 {
			b.WriteString([]string{"title", "state", "depends-on", "meta_data", "x"}[random.Intn(5)])
			b.WriteString(":")
		}
		for n := random.Intn(8); n > 0; n-- {
			b.WriteString(pieces[random.Intn(len(pieces))])
		}
		checkParseField(t, b.String())
	}
}

func TestParseFieldOnKnownLines(t *testing.T) {
	for _, line := range []string{
		`title: "Fix: colon \"quotes\" #hash"`,
		`title: "a" junk # comment`,
		`title: 'single # not a comment'  # comment`,
		`tags: [bug, ui] # Built-in: bug, feature.`,
		`metadata: {note: "x, y"}`,
		`owner:                 # The person`,
		`depends-on: []           # Optional.`,
		`bad key: value`,
		`: value`,
		`title: "unterminated`,
		`title: "ends with backslash\`,
	} {
		checkParseField(t, line)
	}
}
