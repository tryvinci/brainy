package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode/utf8"
)

var whitespaceRE = regexp.MustCompile(`\s+`)

// minEpisodeRunes: skip trivial fragments ("ok", "lol") but keep short dated facts.
const minEpisodeRunes = 12

type Extractor struct{}

func NewExtractor() Extractor {
	return Extractor{}
}

func (Extractor) Extract(req IngestRequest) []ExtractedMemory {
	var extracted []ExtractedMemory
	retainEpisodes := shouldRetainConversationEpisodes(req)
	for _, message := range req.Messages {
		for _, utterance := range splitUtterances(message.Content) {
			if memory, ok := classifySentence(utterance); ok {
				extracted = append(extracted, memory)
				continue
			}
			// Keep free dialogue searchable without breaking labeled pack ingest
			// (campaign / creative / analytics still fall through to pack_label_direct).
			if retainEpisodes {
				if episode, ok := conversationEpisode(utterance); ok {
					extracted = append(extracted, episode)
				}
			}
		}
	}
	return extracted
}

func shouldRetainConversationEpisodes(req IngestRequest) bool {
	if strings.TrimSpace(req.Label) != "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(req.SourceType)) {
	case "conversation", "chat", "dialogue", "message":
		return true
	default:
		return false
	}
}

func conversationEpisode(utterance string) (ExtractedMemory, bool) {
	text := NormalizeText(utterance)
	if text == "" || utf8.RuneCountInString(text) < minEpisodeRunes {
		return ExtractedMemory{}, false
	}
	return ExtractedMemory{
		Kind:       KindFact,
		Content:    text,
		SourceText: utterance,
		Confidence: 0.7,
		Explain: map[string]any{
			"rule":      "conversation_episode",
			"primitive": PrimitiveEpisode,
		},
	}, true
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

// splitUtterances prefers turn boundaries (newlines / Speaker: lines) so
// multi-turn client payloads stay atomic. Falls back to sentence splits.
func splitUtterances(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	lines := strings.Split(text, "\n")
	units := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		units = append(units, line)
	}
	if len(units) > 1 {
		return units
	}

	// Single line / block: sentence-split for preference+fact combos.
	return splitSentences(units[0])
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

	if strings.HasPrefix(lower, "never ") {
		return ExtractedMemory{
			Kind:       KindFact,
			Content:    titleSentence(NormalizeText(sentence)),
			SourceText: sentence,
			Confidence: 0.95,
			Explain: map[string]any{
				"rule": "constraint_never",
			},
		}, true
	}

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
			Content:    titleSentence(NormalizeText(sentence)),
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
			Content:    titleSentence(NormalizeText(sentence)),
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
	lowerNormalized := strings.ToLower(normalized)
	if strings.HasPrefix(lowerNormalized, "prefer ") || strings.HasPrefix(lowerNormalized, "prefers ") {
		return "Prefers " + strings.TrimSpace(normalized[strings.Index(lowerNormalized, "prefer")+6:])
	}
	return titleSentence(normalized)
}

func titleSentence(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
