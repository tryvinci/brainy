package memory

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"brainy/internal/pack"
)

// sanitizeUTF8 strips invalid UTF-8 byte sequences so Postgres TEXT inserts
// cannot fail with SQLSTATE 22021 (invalid_byte_sequence).
func sanitizeUTF8(value string) string {
	if value == "" || utf8.ValidString(value) {
		return value
	}
	return strings.ToValidUTF8(value, "")
}

// NormalizeIngestRequest mutates message contents to valid UTF-8.
func NormalizeIngestRequest(req *IngestRequest) {
	if req == nil {
		return
	}
	for i := range req.Messages {
		req.Messages[i].Content = sanitizeUTF8(req.Messages[i].Content)
		req.Messages[i].Role = sanitizeUTF8(req.Messages[i].Role)
	}
}

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
		Content:           sanitizeUTF8(extracted.Content),
		SourceText:        sanitizeUTF8(extracted.SourceText),
		SourceType:        req.SourceType,
		DedupeKey:         DedupeKey(req.TenantID, req.SubjectID, extracted.Kind, sanitizeUTF8(extracted.Content)),
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

	if entities := ExtractEntities(extracted.Content + " " + extracted.SourceText); len(entities) > 0 {
		if record.Metadata == nil {
			record.Metadata = map[string]any{}
		}
		record.Metadata["entities"] = entities
	}

	record.ObservedAt = ResolveObservedAt(record.Metadata, extracted.When)
	if record.ObservedAt != nil {
		record.Content = EnrichRelativeEventTime(record.Content, *record.ObservedAt)
	}
	// Link semantic objects to raw evidence captured at ingest.
	if req.Metadata != nil {
		if record.Metadata == nil {
			record.Metadata = map[string]any{}
		}
		if eid, ok := req.Metadata["evidence_id"].(string); ok && strings.TrimSpace(eid) != "" {
			record.Metadata["evidence_id"] = strings.TrimSpace(eid)
		}
		if raw, ok := req.Metadata["raw_evidence_ids"]; ok {
			record.Metadata["raw_evidence_ids"] = raw
			if _, has := record.Metadata["evidence_id"]; !has {
				switch ids := raw.(type) {
				case []string:
					if len(ids) > 0 {
						record.Metadata["evidence_id"] = ids[0]
					}
				case []any:
					if len(ids) > 0 {
						if s, ok := ids[0].(string); ok && s != "" {
							record.Metadata["evidence_id"] = s
						}
					}
				}
			}
		}
	}
	if extracted.Explain != nil {
		if pred, ok := extracted.Explain["predicate"].(string); ok && pred != "" {
			if record.Metadata == nil {
				record.Metadata = map[string]any{}
			}
			record.Metadata["predicate"] = pred
			if vn, ok := extracted.Explain["value_norm"].(string); ok {
				record.Metadata["value_norm"] = vn
			}
		}
	}
	// Lineage: ingest metadata.supersedes_memory_id → new.supersedes_id; prior
	// is marked superseded after upsert (Service.applyIngestSupersession).
	if sid := supersedesMemoryIDFromMetadata(record.Metadata); sid != "" {
		record.SupersedesID = sid
	} else if sid := supersedesMemoryIDFromMetadata(req.Metadata); sid != "" {
		if record.Metadata == nil {
			record.Metadata = map[string]any{}
		}
		record.Metadata["supersedes_memory_id"] = sid
		record.SupersedesID = sid
	} else if extracted.Explain != nil {
		if sid, ok := extracted.Explain["supersedes_memory_id"].(string); ok && strings.TrimSpace(sid) != "" {
			if record.Metadata == nil {
				record.Metadata = map[string]any{}
			}
			sid = strings.TrimSpace(sid)
			record.Metadata["supersedes_memory_id"] = sid
			record.SupersedesID = sid
		}
	}
	return record, nil
}

