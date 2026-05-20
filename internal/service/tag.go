package service

import (
	"strings"
	"unicode"
	"unicode/utf8"

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
	processed := name
	for _, sep := range []string{"-", "_", ".", ":", "/", "\\", "\t"} {
		processed = strings.ReplaceAll(processed, sep, " ")
	}
	parts := strings.Fields(processed)
	for _, p := range parts {
		clean := strings.TrimFunc(p, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		})
		if utf8.RuneCountInString(clean) >= 3 {
			seen[strings.ToLower(clean)] = true
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

	tagMap := make(map[string]string)
	tagMap["sync"] = "action"
	tagMap["copy"] = "action"
	tagMap["move"] = "action"
	for token, count := range tokenCounts {
		if count >= 2 {
			tagMap[token] = "keyword"
		}
	}

	return s.db.ReplaceTags(tagMap)
}

func (s *TagService) ListTags() ([]store.Tag, error) {
	return s.db.ListTags()
}