package main

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

var ankiSoundMarker = regexp.MustCompile(`\[sound:[^\]]*\]`)

// Tokenize card HTML into text without rendering scripts, styles, or ruby hints.
// The original field values are retained separately in SQLite.
func plainAnkiText(value string) string {
	tokens := html.NewTokenizer(strings.NewReader(ankiSoundMarker.ReplaceAllString(value, "")))
	var text strings.Builder
	skip := 0
	for {
		kind := tokens.Next()
		if kind == html.ErrorToken {
			break
		}
		token := tokens.Token()
		tag := strings.ToLower(token.Data)
		suppressed := tag == "script" || tag == "style" || tag == "rt" || tag == "rp"
		switch kind {
		case html.StartTagToken:
			if suppressed {
				skip++
			}
		case html.EndTagToken:
			if suppressed && skip > 0 {
				skip--
			}
		case html.TextToken:
			if skip == 0 {
				text.WriteString(token.Data)
			}
		}
		if skip == 0 && (kind == html.StartTagToken || kind == html.EndTagToken || kind == html.SelfClosingTagToken) {
			switch tag {
			case "br", "div", "p", "li", "hr":
				text.WriteByte('\n')
			}
		}
	}
	return strings.TrimSpace(text.String())
}

func extractAnkiReading(vocabulary, answer string) string {
	var lines []string
	for _, line := range strings.Split(answer, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) > 0 {
		candidate := lines[0]
		if strings.HasPrefix(candidate, "【読み】") {
			candidate = strings.TrimSpace(strings.TrimPrefix(candidate, "【読み】"))
			if candidate == "" && len(lines) > 1 {
				candidate = lines[1]
			}
		}
		if isKanaReading(candidate) {
			return candidate
		}
	}
	if isKanaReading(vocabulary) {
		return vocabulary
	}
	return ""
}

func isKanaReading(value string) bool {
	hasKana := false
	for _, character := range value {
		if unicode.Is(unicode.Hiragana, character) || unicode.Is(unicode.Katakana, character) {
			hasKana = true
			continue
		}
		if !unicode.IsSpace(character) && !strings.ContainsRune("ー・()（）/／、〜～-", character) {
			return false
		}
	}
	return hasKana
}
