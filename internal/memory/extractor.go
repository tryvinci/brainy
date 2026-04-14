package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

var whitespaceRE = regexp.MustCompile(`\s+`)

type Extractor struct{}

func NewExtractor() Extractor {
	return Extractor{}
}

func (Extractor) Extract(req IngestRequest) []ExtractedMemory {
	var extracted []ExtractedMemory
	for _, message := range req.Messages {
		for _, sentence := range splitSentences(message.Content) {
			if memory, ok := classifySentence(sentence); ok {
				extracted = append(extracted, memory)
			}
		}
	}
	return extracted
}

func NormalizeText(text string) string {
	text = strings.TrimSpace(text)
	text = whitespaceRE.ReplaceAllString(text, " ")
	return text
}

func DedupeKey(tenantID, subjectID, kind, content string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		tenantID,
		subjectID,
		kind,
		strings.ToLower(NormalizeText(content)),
	}, "::")))
	return hex.EncodeToString(sum[:])
}

func splitSentences(text string) []string {
	normalized := strings.NewReplacer("!", ".", "?", ".").Replace(text)
	parts := strings.Split(normalized, ".")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = NormalizeText(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func classifySentence(sentence string) (ExtractedMemory, bool) {
	lower := strings.ToLower(sentence)

	if strings.Contains(lower, "prefer ") || strings.Contains(lower, "prefers ") {
		return ExtractedMemory{
			Kind:       KindPreference,
			Content:    canonicalizePreference(sentence),
			SourceText: sentence,
			Confidence: 0.92,
			Explain: map[string]any{
				"rule": "preference_prefer",
			},
		}, true
	}

	if strings.Contains(lower, "like ") || strings.Contains(lower, "love ") || strings.Contains(lower, "hate ") || strings.Contains(lower, "dislike ") {
		return ExtractedMemory{
			Kind:       KindPreference,
			Content:    canonicalizePreference(sentence),
			SourceText: sentence,
			Confidence: 0.88,
			Explain: map[string]any{
				"rule": "preference_sentiment",
			},
		}, true
	}

	if strings.HasPrefix(lower, "i am ") || strings.HasPrefix(lower, "i'm ") || strings.HasPrefix(lower, "my name is ") || strings.Contains(lower, "i work at ") || strings.Contains(lower, "i live in ") {
		return ExtractedMemory{
			Kind:       KindProfile,
			Content:    canonicalizeProfile(sentence),
			SourceText: sentence,
			Confidence: 0.9,
			Explain: map[string]any{
				"rule": "profile_identity",
			},
		}, true
	}

	if strings.Contains(lower, " is ") || strings.Contains(lower, " are ") || strings.Contains(lower, " date ") || strings.Contains(lower, "launch") {
		return ExtractedMemory{
			Kind:       KindFact,
			Content:    canonicalizeFact(sentence),
			SourceText: sentence,
			Confidence: 0.78,
			Explain: map[string]any{
				"rule": "fact_statement",
			},
		}, true
	}

	return ExtractedMemory{}, false
}

func canonicalizePreference(sentence string) string {
	normalized := NormalizeText(sentence)
	normalized = strings.TrimPrefix(normalized, "I ")
	normalized = strings.TrimPrefix(normalized, "i ")
	normalized = strings.TrimPrefix(normalized, "I'm ")
	normalized = strings.TrimPrefix(normalized, "i'm ")
	if strings.HasPrefix(strings.ToLower(normalized), "prefer ") || strings.HasPrefix(strings.ToLower(normalized), "prefers ") {
		return "Prefers " + strings.TrimSpace(normalized[strings.Index(strings.ToLower(normalized), "prefer")+6:])
	}
	return titleSentence(normalized)
}

func canonicalizeProfile(sentence string) string {
	return titleSentence(NormalizeText(sentence))
}

func canonicalizeFact(sentence string) string {
	return titleSentence(NormalizeText(sentence))
}

func titleSentence(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
