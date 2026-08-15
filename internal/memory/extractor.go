package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode/utf8"
)

var whitespaceRE = regexp.MustCompile(`\s+`)

// minEpisodeRunes: skip trivial fragments ("ok", "lol") but keep short dated facts.
const minEpisodeRunes = 12

// Extractor turns an ingest request into candidate memories.
// Sync ingest uses DeterministicExtractor (no network).
// The async worker may use ProviderExtractor when configured.
type Extractor interface {
	Extract(ctx context.Context, req IngestRequest) ([]ExtractedMemory, error)
}

type DeterministicExtractor struct{}

func NewDeterministicExtractor() DeterministicExtractor {
	return DeterministicExtractor{}
}

// NewExtractor returns the deterministic extractor (CI / sync default).
func NewExtractor() DeterministicExtractor {
	return NewDeterministicExtractor()
}

func (DeterministicExtractor) Extract(ctx context.Context, req IngestRequest) ([]ExtractedMemory, error) {
	req.Messages = EnrichImageText(ctx, req.Messages)
	var extracted []ExtractedMemory
	retainEpisodes := shouldRetainConversationEpisodes(req)
	var allUtterances []string
	for _, message := range req.Messages {
		assistant := isAssistantRole(message.Role)
		for _, utterance := range splitUtterances(message.Content) {
			allUtterances = append(allUtterances, utterance)
			if memory, ok := classifySentence(utterance); ok {
				extracted = append(extracted, memory)
				continue
			}
			// Keep free dialogue searchable without breaking labeled pack ingest
			// (campaign / creative / analytics still fall through to pack_label_direct).
			if !retainEpisodes {
				continue
			}
			if assistant {
				if isPhaticAssistantText(utterance) {
					continue
				}
				if fact, ok := assistantStatedMemory(utterance); ok {
					extracted = append(extracted, fact)
				}
				continue
			}
			if dated, ok := datedEventMemory(utterance); ok {
				extracted = append(extracted, dated)
				continue
			}
			if episode, ok := conversationEpisode(utterance); ok {
				extracted = append(extracted, episode)
			}
		}
	}
	// Attribute atoms: standalone searchable facts (identity, origin, titles,
	// activities). Closes the Mem0-style ADD-fact gap for conversational ingest.
	if retainEpisodes {
		extracted = append(extracted, extractAttributeAtoms(allUtterances, ResolveObservedAt(req.Metadata, ""))...)
	}
	return filterAssistantRecallEpisodes(req, extracted), nil
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

var yearTokenRE = regexp.MustCompile(`\b(?:19|20)\d{2}\b`)

// datedEventMemory copies a dated turn as a recall-primary fact so search
// does not depend on provenance episodes (Mem0 stores the event, not the turn).
func datedEventMemory(utterance string) (ExtractedMemory, bool) {
	text := NormalizeText(utterance)
	if text == "" || !yearTokenRE.MatchString(text) {
		return ExtractedMemory{}, false
	}
	return ExtractedMemory{
		Kind:       KindFact,
		Content:    text,
		SourceText: utterance,
		Confidence: 0.8,
		Explain: map[string]any{
			"rule": "dated_event",
		},
	}, true
}

func assistantStatedMemory(utterance string) (ExtractedMemory, bool) {
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
			"rule": "assistant_stated",
		},
	}, true
}

func isAssistantRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "assistant", "system":
		return true
	default:
		return false
	}
}

func isPhaticAssistantText(text string) bool {
	lower := strings.ToLower(text)
	cues := []string{
		"congratulations",
		"you're welcome",
		"you are welcome",
		"happy to help",
		"glad i could help",
		"let me know if",
		"that's a great",
		"that's great",
		"that's awesome",
		"that's wonderful",
		"fantastic idea",
	}
	for _, cue := range cues {
		if strings.Contains(lower, cue) {
			return true
		}
	}
	return false
}

func isEpisodeLike(m ExtractedMemory) bool {
	if m.Explain == nil {
		return false
	}
	if r, _ := m.Explain["rule"].(string); r == "conversation_episode" {
		return true
	}
	if p, _ := m.Explain["primitive"].(string); p == PrimitiveEpisode {
		return true
	}
	return false
}

func actorRoleForText(req IngestRequest, text string) string {
	n := strings.ToLower(NormalizeText(text))
	if n == "" {
		return ""
	}
	for _, m := range req.Messages {
		mc := strings.ToLower(NormalizeText(m.Content))
		if mc == "" {
			continue
		}
		if mc == n || strings.Contains(mc, n) || strings.Contains(n, mc) {
			return strings.ToLower(strings.TrimSpace(m.Role))
		}
	}
	return ""
}

func filterAssistantRecallEpisodes(req IngestRequest, extracted []ExtractedMemory) []ExtractedMemory {
	out := make([]ExtractedMemory, 0, len(extracted))
	for _, m := range extracted {
		role := actorRoleForText(req, firstNonEmpty(m.SourceText, m.Content))
		phatic := isPhaticAssistantText(m.Content) || isPhaticAssistantText(m.SourceText)
		if isAssistantRole(role) || (role == "" && phatic && isEpisodeLike(m)) {
			if phatic {
				continue
			}
			if isEpisodeLike(m) {
				if m.Explain == nil {
					m.Explain = map[string]any{}
				}
				m.Explain["rule"] = "assistant_stated"
				delete(m.Explain, "primitive")
			}
		}
		out = append(out, m)
	}
	return out
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
	vis := ""
	if i := strings.Index(text, visibleTextMarker); i >= 0 {
		vis = strings.TrimSpace(text[i:])
		text = strings.TrimSpace(text[:i])
	}
	normalized := strings.NewReplacer("!", ".", "?", ".").Replace(text)
	parts := strings.Split(normalized, ".")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = NormalizeText(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if vis != "" {
		if len(out) == 0 {
			return []string{vis}
		}
		for i := range out {
			out[i] = strings.TrimSpace(out[i] + " " + vis)
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
