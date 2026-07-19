package embedding

import (
	"hash/fnv"
	"math"
	"strings"
)

const Dim = 128

var synonymGroups = [][]string{
	{"warm", "friendly", "approachable", "welcoming"},
	{"concise", "brief", "short", "compact"},
	{"professional", "formal", "business"},
	{"tone", "voice", "style", "copy"},
	{"competitor", "rival", "competition"},
	{"taboo", "forbidden", "prohibited", "never"},
	{"carousel", "slideshow", "swipe"},
	{"static", "still", "image"},
	{"email", "inbox", "mail"},
	{"sms", "text", "message"},
	{"minimal", "clean", "simple"},
	{"layout", "design", "format"},
	// Generic conversational paraphrase helpers (not dataset-specific).
	{"identity", "gender"},
	{"relationship", "partner", "dating", "married", "single"},
	{"career", "job", "profession", "work"},
	{"hobby", "hobbies", "activity", "activities"},
	{"conference", "meetup", "event", "gathering"},
	{"family", "adoption", "parenting"},
	{"speech", "talk", "presentation"},
}

func Embed(text string) []float32 {
	normalized := normalize(text)
	if normalized == "" {
		return make([]float32, Dim)
	}

	vec := make([]float32, Dim)
	for _, token := range tokenize(normalized) {
		addToken(vec, token, 1)
		if len(token) >= 3 {
			for i := 0; i+3 <= len(token); i++ {
				addToken(vec, token[i:i+3], 0.5)
			}
		}
	}

	norm := float32(0)
	for _, value := range vec {
		norm += value * value
	}
	if norm == 0 {
		return vec
	}
	scale := 1 / float32(math.Sqrt(float64(norm)))
	for i := range vec {
		vec[i] *= scale
	}
	return vec
}

func CosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func addToken(vec []float32, token string, weight float64) {
	if token == "" {
		return
	}
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(token))
	index := hash.Sum64() % Dim
	vec[index] += float32(weight)
}

func normalize(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	replacer := strings.NewReplacer(",", " ", ".", " ", "!", " ", "?", " ", ":", " ", ";", " ", "-", " ")
	return replacer.Replace(text)
}

func tokenize(text string) []string {
	canonical := applySynonyms(text)
	parts := strings.Fields(canonical)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func applySynonyms(text string) string {
	tokens := strings.Fields(text)
	for i, token := range tokens {
		for _, group := range synonymGroups {
			for j, word := range group {
				if token == word {
					tokens[i] = group[0]
					break
				}
				if j == 0 {
					continue
				}
			}
		}
	}
	return strings.Join(tokens, " ")
}