// EnrichRelativeEventTime appends an absolute date when dialogue uses relative
// time words ("yesterday", "last Saturday", "10 years ago") so search/answer
// can resolve event time against session observed_at.
func EnrichRelativeEventTime(content string, at time.Time) string {
	lower := strings.ToLower(content)
	var event time.Time
	switch {
	case strings.Contains(lower, "day before yesterday"):
		event = at.AddDate(0, 0, -2)
	case strings.Contains(lower, "two days ago"), strings.Contains(lower, "2 days ago"):
		event = at.AddDate(0, 0, -2)
	case strings.Contains(lower, "three days ago"), strings.Contains(lower, "3 days ago"):
		event = at.AddDate(0, 0, -3)
	case strings.Contains(lower, "a few days ago"), strings.Contains(lower, "couple days ago"), strings.Contains(lower, "couple of days ago"):
		event = at.AddDate(0, 0, -3)
	case strings.Contains(lower, "yesterday"):
		event = at.AddDate(0, 0, -1)
	case strings.Contains(lower, "tomorrow"):
		event = at.AddDate(0, 0, 1)
	case strings.Contains(lower, "last week"), strings.Contains(lower, "a week ago"), strings.Contains(lower, "1 week ago"):
		event = at.AddDate(0, 0, -7)
	case strings.Contains(lower, "two weeks ago"), strings.Contains(lower, "2 weeks ago"):
		event = at.AddDate(0, 0, -14)
	case strings.Contains(lower, "next week"):
		event = at.AddDate(0, 0, 7)
	case strings.Contains(lower, "last month"), strings.Contains(lower, "a month ago"):
		event = at.AddDate(0, -1, 0)
	case strings.Contains(lower, "next month"):
		event = at.AddDate(0, 1, 0)
	case strings.Contains(lower, "last year"), strings.Contains(lower, "a year ago"), strings.Contains(lower, "1 year ago"):
		event = at.AddDate(-1, 0, 0)
	case strings.Contains(lower, "next year"):
		event = at.AddDate(1, 0, 0)
	case strings.Contains(lower, "this weekend"), strings.Contains(lower, "coming weekend"), strings.Contains(lower, "next weekend"):
		event = nextWeekend(at)
	case strings.Contains(lower, "last weekend"):
		event = previousWeekend(at)
	case strings.Contains(lower, "the week before"), strings.Contains(lower, "a week before"):
		event = at.AddDate(0, 0, -7)
	case strings.Contains(lower, "today"), strings.Contains(lower, "this week"), strings.Contains(lower, "this month"):
		event = at
	default:
		if days := parseDaysOffset(lower); days != 0 {
			event = at.AddDate(0, 0, days)
			break
		}
		if monthEvent, ok := parseMonthAgainstObserved(lower, at); ok {
			event = monthEvent
			break
		}
		if years := parseYearsAgo(lower); years > 0 {
			event = at.AddDate(-years, 0, 0)
			break
		}
		if wd, ok := parseLastWeekday(lower); ok {
			event = previousWeekday(at, wd)
			break
		}
		if wd, ok := parseThisWeekday(lower); ok {
			event = thisWeekday(at, wd)
			break
		}
		return content
	}
	if event.IsZero() {
		return content
	}
	stamp := event.Format("2 January 2006")
	if strings.Contains(lower, strings.ToLower(stamp)) {
		return content
	}
	// Also expose ISO year for "how long ago" style questions.
	extra := stamp
	if years := parseYearsAgo(lower); years > 0 {
		extra = stamp + "; " + strconv.Itoa(years) + " years ago"
	}
	return content + " (" + extra + ")"
}

func previousWeekday(at time.Time, day time.Weekday) time.Time {
	d := at
	for i := 0; i < 7; i++ {
		d = d.AddDate(0, 0, -1)
		if d.Weekday() == day {
			return d
		}
	}
	return at.AddDate(0, 0, -7)
}

func thisWeekday(at time.Time, day time.Weekday) time.Time {
	d := at
	for i := 0; i < 7; i++ {
		if d.Weekday() == day {
			return d
		}
		d = d.AddDate(0, 0, -1)
	}
	return at
}

func parseLastWeekday(lower string) (time.Weekday, bool) {
	mapping := []struct {
		needle string
		day    time.Weekday
	}{
		{"last sunday", time.Sunday},
		{"last monday", time.Monday},
		{"last tuesday", time.Tuesday},
		{"last wednesday", time.Wednesday},
		{"last thursday", time.Thursday},
		{"last friday", time.Friday},
		{"last saturday", time.Saturday},
	}
	for _, item := range mapping {
		if strings.Contains(lower, item.needle) {
			return item.day, true
		}
	}
	return 0, false
}

