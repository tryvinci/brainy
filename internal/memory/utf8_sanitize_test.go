package memory

import "testing"

func TestSanitizeUTF8StripsInvalidBytes(t *testing.T) {
	raw := "hello\x80world"
	got := sanitizeUTF8(raw)
	if got != "helloworld" {
		t.Fatalf("sanitizeUTF8=%q", got)
	}
	if sanitizeUTF8("ok") != "ok" {
		t.Fatal("valid utf8 changed")
	}
}

func TestBuildMemoryRecordSanitizesInvalidUTF8(t *testing.T) {
	record, err := BuildMemoryRecord("mem_utf8", mustParseTime(t, "2026-08-11T00:00:00Z"), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
	}, ExtractedMemory{
		Kind:       KindFact,
		Content:    "bad\x80fact",
		SourceText: "src\x80text",
		Confidence: 0.9,
		Explain:    map[string]any{"rule": "provider_extract"},
	}, nil)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if record.Content != "badfact" {
		t.Fatalf("content=%q", record.Content)
	}
	if record.SourceText != "srctext" {
		t.Fatalf("source=%q", record.SourceText)
	}
}

func TestNormalizeIngestRequestSanitizesMessages(t *testing.T) {
	req := IngestRequest{
		Messages: []Message{{Role: "user", Content: "x\x80y"}},
	}
	NormalizeIngestRequest(&req)
	if req.Messages[0].Content != "xy" {
		t.Fatalf("content=%q", req.Messages[0].Content)
	}
}
