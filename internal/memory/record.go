package memory

import (
	"strings"
	"time"

	"brainy/internal/pack"
)

// BuildMemoryRecord materializes an extracted memory into a storeable record.
// Shared by sync Service ingest and async job Processor so behavior stays aligned.
func BuildMemoryRecord(memoryID string, now time.Time, req IngestRequest, extracted ExtractedMemory, packs *pack.Registry) (MemoryRecord, error) {
	version := "deterministic-v1"
	if rule, _ := extracted.Explain["rule"].(string); rule == "provider_extract" {
		version = providerExtractionVersion
	}

	record := MemoryRecord{
		MemoryID:          memoryID,
		TenantID:          req.TenantID,
		SubjectID:         req.SubjectID,
		Kind:              extracted.Kind,
		Content:           extracted.Content,
		SourceText:        extracted.SourceText,
		SourceType:        req.SourceType,
		DedupeKey:         DedupeKey(req.TenantID, req.SubjectID, extracted.Kind, extracted.Content),
		Status:            StatusActive,
		Confidence:        extracted.Confidence,
		ExtractionVersion: version,
		Explain:           extracted.Explain,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := ApplyVerticalPack(&record, req, extracted.Kind, extracted.Content, packs); err != nil {
		return MemoryRecord{}, err
	}

	if rule, _ := extracted.Explain["rule"].(string); rule == "conversation_episode" {
		record.Primitive = PrimitiveEpisode
		record.ExtractionVersion = "conversational-v1"
		if p, ok := extracted.Explain["primitive"].(string); ok && p != "" {
			record.Primitive = p
		}
	}
	if p, ok := extracted.Explain["primitive"].(string); ok && p != "" && record.Primitive == "" {
		record.Primitive = p
	}

	if extracted.When != "" {
		if record.Metadata == nil {
			record.Metadata = map[string]any{}
		}
		record.Metadata["when"] = extracted.When
	}
	if extracted.Duration != "" {
		if record.Metadata == nil {
			record.Metadata = map[string]any{}
		}
		record.Metadata["duration"] = extracted.Duration
	}

	record.ObservedAt = ResolveObservedAt(record.Metadata, extracted.When)
	if record.ObservedAt != nil {
		record.Content = EnrichRelativeEventTime(record.Content, *record.ObservedAt)
	}
	return record, nil
}

// EnrichRelativeEventTime appends an absolute date when dialogue uses relative
// time words ("yesterday", "last week") so search/answer can resolve event time.
func EnrichRelativeEventTime(content string, at time.Time) string {
	lower := strings.ToLower(content)
	relative := false
	for _, marker := range []string{
		"yesterday", "today", "tomorrow",
		"last week", "last month", "last year",
		"next week", "next month", "next year",
		"this week", "this month",
		"last saturday", "last sunday", "last monday", "last tuesday", "last wednesday", "last thursday", "last friday",
		"this saturday", "this sunday",
	} {
		if strings.Contains(lower, marker) {
			relative = true
			break
		}
	}
	if !relative {
		return content
	}
	stamp := at.Format("2 January 2006")
	if strings.Contains(lower, strings.ToLower(stamp)) {
		return content
	}
	return content + " (" + stamp + ")"
}

// ResolveObservedAt prefers metadata.observed_at, then a provider when slot.
func ResolveObservedAt(metadata map[string]any, when string) *time.Time {
	if metadata != nil {
		if raw, ok := metadata["observed_at"]; ok {
			if ts := parseFlexibleTime(raw); ts != nil {
				return ts
			}
		}
	}
	if when != "" {
		return parseFlexibleTime(when)
	}
	return nil
}

func parseFlexibleTime(raw any) *time.Time {
	switch v := raw.(type) {
	case time.Time:
		t := v.UTC()
		return &t
	case *time.Time:
		if v == nil {
			return nil
		}
		t := v.UTC()
		return &t
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return nil
		}
		layouts := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02",
			"2 January 2006",
			"2 Jan 2006",
			"January 2 2006",
			"Jan 2 2006",
			"2 May 2006",
			"7 May 2003",
			"2006/01/02",
			// LOCOMO session timestamps: "1:56 pm on 8 May, 2023"
			"3:04 pm on 2 January, 2006",
			"3:04 am on 2 January, 2006",
			"15:04 on 2 January, 2006",
			"3:04pm on 2 January, 2006",
			"3:04am on 2 January, 2006",
			"2 January, 2006",
			"2 Jan, 2006",
		}
		for _, layout := range layouts {
			if ts, err := time.Parse(layout, s); err == nil {
				t := ts.UTC()
				return &t
			}
		}
		// Fallback: strip leading clock time ("... on 8 May, 2023").
		if idx := strings.LastIndex(strings.ToLower(s), " on "); idx >= 0 {
			if ts := parseFlexibleTime(strings.TrimSpace(s[idx+4:])); ts != nil {
				return ts
			}
		}
		// Common dialogue form: "7 May 2023"
		if ts, err := time.Parse("2 January 2006", s); err == nil {
			t := ts.UTC()
			return &t
		}
		if ts, err := time.Parse("2 Jan 2006", s); err == nil {
			t := ts.UTC()
			return &t
		}
	}
	return nil
}

// EventTime returns observed_at when set, otherwise updated_at.
func EventTime(record MemoryRecord) time.Time {
	if record.ObservedAt != nil {
		return *record.ObservedAt
	}
	return record.UpdatedAt
}