func parseThisWeekday(lower string) (time.Weekday, bool) {
	mapping := []struct {
		needle string
		day    time.Weekday
	}{
		{"this sunday", time.Sunday},
		{"this monday", time.Monday},
		{"this tuesday", time.Tuesday},
		{"this wednesday", time.Wednesday},
		{"this thursday", time.Thursday},
		{"this friday", time.Friday},
		{"this saturday", time.Saturday},
	}
	for _, item := range mapping {
		if strings.Contains(lower, item.needle) {
			return item.day, true
		}
	}
	return 0, false
}

func nextWeekend(at time.Time) time.Time {
	d := at
	for i := 0; i < 8; i++ {
		if d.Weekday() == time.Saturday {
			return d
		}
		d = d.AddDate(0, 0, 1)
	}
	return at.AddDate(0, 0, 6)
}

func previousWeekend(at time.Time) time.Time {
	d := at
	for i := 0; i < 8; i++ {
		d = d.AddDate(0, 0, -1)
		if d.Weekday() == time.Saturday {
			return d
		}
	}
	return at.AddDate(0, 0, -7)
}

func parseDaysOffset(lower string) int {
	// "in 3 days", "in two weeks", "3 days from now"
	words := map[string]int{"one": 1, "two": 2, "three": 3, "four": 4, "five": 5, "six": 6, "seven": 7}
	for word, n := range words {
		if strings.Contains(lower, "in "+word+" days") || strings.Contains(lower, word+" days from now") {
			return n
		}
		if strings.Contains(lower, "in "+word+" weeks") || strings.Contains(lower, word+" weeks from now") {
			return n * 7
		}
	}
	for n := 1; n <= 14; n++ {
		if strings.Contains(lower, "in "+strconv.Itoa(n)+" days") || strings.Contains(lower, strconv.Itoa(n)+" days from now") {
			return n
		}
		if strings.Contains(lower, "in "+strconv.Itoa(n)+" weeks") {
			return n * 7
		}
	}
	return 0
}

func parseMonthAgainstObserved(lower string, at time.Time) (time.Time, bool) {
	months := []struct {
		name  string
		month time.Month
	}{
		{"january", time.January}, {"february", time.February}, {"march", time.March},
		{"april", time.April}, {"may", time.May}, {"june", time.June},
		{"july", time.July}, {"august", time.August}, {"september", time.September},
		{"october", time.October}, {"november", time.November}, {"december", time.December},
	}
	for _, m := range months {
		// Require planning/when context so bare month names in titles don't fire.
		if !strings.Contains(lower, m.name) {
			continue
		}
		if !(strings.Contains(lower, "in "+m.name) || strings.Contains(lower, "next "+m.name) ||
			strings.Contains(lower, "early "+m.name) || strings.Contains(lower, "late "+m.name) ||
			strings.Contains(lower, "planning") || strings.Contains(lower, "going")) {
			continue
		}
		year := at.Year()
		candidate := time.Date(year, m.month, 15, 0, 0, 0, 0, time.UTC)
		if strings.Contains(lower, "next "+m.name) && candidate.Before(at) {
			candidate = candidate.AddDate(1, 0, 0)
		} else if candidate.Before(at.AddDate(0, -6, 0)) {
			candidate = candidate.AddDate(1, 0, 0)
		}
		return candidate, true
	}
	return time.Time{}, false
}

func parseYearsAgo(lower string) int {
	words := map[string]int{
		"one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
		"six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
	}
	for word, n := range words {
		if strings.Contains(lower, word+" years ago") || strings.Contains(lower, word+" year ago") {
			return n
		}
	}
	for n := 1; n <= 20; n++ {
		needle := strconv.Itoa(n) + " years ago"
		if n == 1 {
			needle = "1 year ago"
		}
		if strings.Contains(lower, needle) || strings.Contains(lower, strconv.Itoa(n)+" years ago") {
			return n
		}
	}
	return 0
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
