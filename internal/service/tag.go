package service

import (
	"strings"
	"unicode"

	"rcloneflow/internal/store"
)

type TagService struct {
	db *store.DB
}

func NewTagService(db *store.DB) *TagService {
	return &TagService{db: db}
}

type tokenFreq struct {
	token string
	count int
}

func containsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func (s *TagService) extractChineseTokens(name string) []tokenFreq {
	tokens := make(map[string]bool)
	runes := []rune(name)
	for i := 0; i < len(runes); i++ {
		if !unicode.Is(unicode.Han, runes[i]) {
			continue
		}
		end := i
		for end < len(runes) && unicode.Is(unicode.Han, runes[end]) {
			end++
		}
		seg := string(runes[i:end])
		charLen := end - i
		for width := 2; width <= charLen; width++ {
			for start := 0; start+width <= charLen; start++ {
				tokens[string([]rune(seg)[start:start+width])] = true
			}
		}
		i = end
	}
	result := make([]tokenFreq, 0, len(tokens))
	for t := range tokens {
		result = append(result, tokenFreq{token: t, count: 1})
	}
	return result
}

func (s *TagService) extractEnglishTokens(name string) []tokenFreq {
	seen := make(map[string]bool)
	lower := strings.ToLower(name)

	i := 0
	for i < len(lower) {
		r := rune(lower[i])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			i++
			continue
		}

		start := i
		for i < len(lower) {
			r := rune(lower[i])
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				break
			}
			i++
		}

		word := lower[start:i]
		if len(word) >= 3 {
			for w := 3; w <= len(word); w++ {
				for s := 0; s+w <= len(word); s++ {
					seen[word[s:s+w]] = true
				}
			}
		}
	}

	result := make([]tokenFreq, 0, len(seen))
	for t := range seen {
		result = append(result, tokenFreq{token: t, count: 1})
	}
	return result
}

func (s *TagService) extractTokens(name string) []tokenFreq {
	if containsChinese(name) {
		return s.extractChineseTokens(name)
	}
	return s.extractEnglishTokens(name)
}

func (s *TagService) RecalcTags() error {
	tasks, err := s.db.ListTasks()
	if err != nil {
		return err
	}

	tokenCounts := make(map[string]int)
	for _, task := range tasks {
		toks := s.extractTokens(task.Name)
		for _, tk := range toks {
			tokenCounts[tk.token]++
		}
	}

	entries := []store.TagEntry{
		{Tag: "sync", Type: "action", Selected: true},
		{Tag: "copy", Type: "action", Selected: true},
		{Tag: "move", Type: "action", Selected: true},
	}
	for token, count := range tokenCounts {
		if count >= 2 {
			entries = append(entries, store.TagEntry{Tag: token, Type: "keyword", Selected: false})
		}
	}

	return s.db.MergeTags(entries)
}

func (s *TagService) ListTags() ([]store.Tag, error) {
	return s.db.ListTags()
}

func (s *TagService) SelectTag(tag string, selected bool) error {
	return s.db.SetTagSelected(tag, selected)
}

func (s *TagService) CreateManualTag(tag string) error {
	return s.db.CreateManualTag(tag)
}

func (s *TagService) DeleteManualTag(tag string) error {
	return s.db.DeleteManualTag(tag)
}