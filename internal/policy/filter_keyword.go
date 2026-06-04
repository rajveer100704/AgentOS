package policy

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type KeywordFilter struct {
	name     string
	action   Action
	keywords []string
}

func NewKeywordFilter(name string, action Action, keywords []string) *KeywordFilter {
	normalized := make([]string, len(keywords))
	for i, k := range keywords {
		normalized[i] = normalizeText(k)
	}
	return &KeywordFilter{
		name:     name,
		action:   action,
		keywords: normalized,
	}
}

func (f *KeywordFilter) Name() string   { return f.name }
func (f *KeywordFilter) Action() Action { return f.action }

func (f *KeywordFilter) Check(content string) *Violation {
	normalized := normalizeText(content)
	for _, kw := range f.keywords {
		if strings.Contains(normalized, kw) {
			if f.action == ActionBlock {
				return &Violation{
					PolicyName: f.name,
					Action:     f.action,
					Message:    fmt.Sprintf("blocked keyword detected: %q", kw),
				}
			}
			return &Violation{
				PolicyName: f.name,
				Action:     f.action,
				Message:    fmt.Sprintf("keyword detected: %q", kw),
			}
		}
	}
	return nil
}

var multiSpaceRe = regexp.MustCompile(`\s+`)

// normalizeText applies NFKC Unicode normalization, lowercases, and collapses
// whitespace so that homoglyph substitutions (Cyrillic і vs Latin i) and
// spacing tricks cannot bypass keyword matching.
func normalizeText(s string) string {
	// NFKC decomposes compatibility characters and recomposes — this maps
	// lookalike codepoints (e.g. Cyrillic а→a after case fold) to their
	// canonical forms.
	s = norm.NFKC.String(s)
	// Map to lowercase using Unicode-aware case folding
	s = strings.Map(unicode.ToLower, s)
	// Collapse all whitespace runs to single space
	s = multiSpaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}
