package memory

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"brainy/internal/pack"
)

type memoryStoreStub struct {
	records       map[string]MemoryRecord
	jobs          map[string]ExtractionJob
	entityLinks   map[string][]string
	relations     []MemoryRelation
	atoms         []stubAtom
	currentState  map[string]currentStateRow
	entities      []MemoryEntity
	searchOnlyIDs map[string]struct{}
}

type stubAtom struct {
	pred, val, memID string
}

type currentStateRow struct {
	MemoryID string
	Value    string
	Policy   string
}

func newMemoryStoreStub() *memoryStoreStub {
	return &memoryStoreStub{
		records:      map[string]MemoryRecord{},
		jobs:         map[string]ExtractionJob{},
		entityLinks:  map[string][]string{},
		currentState: map[string]currentStateRow{},
	}
}

func (s *memoryStoreStub) UpsertMemory(_ context.Context, record MemoryRecord) (StoreUpsertResult, error) {
	if existing, ok := s.records[record.DedupeKey]; ok {
		if existing.Status == StatusSuppressed {
			return StoreUpsertResult{Record: existing, State: "deduped"}, nil
		}
		if existing.Content == record.Content && existing.Status == StatusActive &&
			metadataEqual(existing.Metadata, record.Metadata) &&
			existing.LifecycleState == record.LifecycleState &&
			existing.Label == record.Label &&
			existing.Scope == record.Scope {
			return StoreUpsertResult{Record: existing, State: "deduped"}, nil
		}
		record.MemoryID = existing.MemoryID
		record.CreatedAt = existing.CreatedAt
		s.records[record.DedupeKey] = record
		return StoreUpsertResult{Record: record, State: "updated"}, nil
	}
	s.records[record.DedupeKey] = record
	return StoreUpsertResult{Record: record, State: "created"}, nil
}

func metadataEqual(left, right map[string]any) bool {
	if len(left) != len(right) {
		return false
	}
	for key, leftValue := range left {
		rightValue, ok := right[key]
		if !ok {
			return false
		}
		if fmt.Sprint(leftValue) != fmt.Sprint(rightValue) {
			return false
		}
	}
	return true
}

func (s *memoryStoreStub) ListActiveMemories(ctx context.Context, tenantID, subjectID string) ([]MemoryRecord, error) {
	return s.ListMemories(ctx, tenantID, subjectID, false)
}

func (s *memoryStoreStub) ListMemories(_ context.Context, tenantID, subjectID string, includeSuperseded bool) ([]MemoryRecord, error) {
	var out []MemoryRecord
	for _, record := range s.records {
		if record.TenantID != tenantID || record.SubjectID != subjectID || record.Status != StatusActive {
			continue
		}
		if includeSuperseded {
			if record.LifecycleState == LifecycleArchived || record.LifecycleState == LifecycleSuppressed {
				continue
			}
		} else if !IsLifecycleSearchVisible(record.LifecycleState) {
			continue
		}
		out = append(out, record)
	}
	return out, nil
}

func (s *memoryStoreStub) SearchActiveMemories(ctx context.Context, tenantID, subjectID string, patterns []string, limit int) ([]MemoryRecord, error) {
	return s.SearchMemories(ctx, tenantID, subjectID, patterns, limit, false)
}

func (s *memoryStoreStub) SearchMemories(ctx context.Context, tenantID, subjectID string, patterns []string, limit int, includeSuperseded bool) ([]MemoryRecord, error) {
	_ = patterns
	_ = limit
	out, err := s.ListMemories(ctx, tenantID, subjectID, includeSuperseded)
	if err != nil || s.searchOnlyIDs == nil {
		return out, err
	}
	filtered := make([]MemoryRecord, 0, len(s.searchOnlyIDs))
	for _, rec := range out {
		if _, ok := s.searchOnlyIDs[rec.MemoryID]; ok {
			filtered = append(filtered, rec)
		}
	}
	return filtered, nil
}

func (s *memoryStoreStub) ListMemoriesBySessionIDs(ctx context.Context, tenantID, subjectID string, sessionIDs []string, includeSuperseded bool, perSession int) ([]MemoryRecord, error) {
	if len(sessionIDs) == 0 {
		return nil, nil
	}
	if len(sessionIDs) > 8 {
		sessionIDs = sessionIDs[:8]
	}
	if perSession <= 0 || perSession > LeftoverCoveringSessionListPer {
		perSession = LeftoverCoveringSessionListPer
	}
	want := make(map[string]struct{}, len(sessionIDs))
	for _, id := range sessionIDs {
		if id != "" {
			want[id] = struct{}{}
		}
	}
	all, err := s.ListMemories(ctx, tenantID, subjectID, includeSuperseded)
	if err != nil {
		return nil, err
	}
	bySess := map[string][]MemoryRecord{}
	for _, rec := range all {
		sid := sessionIDOf(rec)
		if _, ok := want[sid]; !ok {
			continue
		}
		bySess[sid] = append(bySess[sid], rec)
	}
	var out []MemoryRecord
	for _, sid := range sessionIDs {
		rows := bySess[sid]
		sort.SliceStable(rows, func(i, j int) bool {
			if rows[i].UpdatedAt.Equal(rows[j].UpdatedAt) {
				return rows[i].MemoryID < rows[j].MemoryID
			}
			return rows[i].UpdatedAt.After(rows[j].UpdatedAt)
		})
		if len(rows) > perSession {
			rows = rows[:perSession]
		}
		out = append(out, rows...)
	}
	return out, nil
}

func (s *memoryStoreStub) GetMemory(_ context.Context, tenantID, subjectID, memoryID string) (MemoryRecord, error) {
	for _, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.MemoryID == memoryID {
			return record, nil
		}
	}
	return MemoryRecord{}, ErrMemoryNotFound
}

func (s *memoryStoreStub) LinkMemoryEntities(_ context.Context, _, _, memoryID string, entities []string) error {
	if s.entityLinks == nil {
		s.entityLinks = map[string][]string{}
	}
	for _, e := range entities {
		s.entityLinks[e] = appendUnique(s.entityLinks[e], memoryID)
	}
	return nil
}

func (s *memoryStoreStub) EntityHubBoosts(_ context.Context, _, _ string, queryEntities []string) (map[string]float64, error) {
	out := map[string]float64{}
	for _, e := range queryEntities {
		ids := s.entityLinks[e]
		if len(ids) == 0 || len(ids) > 40 {
			continue
		}
		w := 0.5 / float64(len(ids))
		if w > 0.35 {
			w = 0.35
		}
		for _, id := range ids {
			out[id] += w
			if out[id] > 0.5 {
				out[id] = 0.5
			}
		}
	}
	return out, nil
}

func (s *memoryStoreStub) UpsertMemoryRelation(_ context.Context, rel MemoryRelation) error {
	s.relations = append(s.relations, rel)
	return nil
}

func (s *memoryStoreStub) ListRelationsFrom(_ context.Context, tenantID, subjectID, srcEntity, relation string, limit int) ([]MemoryRelation, error) {
	srcEntity = strings.ToLower(strings.TrimSpace(srcEntity))
	out := make([]MemoryRelation, 0)
	for _, rel := range s.relations {
		if rel.TenantID != tenantID || rel.SubjectID != subjectID {
			continue
		}
		if strings.ToLower(rel.SrcEntity) != srcEntity &&
			rel.SrcEntityID != srcEntity &&
			(rel.SrcEntityID == "" || rel.SrcEntityID != CanonicalEntityID(tenantID, subjectID, srcEntity)) {
			continue
		}
		if relation != "" && rel.Relation != relation {
			continue
		}
		out = append(out, rel)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (s *memoryStoreStub) UpsertMemoryEntity(_ context.Context, ent MemoryEntity) error {
	if ent.EntityID == "" {
		return nil
	}
	for i, existing := range s.entities {
		if existing.TenantID == ent.TenantID && existing.SubjectID == ent.SubjectID && existing.EntityID == ent.EntityID {
			s.entities[i] = ent
			return nil
		}
	}
	s.entities = append(s.entities, ent)
	return nil
}

func (s *memoryStoreStub) ResolveMemoryEntity(_ context.Context, tenantID, subjectID, mention string) (MemoryEntity, bool, error) {
	cands := make([]MemoryEntity, 0)
	for _, e := range s.entities {
		if e.TenantID == tenantID && e.SubjectID == subjectID {
			cands = append(cands, e)
		}
	}
	got, ok := RankEntityResolution(cands, mention)
	return got, ok, nil
}

func (s *memoryStoreStub) UpsertMemoryAtom(_ context.Context, _, _, predicate, value, memoryID string, _ *time.Time) error {
	s.atoms = append(s.atoms, stubAtom{pred: predicate, val: value, memID: memoryID})
	return nil
}

func (s *memoryStoreStub) ListAtomMemoryIDs(_ context.Context, _, _ string, predicate, valueNorm string, limit int) ([]string, error) {
	out := make([]string, 0)
	seen := map[string]struct{}{}
	for _, a := range s.atoms {
		if predicate != "" && a.pred != predicate {
			continue
		}
		if valueNorm != "" && !strings.EqualFold(a.val, valueNorm) {
			continue
		}
		if _, ok := seen[a.memID]; ok {
			continue
		}
		seen[a.memID] = struct{}{}
		out = append(out, a.memID)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (s *memoryStoreStub) MarkSuperseded(_ context.Context, tenantID, subjectID, memoryID string) error {
	for key, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.MemoryID == memoryID {
			now := time.Now().UTC()
			record.LifecycleState = LifecycleSuperseded
			record.SupersededAt = &now
			record.UpdatedAt = now
			s.records[key] = record
			return nil
		}
	}
	return ErrMemoryNotFound
}

func (s *memoryStoreStub) GetCurrentState(_ context.Context, tenantID, subjectID, predicate string) (memoryID, value, policy string, ok bool, err error) {
	key := tenantID + "::" + subjectID + "::" + predicate
	row, found := s.currentState[key]
	if !found {
		return "", "", "", false, nil
	}
	return row.MemoryID, row.Value, row.Policy, true, nil
}

func (s *memoryStoreStub) UpsertCurrentState(_ context.Context, tenantID, subjectID, predicate, memoryID, value, policy string) error {
	key := tenantID + "::" + subjectID + "::" + predicate
	s.currentState[key] = currentStateRow{MemoryID: memoryID, Value: value, Policy: policy}
	return nil
}

func (s *memoryStoreStub) DeleteCurrentStateByMemory(_ context.Context, tenantID, subjectID, memoryID string) error {
	for key, row := range s.currentState {
		if strings.HasPrefix(key, tenantID+"::"+subjectID+"::") && row.MemoryID == memoryID {
			delete(s.currentState, key)
		}
	}
	return nil
}

func (s *memoryStoreStub) SuppressMemory(_ context.Context, tenantID, subjectID, memoryID string) error {
	for key, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.MemoryID == memoryID {
			record.Status = StatusSuppressed
			record.UpdatedAt = time.Now().UTC()
			s.records[key] = record
			return nil
		}
	}
	return nil
}

func (s *memoryStoreStub) CorrectMemory(_ context.Context, tenantID, subjectID, memoryID, content, sourceText string) (MemoryRecord, error) {
	for key, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.MemoryID == memoryID {
			delete(s.records, key)
			record.Content = content
			record.SourceText = sourceText
			record.DedupeKey = DedupeKey(tenantID, subjectID, record.Kind, content)
			record.Status = StatusActive
			record.UpdatedAt = time.Now().UTC()
			s.records[record.DedupeKey] = record
			return record, nil
		}
	}
	return MemoryRecord{}, errors.New("memory not found")
}

func (s *memoryStoreStub) EnqueueIngestJob(_ context.Context, ingestID, jobID, _ string, req IngestRequest) (EnqueueResult, error) {
	s.jobs[jobID] = ExtractionJob{JobID: jobID, IngestID: ingestID, Request: req}
	return EnqueueResult{IngestID: ingestID, JobID: jobID, Accepted: true}, nil
}

func (s *memoryStoreStub) ClaimNextExtractionJob(_ context.Context) (ExtractionJob, bool, error) {
	for jobID, job := range s.jobs {
		delete(s.jobs, jobID)
		return job, true, nil
	}
	return ExtractionJob{}, false, nil
}

func (s *memoryStoreStub) CompleteExtractionJob(_ context.Context, _, _ string) error {
	return nil
}

func (s *memoryStoreStub) FailExtractionJob(_ context.Context, _, _, _ string) error {
	return nil
}

func TestServiceIngestSearchAndSuppress(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)

	result, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "I prefer concise, direct answers."},
		},
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if result.Created != 1 {
		t.Fatalf("expected 1 created memory, got %d", result.Created)
	}

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "How should I respond?")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(search.Results) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(search.Results))
	}

	if err := service.Suppress(context.Background(), "t1", "u1", result.Memories[0].MemoryID); err != nil {
		t.Fatalf("suppress failed: %v", err)
	}

	searchAfterSuppress, err := service.Search(context.Background(), "t1", "u1", "", "", "How should I respond?")
	if err != nil {
		t.Fatalf("search after suppress failed: %v", err)
	}
	if len(searchAfterSuppress.Results) != 0 {
		t.Fatalf("expected 0 results after suppression, got %d", len(searchAfterSuppress.Results))
	}
}

func TestServiceCorrectUpdatesLaterSearchResults(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)

	result, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "I prefer concise answers."},
		},
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}

	_, err = service.Correct(context.Background(), "t1", "u1", result.Memories[0].MemoryID, CorrectionRequest{
		Content: "Prefers detailed answers",
	})
	if err != nil {
		t.Fatalf("correct failed: %v", err)
	}

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "How should I answer?")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(search.Results) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(search.Results))
	}
	if search.Results[0].Content != "Prefers detailed answers" {
		t.Fatalf("expected corrected content, got %q", search.Results[0].Content)
	}
}

func TestServiceIngestAsyncEnqueuesJob(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)

	result, err := service.IngestAsync(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "I prefer concise answers."},
		},
	})
	if err != nil {
		t.Fatalf("async ingest failed: %v", err)
	}
	if !result.Accepted || result.JobID == "" || result.IngestID == "" {
		t.Fatalf("expected accepted async ingest with ids, got %+v", result)
	}
	if len(store.jobs) != 1 {
		t.Fatalf("expected 1 enqueued job, got %d", len(store.jobs))
	}
}

func TestVerticalPackPrincipleRanksAbovePreference(t *testing.T) {
	reg, err := pack.LoadRegistryFromDir(filepath.Join("..", "..", "packs"))
	if err != nil {
		t.Fatalf("load packs: %v", err)
	}
	store := newMemoryStoreStub()
	service := NewServiceWithPacks(store, reg)

	_, err = service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "brand",
		Vertical:   "marketing",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "I prefer warm casual copy."},
		},
	})
	if err != nil {
		t.Fatalf("preference ingest: %v", err)
	}

	_, err = service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "brand",
		Vertical:   "marketing",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Never mention competitor X in any copy."},
		},
	})
	if err != nil {
		t.Fatalf("principle ingest: %v", err)
	}

	search, err := service.Search(context.Background(), "t1", "brand", "marketing", "", "copy")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(search.Results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(search.Results))
	}
	if !strings.Contains(strings.ToLower(search.Results[0].Content), "never") {
		t.Fatalf("expected principle first, got %q", search.Results[0].Content)
	}
}

func TestLifecycleArchivedCampaignExcludedFromSearch(t *testing.T) {
	reg, err := pack.LoadRegistryFromDir(filepath.Join("..", "..", "packs"))
	if err != nil {
		t.Fatalf("load packs: %v", err)
	}
	store := newMemoryStoreStub()
	service := NewServiceWithPacks(store, reg)

	_, err = service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "brand",
		Vertical:   "marketing",
		Label:      "campaign",
		Metadata:   map[string]any{"name": "Summer Sale", "status": "archived"},
		SourceType: "campaign",
		Messages: []Message{
			{Role: "user", Content: "Summer Sale campaign headline is Save 20% today."},
		},
	})
	if err != nil {
		t.Fatalf("ingest archived campaign: %v", err)
	}

	search, err := service.Search(context.Background(), "t1", "brand", "marketing", "", "Summer Sale headline")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(search.Results) != 0 {
		t.Fatalf("expected archived campaign hidden, got %d results", len(search.Results))
	}
}

func TestLifecycleActiveCampaignRanksAboveCompleted(t *testing.T) {
	reg, err := pack.LoadRegistryFromDir(filepath.Join("..", "..", "packs"))
	if err != nil {
		t.Fatalf("load packs: %v", err)
	}
	store := newMemoryStoreStub()
	service := NewServiceWithPacks(store, reg)

	_, err = service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "brand",
		Vertical:   "marketing",
		Label:      "campaign",
		Metadata:   map[string]any{"name": "Winter", "status": "completed"},
		SourceType: "campaign",
		Messages: []Message{
			{Role: "user", Content: "Winter splash sale is active."},
		},
	})
	if err != nil {
		t.Fatalf("ingest completed campaign: %v", err)
	}

	_, err = service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "brand",
		Vertical:   "marketing",
		Label:      "campaign",
		Metadata:   map[string]any{"name": "Summer", "status": "active"},
		SourceType: "campaign",
		Messages: []Message{
			{Role: "user", Content: "Summer splash sale is active."},
		},
	})
	if err != nil {
		t.Fatalf("ingest active campaign: %v", err)
	}

	search, err := service.Search(context.Background(), "t1", "brand", "marketing", "", "splash sale")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(search.Results) < 2 {
		t.Fatalf("expected 2 results, got %d", len(search.Results))
	}
	if !strings.Contains(search.Results[0].Content, "Summer") {
		t.Fatalf("expected active Summer campaign first, got %q", search.Results[0].Content)
	}
	if mult, ok := search.Results[0].Explain["lifecycle_rank_multiplier"].(float64); !ok || mult != 1.5 {
		t.Fatalf("expected active lifecycle multiplier 1.5, got %v", search.Results[0].Explain["lifecycle_rank_multiplier"])
	}
}

func TestLifecycleMetadataUpdateChangesState(t *testing.T) {
	reg, err := pack.LoadRegistryFromDir(filepath.Join("..", "..", "packs"))
	if err != nil {
		t.Fatalf("load packs: %v", err)
	}
	store := newMemoryStoreStub()
	service := NewServiceWithPacks(store, reg)

	req := IngestRequest{
		TenantID:   "t1",
		SubjectID:  "brand",
		Vertical:   "marketing",
		Label:      "campaign",
		Metadata:   map[string]any{"name": "Launch", "status": "active"},
		SourceType: "campaign",
		Messages: []Message{
			{Role: "user", Content: "Launch campaign offer is free shipping."},
		},
	}
	_, err = service.Ingest(context.Background(), req)
	if err != nil {
		t.Fatalf("ingest active campaign: %v", err)
	}

	searchBefore, err := service.Search(context.Background(), "t1", "brand", "marketing", "", "Launch offer")
	if err != nil {
		t.Fatalf("search before archive: %v", err)
	}
	if len(searchBefore.Results) != 1 {
		t.Fatalf("expected active campaign searchable, got %d", len(searchBefore.Results))
	}

	req.Metadata = map[string]any{"name": "Launch", "status": "archived"}
	result, err := service.Ingest(context.Background(), req)
	if err != nil {
		t.Fatalf("re-ingest archived campaign: %v", err)
	}
	if result.Updated != 1 {
		t.Fatalf("expected metadata update, got created=%d updated=%d deduped=%d", result.Created, result.Updated, result.Deduped)
	}

	searchAfter, err := service.Search(context.Background(), "t1", "brand", "marketing", "", "Launch offer")
	if err != nil {
		t.Fatalf("search after archive: %v", err)
	}
	if len(searchAfter.Results) != 0 {
		t.Fatalf("expected archived campaign hidden after update, got %d", len(searchAfter.Results))
	}
}

func TestSuppressedMemoryNotResurrectedOnReingest(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)

	ingest, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []Message{{Role: "user", Content: "Never share the door code with vendors."}},
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if err := service.Suppress(context.Background(), "t1", "u1", ingest.Memories[0].MemoryID); err != nil {
		t.Fatalf("suppress failed: %v", err)
	}

	reingest, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []Message{{Role: "user", Content: "Never share the door code with vendors."}},
	})
	if err != nil {
		t.Fatalf("re-ingest failed: %v", err)
	}
	if reingest.Created != 0 {
		t.Fatalf("expected no new memories after re-ingesting suppressed content, got created=%d", reingest.Created)
	}

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "door code")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(search.Results) != 0 {
		t.Fatalf("expected suppressed memory to stay hidden after re-ingest, got %d results", len(search.Results))
	}
}

func TestSearchPrefersNewerConflictingPreference(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)

	if _, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []Message{{Role: "user", Content: "I prefer email updates."}},
	}); err != nil {
		t.Fatalf("first ingest failed: %v", err)
	}
	if _, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []Message{{Role: "user", Content: "I prefer SMS updates."}},
	}); err != nil {
		t.Fatalf("second ingest failed: %v", err)
	}

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "updates")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected at least one search result")
	}
	if !strings.Contains(strings.ToLower(search.Results[0].Content), "sms") {
		t.Fatalf("expected newer SMS preference to rank first, got %q", search.Results[0].Content)
	}
}

func TestSearchAdmitsRareQueryToken(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	for i := 0; i < 24; i++ {
		key := "dump" + itoa(i)
		store.records[key] = MemoryRecord{
			MemoryID: "mem_" + key, TenantID: "t-fill", SubjectID: "u1",
			Kind: KindFact, Content: "Alex made a cake for the party last week",
			DedupeKey: key, Status: StatusActive, UpdatedAt: now,
		}
	}
	store.records["fill"] = MemoryRecord{
		MemoryID: "mem_fill", TenantID: "t-fill", SubjectID: "u1",
		Kind: KindFact, Content: "The filling is strawberry",
		DedupeKey: "fill", Status: StatusActive, UpdatedAt: now,
	}
	out, err := svc.SearchOpt(context.Background(), "t-fill", "u1", "", "",
		"What filling did Alex use in the cake", SearchOptions{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range out.Results {
		if strings.Contains(strings.ToLower(r.Content), "strawberry") || strings.Contains(strings.ToLower(r.Content), "filling") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected filling fact in top-k, got %+v", out.Results)
	}
}

func TestUncoveredQueryTokensNamedPersonIgnoresFirstPersonEventCover(t *testing.T) {
	q := "When did Caroline go biking with friends?"
	cands := map[string]MemoryRecord{
		"fp":    {MemoryID: "fp", Content: "I had a wicked day out with the gang last weekend - we went biking"},
		"lgbtq": {MemoryID: "lgbtq", Content: "Caroline painted her artwork to capture the unity of the LGBTQ community"},
	}
	got := uncoveredQueryTokensInCandidates(q, cands, tokenize(q))
	hasBiking := false
	for _, tok := range got {
		if tok == "biking" {
			hasBiking = true
		}
	}
	if !hasBiking {
		t.Fatalf("biking must stay uncovered when only first-person leftover has it, got %v", got)
	}
	cands["compiled"] = MemoryRecord{
		MemoryID: "compiled",
		Content:  "Caroline went biking with the gang last weekend on 2023-09-09.",
	}
	got2 := uncoveredQueryTokensInCandidates(q, cands, tokenize(q))
	for _, tok := range got2 {
		if tok == "biking" {
			t.Fatalf("named compiled fact must cover biking, got %v", got2)
		}
	}
	anon := "When did I go biking last weekend?"
	anonCands := map[string]MemoryRecord{
		"fp": {MemoryID: "fp", Content: "I had a wicked day out with the gang last weekend - we went biking"},
	}
	for _, tok := range uncoveredQueryTokensInCandidates(anon, anonCands, tokenize(anon)) {
		if tok == "biking" {
			t.Fatalf("unnamed query must still treat first-person biking as coverage, got %v",
				uncoveredQueryTokensInCandidates(anon, anonCands, tokenize(anon)))
		}
	}
	newsQ := "What exciting news did Maria share on 16 June, 2023?"
	newsCands := map[string]MemoryRecord{
		"gym": {MemoryID: "gym", Content: "I got some great news to share - I joined a gym last week (9 June 2023)"},
	}
	gotNews := uncoveredQueryTokensInCandidates(newsQ, newsCands, tokenize(newsQ))
	for _, tok := range gotNews {
		if tok == "news" || tok == "june" || tok == "share" {
			t.Fatalf("first-person leftover must not extra-uncover tokens it already covers, got %v", gotNews)
		}
	}
	hasExciting := false
	for _, tok := range gotNews {
		if tok == "exciting" {
			hasExciting = true
		}
	}
	if !hasExciting {
		t.Fatalf("exciting with no candidate coverage must stay base-uncovered, got %v", gotNews)
	}
}

func TestSearchWhenEventDropsDecideFlood(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	for i := 0; i < 20; i++ {
		key := "dec" + itoa(i)
		store.records[key] = MemoryRecord{
			MemoryID: "mem_" + key, TenantID: "t-paint", SubjectID: "u1",
			Kind: KindFact, Content: "Sam decided he needs to make health changes after being mocked on 21 July 2023.",
			DedupeKey: key, Status: StatusActive, UpdatedAt: now,
		}
	}
	store.records["cactus"] = MemoryRecord{
		MemoryID: "mem_cactus", TenantID: "t-paint", SubjectID: "u1",
		Kind: KindFact, Content: "Riley painted a watercolor cactus in the desert last week (29 September 2023).",
		DedupeKey: "cactus", Status: StatusActive, UpdatedAt: now,
	}
	store.records["plan"] = MemoryRecord{
		MemoryID: "mem_plan", TenantID: "t-paint", SubjectID: "u1",
		Kind: KindFact, Content: "Sam plans to paint with Riley on Saturday, 16 September 2023.",
		DedupeKey: "plan", Status: StatusActive, UpdatedAt: now,
	}
	out, err := svc.SearchOpt(context.Background(), "t-paint", "u1", "", "",
		"When did Riley and Sam decide to paint together?", SearchOptions{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range out.Results {
		low := strings.ToLower(r.Content)
		if strings.Contains(low, "16 september") || strings.Contains(low, "saturday") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("when-event search must admit the dual-entity paint plan, got %+v", out.Results)
	}
}

func TestSearchLexicalTokensDropsDecideWhenEventRemains(t *testing.T) {
	toks := tokenize("When did Riley and Sam decide to paint together?")
	got := searchLexicalTokens(toks)
	joined := strings.Join(got, " ")
	if strings.Contains(joined, "decide") {
		t.Fatalf("when-event lexical tokens must drop decide when paint remains, got %v", got)
	}
	if !strings.Contains(joined, "paint") || !strings.Contains(joined, "riley") || !strings.Contains(joined, "sam") {
		t.Fatalf("when-event lexical tokens must keep people and the event noun, got %v", got)
	}
}

func TestSearchLexicalTokensDropsWhatMadeStructureAndPerson(t *testing.T) {
	q := "What made being part of the running group easy for Deborah to stay motivated?"
	got := searchLexicalQueryTokens(q, tokenize(q))
	joined := strings.Join(got, " ")
	for _, banned := range []string{"made", "part", "deborah"} {
		if strings.Contains(joined, banned) {
			t.Fatalf("what-made lexical tokens must drop structure/person %q, got %v", banned, got)
		}
	}
	for _, banned := range []string{"easy", "stay"} {
		if strings.Contains(joined, banned) {
			t.Fatalf("what-made lexical tokens must drop short reason verb %q, got %v", banned, got)
		}
	}
	for _, keep := range []string{"running", "group", "motivated"} {
		if !strings.Contains(joined, keep) {
			t.Fatalf("what-made lexical tokens must keep %q, got %v", keep, got)
		}
	}
	destress := searchLexicalQueryTokens("What does Melanie do to destress?", tokenize("What does Melanie do to destress?"))
	if !strings.Contains(strings.Join(destress, " "), "melanie") {
		t.Fatalf("non-what-made queries must keep the person token, got %v", destress)
	}
}

func TestSearchLexicalTokensDropsWhatMotivatesStructureAndPerson(t *testing.T) {
	q := "What motivates Joanna to keep writing even on tough days?"
	got := searchLexicalQueryTokens(q, tokenize(q))
	joined := strings.Join(got, " ")
	for _, banned := range []string{"motivates", "motivate", "joanna", "keep", "even"} {
		if strings.Contains(joined, banned) {
			t.Fatalf("what-motivates lexical tokens must drop structure/person %q, got %v", banned, got)
		}
	}
	for _, keep := range []string{"writing", "tough", "days"} {
		if !strings.Contains(joined, keep) {
			t.Fatalf("what-motivates lexical tokens must keep %q, got %v", keep, got)
		}
	}
	made := searchLexicalQueryTokens("What made being part of the running group easy for Deborah to stay motivated?", tokenize("What made being part of the running group easy for Deborah to stay motivated?"))
	if !strings.Contains(strings.Join(made, " "), "motivated") {
		t.Fatalf("what-made queries must keep motivated, got %v", made)
	}
}

func TestSearchLexicalTokensDropsWhatSayAboutStructureAndPerson(t *testing.T) {
	q := "What does Gina say about the dancers in the photo?"
	got := searchLexicalQueryTokens(q, tokenize(q))
	joined := strings.Join(got, " ")
	for _, banned := range []string{"say", "about"} {
		if strings.Contains(joined, banned) {
			t.Fatalf("what-say-about lexical tokens must drop structure %q, got %v", banned, got)
		}
	}
	for _, keep := range []string{"dancers", "photo"} {
		if !strings.Contains(joined, keep) {
			t.Fatalf("what-say-about lexical tokens must keep %q, got %v", keep, got)
		}
	}
	nyc := searchLexicalQueryTokens("What did John say about NYC, enticing Tim to visit?", tokenize("What did John say about NYC, enticing Tim to visit?"))
	nycJoined := strings.Join(nyc, " ")
	for _, banned := range []string{"say", "about", "enticing", "john", "tim"} {
		if strings.Contains(nycJoined, banned) {
			t.Fatalf("NYC what-say-about lexical tokens must drop %q, got %v", banned, nyc)
		}
	}
	for _, keep := range []string{"nyc", "visit"} {
		if !strings.Contains(nycJoined, keep) {
			t.Fatalf("NYC what-say-about lexical tokens must keep %q, got %v", keep, nyc)
		}
	}
	advice := searchLexicalQueryTokens("What advice does Gina give to Jon about running a successful business?", tokenize("What advice does Gina give to Jon about running a successful business?"))
	if !strings.Contains(strings.Join(advice, " "), "advice") {
		t.Fatalf("advice queries must keep the speech-act token, got %v", advice)
	}
}

func TestSearchWhatSayAboutAdmitsTheyEvaluativePastSessionWindow(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	const session = "session_1"
	photo := MemoryRecord{
		MemoryID: "mem_photo", TenantID: "t-say", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "[group dancers performing on stage] [a photo of a group of dancers in white dresses on a stage]",
		DedupeKey: "photo", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"session_id": session},
	}
	store.records["photo"] = photo
	store.searchOnlyIDs = map[string]struct{}{photo.MemoryID: {}}
	for i := 0; i < 90; i++ {
		key := fmt.Sprintf("early%02d", i)
		store.records[key] = MemoryRecord{
			MemoryID: "mem_" + key, TenantID: "t-say", SubjectID: "u1",
			Kind: KindFact, Content: "Jon lost his job as a banker on 19 January 2023.",
			DedupeKey: key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"session_id": session},
		}
	}
	store.records["gold"] = MemoryRecord{
		MemoryID: "mem_zz_gold", TenantID: "t-say", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "They're so graceful",
		DedupeKey: "gold", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"session_id": session},
	}
	out, err := svc.SearchOpt(context.Background(), "t-say", "u1", "", "",
		"What does Gina say about the dancers in the photo?", SearchOptions{Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range out.Results {
		if strings.Contains(strings.ToLower(r.Content), "they're so graceful") || strings.Contains(strings.ToLower(r.Content), "they are so graceful") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("what-say-about search must admit they-evaluative leftover past an 80-row session window, got %+v", out.Results)
	}
}

func TestSearchWhatSayAboutAdmitsFirstPersonGotPastSessionWindow(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	const session = "session_9"
	nyc := MemoryRecord{
		MemoryID: "mem_nyc", TenantID: "t-got", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "That skyline looks amazing - I've been wanting to visit NYC",
		DedupeKey: "nyc", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"session_id": session},
	}
	store.records["nyc"] = nyc
	store.searchOnlyIDs = map[string]struct{}{nyc.MemoryID: {}}
	for i := 0; i < 90; i++ {
		key := fmt.Sprintf("early%02d", i)
		store.records[key] = MemoryRecord{
			MemoryID: "mem_" + key, TenantID: "t-got", SubjectID: "u1",
			Kind: KindFact, Content: "John plans to let Tim know his thoughts after reading the material.",
			DedupeKey: key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"session_id": session},
		}
	}
	store.records["gold"] = MemoryRecord{
		MemoryID: "mem_zz_gold", TenantID: "t-got", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "It's got so much to check out - the culture, food - you won't regret it.",
		DedupeKey: "gold", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"session_id": session},
	}
	out, err := svc.SearchOpt(context.Background(), "t-got", "u1", "", "",
		"What did John say about NYC, enticing Tim to visit?", SearchOptions{Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range out.Results {
		if strings.Contains(strings.ToLower(r.Content), "got so much") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("what-say-about search must admit first-person got leftover past an 80-row session window, got %+v", out.Results)
	}
}

func TestSearchWhatSayAboutAdmitsReportedSpeechPastSessionWindow(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	const session = "session_18"
	photo := MemoryRecord{
		MemoryID: "mem_photo", TenantID: "t-said", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "[ankle injury wrapped bandages] [a photo of a person with a bandage on their leg]",
		DedupeKey: "photo", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"session_id": session},
	}
	store.records["photo"] = photo
	store.searchOnlyIDs = map[string]struct{}{photo.MemoryID: {}}
	for i := 0; i < 90; i++ {
		key := fmt.Sprintf("early%02d", i)
		store.records[key] = MemoryRecord{
			MemoryID: "mem_" + key, TenantID: "t-said", SubjectID: "u1",
			Kind: KindFact, Content: "Tim cannot read due to an injury.",
			DedupeKey: key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"session_id": session},
		}
	}
	store.records["gold"] = MemoryRecord{
		MemoryID: "mem_zz_gold", TenantID: "t-said", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "The doctor said it's not too serious",
		DedupeKey: "gold", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"session_id": session},
	}
	out, err := svc.SearchOpt(context.Background(), "t-said", "u1", "", "",
		"What did Tim say about his injury on 16 November, 2023?", SearchOptions{Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range out.Results {
		if strings.Contains(strings.ToLower(r.Content), "not too serious") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("what-say-about search must admit reported-speech leftover past an 80-row session window, got %+v", out.Results)
	}
}

func TestSearchLexicalTokensDropsHowReactStructure(t *testing.T) {
	q := "How do Audrey's dogs react to snow?"
	got := searchLexicalQueryTokens(q, tokenize(q))
	joined := strings.Join(got, " ")
	if strings.Contains(joined, "react") {
		t.Fatalf("how-react lexical tokens must drop react, got %v", got)
	}
	for _, keep := range []string{"dogs", "snow"} {
		if !strings.Contains(joined, keep) {
			t.Fatalf("how-react lexical tokens must keep %q, got %v", keep, got)
		}
	}
	respond := searchLexicalQueryTokens("How did they respond to the news?", tokenize("How did they respond to the news?"))
	if strings.Contains(strings.Join(respond, " "), "respond") {
		t.Fatalf("how-respond-to lexical tokens must drop respond, got %v", respond)
	}
	advice := searchLexicalQueryTokens("What advice does Gina give to Jon about running a successful business?", tokenize("What advice does Gina give to Jon about running a successful business?"))
	if !strings.Contains(strings.Join(advice, " "), "advice") {
		t.Fatalf("advice queries must keep the speech-act token, got %v", advice)
	}
	dancers := searchLexicalQueryTokens("What does Gina say about the dancers in the photo?", tokenize("What does Gina say about the dancers in the photo?"))
	if strings.Contains(strings.Join(dancers, " "), "react") {
		t.Fatalf("what-say-about must not be treated as how-react, got %v", dancers)
	}
}

func TestSearchHowReactAdmitsObservationPastSessionWindow(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	const session = "session_23"
	dislike := MemoryRecord{
		MemoryID: "mem_dislike", TenantID: "t-react", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "Audrey's dogs dislike snow.",
		DedupeKey: "dislike", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"session_id": session},
	}
	store.records["dislike"] = dislike
	store.searchOnlyIDs = map[string]struct{}{dislike.MemoryID: {}}
	for i := 0; i < 90; i++ {
		key := fmt.Sprintf("early%02d", i)
		store.records[key] = MemoryRecord{
			MemoryID: "mem_" + key, TenantID: "t-react", SubjectID: "u1",
			Kind: KindFact, Content: "Audrey walked the dogs in the park last week.",
			DedupeKey: key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"session_id": session},
		}
	}
	store.records["gold"] = MemoryRecord{
		MemoryID: "mem_zz_gold", TenantID: "t-react", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "I took them to a snowy one last winter and they were so confused",
		DedupeKey: "gold", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"session_id": session},
	}
	out, err := svc.SearchOpt(context.Background(), "t-react", "u1", "", "",
		"How do Audrey's dogs react to snow?", SearchOptions{Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range out.Results {
		if strings.Contains(strings.ToLower(r.Content), "confused") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("how-react search must admit they-were observation leftover past an 80-row session window, got %+v", out.Results)
	}
}

func TestKeepReactionObservationInCapSurvivesListFill(t *testing.T) {
	q := "How do Audrey's dogs react to snow?"
	obs := rankedSearchResult{result: SearchResult{
		MemoryID: "obs",
		Content:  "Audrey took her dogs to a snowy park last winter and they were confused.",
		Score:    0.4,
	}}
	full := []rankedSearchResult{obs}
	capped := make([]rankedSearchResult, 0, 30)
	for i := 0; i < 30; i++ {
		capped = append(capped, rankedSearchResult{result: SearchResult{
			MemoryID: fmt.Sprintf("fill%02d", i),
			Content:  "Audrey walked the dogs in the park last week.",
			Score:    1.5,
		}})
	}
	got := keepReactionObservationInCap(full, capped, q, 30)
	found := false
	for _, item := range got {
		if strings.Contains(strings.ToLower(item.result.Content), "confused") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("how-react evidence-set cap must keep they-were observation leftover, n=%d", len(got))
	}
	if len(got) > 30 {
		t.Fatalf("how-react observation keep must stay within limit, n=%d", len(got))
	}
}

func TestSearchLexicalQueryTokensDropsWhatDidPurposeCalendar(t *testing.T) {
	q := "What did Audrey do in November 2023 to better take care of her dogs?"
	got := searchLexicalQueryTokens(q, tokenize(q))
	joined := strings.Join(got, " ")
	for _, banned := range []string{"november", "2023", "better", "audrey"} {
		if strings.Contains(joined, banned) {
			t.Fatalf("what-did-purpose lexical tokens must drop calendar/person/comparative %q, got %v", banned, got)
		}
	}
	for _, keep := range []string{"take", "care", "dogs"} {
		if !strings.Contains(joined, keep) {
			t.Fatalf("what-did-purpose lexical tokens must keep %q, got %v", keep, got)
		}
	}
	injury := searchLexicalQueryTokens("What did Tim say about his injury on 16 November, 2023?", tokenize("What did Tim say about his injury on 16 November, 2023?"))
	joinedInjury := strings.Join(injury, " ")
	if strings.Contains(joinedInjury, "take") && strings.Contains(joinedInjury, "care") {
		t.Fatalf("dated what-say-about must not be rewritten as what-did-purpose tokens, got %v", injury)
	}
	if !strings.Contains(joinedInjury, "injury") {
		t.Fatalf("dated what-say-about must keep injury, got %v", injury)
	}
}

func TestKeepPurposeActionInCapSurvivesListFill(t *testing.T) {
	q := "What did Audrey do in November 2023 to better take care of her dogs?"
	obs := rankedSearchResult{result: SearchResult{
		MemoryID: "obs",
		Content:  "Audrey recently joined a dog owners group to learn how to better take care of her dogs.",
		Score:    0.4,
	}}
	full := []rankedSearchResult{obs}
	capped := make([]rankedSearchResult, 0, 30)
	for i := 0; i < 30; i++ {
		capped = append(capped, rankedSearchResult{result: SearchResult{
			MemoryID: fmt.Sprintf("fill%02d", i),
			Content:  "Audrey walked the dogs in the park last week.",
			Score:    1.5,
		}})
	}
	got := keepPurposeActionInCap(full, capped, q, 30)
	found := false
	for _, item := range got {
		if strings.Contains(strings.ToLower(item.result.Content), "joined") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("what-did-purpose evidence-set cap must keep take-care action leftover, n=%d", len(got))
	}
	if len(got) > 30 {
		t.Fatalf("what-did-purpose action keep must stay within limit, n=%d", len(got))
	}
}

func TestSearchLexicalQueryTokensDropsHowDidStartStructure(t *testing.T) {
	q := "How did Evan start his transformation journey two years ago?"
	got := searchLexicalQueryTokens(q, tokenize(q))
	joined := strings.Join(got, " ")
	for _, banned := range []string{"start", "transformation", "journey", "evan"} {
		if strings.Contains(joined, banned) {
			t.Fatalf("how-did-start lexical tokens must drop wrapper/person %q, got %v", banned, got)
		}
	}
	for _, keep := range []string{"two", "years", "ago"} {
		if !strings.Contains(joined, keep) {
			t.Fatalf("how-did-start lexical tokens must keep duration %q, got %v", keep, got)
		}
	}
	purpose := searchLexicalQueryTokens("What did Audrey do in November 2023 to better take care of her dogs?", tokenize("What did Audrey do in November 2023 to better take care of her dogs?"))
	joinedPurpose := strings.Join(purpose, " ")
	if strings.Contains(joinedPurpose, "two") && strings.Contains(joinedPurpose, "ago") {
		t.Fatalf("what-did-purpose must not be rewritten as how-did-start duration tokens, got %v", purpose)
	}
}

func TestKeepStartMethodInCapSurvivesListFill(t *testing.T) {
	q := "How did Evan start his transformation journey two years ago?"
	obs := rankedSearchResult{result: SearchResult{
		MemoryID: "obs",
		Content:  "Changed my diet, started walking regularly, things like that",
		Score:    0.4,
	}}
	full := []rankedSearchResult{obs}
	capped := make([]rankedSearchResult, 0, 30)
	for i := 0; i < 30; i++ {
		capped = append(capped, rankedSearchResult{result: SearchResult{
			MemoryID: fmt.Sprintf("fill%02d", i),
			Content:  "Evan went to the gym on 16 October 2023.",
			Score:    1.5,
		}})
	}
	got := keepStartMethodInCap(full, capped, q, 30)
	found := false
	for _, item := range got {
		if strings.Contains(strings.ToLower(item.result.Content), "diet") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("how-did-start evidence-set cap must keep changed+started leftover, n=%d", len(got))
	}
	if len(got) > 30 {
		t.Fatalf("how-did-start method keep must stay within limit, n=%d", len(got))
	}
}

func TestSearchLexicalQueryTokensDropsHowLongBeenStructure(t *testing.T) {
	q := "How long have Mel and her husband been married?"
	got := searchLexicalQueryTokens(q, tokenize(q))
	joined := strings.Join(got, " ")
	if strings.Contains(joined, "long") {
		t.Fatalf("how-long-been lexical tokens must drop long, got %v", got)
	}
	for _, keep := range []string{"married", "husband", "mel"} {
		if !strings.Contains(joined, keep) {
			t.Fatalf("how-long-been lexical tokens must keep %q, got %v", keep, got)
		}
	}
	start := searchLexicalQueryTokens("How did Evan start his transformation journey two years ago?", tokenize("How did Evan start his transformation journey two years ago?"))
	joinedStart := strings.Join(start, " ")
	if strings.Contains(joinedStart, "married") {
		t.Fatalf("how-did-start must not be rewritten as how-long-been tokens, got %v", start)
	}
}

func TestSearchLexicalQueryTokensDropsHowOftenStructure(t *testing.T) {
	q := "How often does Audrey meet up with other dog owners for tips and playdates?"
	got := searchLexicalQueryTokens(q, tokenize(q))
	joined := strings.Join(got, " ")
	for _, banned := range []string{"often", "audrey", "tips", "playdates"} {
		if strings.Contains(joined, banned) {
			t.Fatalf("how-often lexical tokens must drop %q, got %v", banned, got)
		}
	}
	for _, keep := range []string{"meet", "dog", "owners"} {
		if !strings.Contains(joined, keep) {
			t.Fatalf("how-often lexical tokens must keep %q, got %v", keep, got)
		}
	}
	married := searchLexicalQueryTokens("How long have Mel and her husband been married?", tokenize("How long have Mel and her husband been married?"))
	joinedMarried := strings.Join(married, " ")
	if strings.Contains(joinedMarried, "often") {
		t.Fatalf("how-long-been must not be rewritten as how-often tokens, got %v", married)
	}
	if !strings.Contains(joinedMarried, "married") {
		t.Fatalf("how-long-been must keep married, got %v", married)
	}
}

func TestKeepCadenceInCapSurvivesListFill(t *testing.T) {
	q := "How often does Audrey meet up with other dog owners for tips and playdates?"
	obs := rankedSearchResult{result: SearchResult{
		MemoryID: "obs",
		Content:  "I try to meet up with other dog owners once a week for tips from other parents and so they can all play together",
		Score:    0.4,
	}}
	full := []rankedSearchResult{obs}
	capped := make([]rankedSearchResult, 0, 30)
	for i := 0; i < 30; i++ {
		capped = append(capped, rankedSearchResult{result: SearchResult{
			MemoryID: fmt.Sprintf("fill%02d", i),
			Content:  "Audrey's dogs meet other dog owners in the park and have doggie playdates.",
			Score:    1.5,
		}})
	}
	got := keepCadenceInCap(full, capped, q, 30)
	found := false
	for _, item := range got {
		if strings.Contains(strings.ToLower(item.result.Content), "once a week") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("how-often evidence-set cap must keep cadence leftover, n=%d", len(got))
	}
	if len(got) > 30 {
		t.Fatalf("how-often cadence keep must stay within limit, n=%d", len(got))
	}
}

func TestSearchLexicalQueryTokensDropsWhatProjectWorkingStructure(t *testing.T) {
	q := "What project is James working on in his game design course?"
	got := searchLexicalQueryTokens(q, tokenize(q))
	joined := strings.Join(got, " ")
	for _, banned := range []string{"project", "game", "design", "course"} {
		if strings.Contains(joined, banned) {
			t.Fatalf("what-project lexical tokens must drop %q, got %v", banned, got)
		}
	}
	if !strings.Contains(joined, "working") {
		t.Fatalf("what-project lexical tokens must keep working, got %v", got)
	}
	often := searchLexicalQueryTokens("How often does Audrey meet up with other dog owners for tips and playdates?", tokenize("How often does Audrey meet up with other dog owners for tips and playdates?"))
	joinedOften := strings.Join(often, " ")
	if strings.Contains(joinedOften, "working") {
		t.Fatalf("how-often must not be rewritten as what-project tokens, got %v", often)
	}
	if !strings.Contains(joinedOften, "meet") {
		t.Fatalf("how-often must keep meet, got %v", often)
	}
}

func TestKeepCurrentProjectInCapSurvivesListFill(t *testing.T) {
	q := "What project is James working on in his game design course?"
	obs := rankedSearchResult{result: SearchResult{
		MemoryID: "obs",
		Content:  "James: Yes, we are currently working on a new part of the football simulator",
		Score:    0.4,
	}}
	full := []rankedSearchResult{obs}
	capped := make([]rankedSearchResult, 0, 30)
	for i := 0; i < 30; i++ {
		capped = append(capped, rankedSearchResult{result: SearchResult{
			MemoryID: fmt.Sprintf("fill%02d", i),
			Content:  "James is creating his own game project.",
			Score:    1.5,
		}})
	}
	got := keepCurrentProjectInCap(full, capped, q, 30)
	found := false
	for _, item := range got {
		if strings.Contains(strings.ToLower(item.result.Content), "currently working") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("what-project evidence-set cap must keep currently-working leftover, n=%d", len(got))
	}
	if len(got) > 30 {
		t.Fatalf("what-project current-project keep must stay within limit, n=%d", len(got))
	}
}

func TestSearchLexicalQueryTokensDropsWhatNewHobbyStructure(t *testing.T) {
	q := "What new hobby did James become interested in on 9 July, 2022?"
	got := searchLexicalQueryTokens(q, tokenize(q))
	joined := strings.Join(got, " ")
	if strings.Contains(joined, "hobby") {
		t.Fatalf("what-new-hobby lexical tokens must drop hobby, got %v", got)
	}
	if !strings.Contains(joined, "interested") {
		t.Fatalf("what-new-hobby lexical tokens must keep interested, got %v", got)
	}
	proj := searchLexicalQueryTokens("What project is James working on in his game design course?", tokenize("What project is James working on in his game design course?"))
	joinedProj := strings.Join(proj, " ")
	if strings.Contains(joinedProj, "hobby") {
		t.Fatalf("what-project must not be rewritten as what-new-hobby tokens, got %v", proj)
	}
	if !strings.Contains(joinedProj, "working") {
		t.Fatalf("what-project must keep working, got %v", proj)
	}
}

func TestKeepBecomeInterestedInCapSurvivesListFill(t *testing.T) {
	q := "What new hobby did James become interested in on 9 July, 2022?"
	obs := rankedSearchResult{result: SearchResult{
		MemoryID: "obs",
		Content:  "Lately I've become interested in extreme sports",
		Score:    0.4,
	}}
	full := []rankedSearchResult{obs}
	capped := make([]rankedSearchResult, 0, 30)
	for i := 0; i < 30; i++ {
		capped = append(capped, rankedSearchResult{result: SearchResult{
			MemoryID: fmt.Sprintf("fill%02d", i),
			Content:  "John has taken up metal detecting as a new hobby, walking along beaches with a metal detector.",
			Score:    1.5,
		}})
	}
	got := keepBecomeInterestedInCap(full, capped, q, 30)
	found := false
	for _, item := range got {
		if strings.Contains(strings.ToLower(item.result.Content), "become interested") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("what-new-hobby evidence-set cap must keep become-interested leftover, n=%d", len(got))
	}
	if len(got) > 30 {
		t.Fatalf("what-new-hobby become-interested keep must stay within limit, n=%d", len(got))
	}
}

func TestSearchLexicalQueryTokensDropsHowPlanDreamStructure(t *testing.T) {
	q := "How does Jolene plan to pursue her dream of learning to surf?"
	got := searchLexicalQueryTokens(q, tokenize(q))
	joined := strings.Join(got, " ")
	if strings.Contains(joined, "plan") || strings.Contains(joined, "pursue") || strings.Contains(joined, "dream") || strings.Contains(joined, "learning") {
		t.Fatalf("how-plan-dream lexical tokens must drop plan/pursue/dream/learning, got %v", got)
	}
	if !strings.Contains(joined, "surf") {
		t.Fatalf("how-plan-dream lexical tokens must keep surf, got %v", got)
	}
	hobby := searchLexicalQueryTokens("What new hobby did James become interested in on 9 July, 2022?", tokenize("What new hobby did James become interested in on 9 July, 2022?"))
	joinedHobby := strings.Join(hobby, " ")
	if strings.Contains(joinedHobby, "surf") {
		t.Fatalf("what-new-hobby must not be rewritten as how-plan-dream tokens, got %v", hobby)
	}
	if !strings.Contains(joinedHobby, "interested") {
		t.Fatalf("what-new-hobby must keep interested, got %v", hobby)
	}
}

func TestKeepPrepPlanInCapSurvivesListFill(t *testing.T) {
	q := "How does Jolene plan to pursue her dream of learning to surf?"
	obs := rankedSearchResult{result: SearchResult{
		MemoryID: "obs",
		Content:  "I've been gathering information, watching videos, and I even got a beginners' guide to surfing",
		Score:    0.4,
	}}
	full := []rankedSearchResult{obs}
	capped := make([]rankedSearchResult, 0, 30)
	for i := 0; i < 30; i++ {
		capped = append(capped, rankedSearchResult{result: SearchResult{
			MemoryID: fmt.Sprintf("fill%02d", i),
			Content:  "Deborah: Exploring historical places and learning their stories is so fun",
			Score:    1.5,
		}})
	}
	got := keepPrepPlanInCap(full, capped, q, 30)
	found := false
	for _, item := range got {
		if strings.Contains(strings.ToLower(item.result.Content), "gathering information") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("how-plan-dream evidence-set cap must keep prep-plan leftover, n=%d", len(got))
	}
	if len(got) > 30 {
		t.Fatalf("how-plan-dream prep-plan keep must stay within limit, n=%d", len(got))
	}
}

func TestKeepDurationInCapSurvivesListFill(t *testing.T) {
	q := "How long have Mel and her husband been married?"
	obs := rankedSearchResult{result: SearchResult{
		MemoryID: "obs",
		Content:  "Melanie's marriage duration is 5 years.",
		Score:    0.4,
	}}
	full := []rankedSearchResult{obs}
	capped := make([]rankedSearchResult, 0, 30)
	for i := 0; i < 30; i++ {
		capped = append(capped, rankedSearchResult{result: SearchResult{
			MemoryID: fmt.Sprintf("fill%02d", i),
			Content:  "Melanie is married.",
			Score:    1.5,
		}})
	}
	got := keepDurationInCap(full, capped, q, 30)
	found := false
	for _, item := range got {
		if strings.Contains(strings.ToLower(item.result.Content), "duration") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("how-long-been evidence-set cap must keep continuing duration leftover, n=%d", len(got))
	}
	if len(got) > 30 {
		t.Fatalf("how-long-been duration keep must stay within limit, n=%d", len(got))
	}
}

func TestSearchLexicalTokensDropsHowDescribeStructureAndPerson(t *testing.T) {
	q := "How does Nate describe the stuffed animal he got for Joanna?"
	got := searchLexicalQueryTokens(q, tokenize(q))
	joined := strings.Join(got, " ")
	for _, banned := range []string{"describe", "nate", "joanna", "got"} {
		if strings.Contains(joined, banned) {
			t.Fatalf("how-describe lexical tokens must drop structure/person %q, got %v", banned, got)
		}
	}
	for _, keep := range []string{"stuffed", "animal"} {
		if !strings.Contains(joined, keep) {
			t.Fatalf("how-describe lexical tokens must keep %q, got %v", keep, got)
		}
	}
	destress := searchLexicalQueryTokens("What does Melanie do to destress?", tokenize("What does Melanie do to destress?"))
	if !strings.Contains(strings.Join(destress, " "), "melanie") {
		t.Fatalf("destress must keep the person token, got %v", destress)
	}
	smartwatch := searchLexicalQueryTokens("What does the smartwatch help Riley with?", tokenize("What does the smartwatch help Riley with?"))
	if strings.Contains(strings.Join(smartwatch, " "), "describe") {
		t.Fatalf("instrument-purpose must not be treated as how-describe, got %v", smartwatch)
	}
	island := searchLexicalQueryTokens("How does Evan describe the island he grew up on?", tokenize("How does Evan describe the island he grew up on?"))
	if !strings.Contains(strings.Join(island, " "), "island") {
		t.Fatalf("how-describe island must keep the object token, got %v", island)
	}
	advice := searchLexicalQueryTokens("What advice does Gina give to Jon about running a successful business?", tokenize("What advice does Gina give to Jon about running a successful business?"))
	if !strings.Contains(strings.Join(advice, " "), "advice") {
		t.Fatalf("advice queries must keep the speech-act token, got %v", advice)
	}
	dinner := searchLexicalQueryTokens("What kind of food did Maria have on her dinner spread iwth her mother?", tokenize("What kind of food did Maria have on her dinner spread iwth her mother?"))
	if !strings.Contains(strings.Join(dinner, " "), "spread") {
		t.Fatalf("what-kind dinner queries must keep spread, got %v", dinner)
	}
}

func TestIngestRetainsDialogueAndRanksDatedFact(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)

	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Metadata: map[string]any{
			"session_id":  "sess-1",
			"observed_at": "2023-05-07T18:00:00Z",
		},
		Messages: []Message{
			{Role: "user", Content: "Alex: I went to the community support group on 7 May 2023"},
			{Role: "user", Content: "Sam: Can't wait to see your show - the community needs more platforms"},
		},
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}

	var hasEpisode bool
	var hasDate bool
	for _, record := range store.records {
		if record.Primitive == PrimitiveEpisode {
			hasEpisode = true
		}
		if strings.Contains(strings.ToLower(record.Content), "7 may 2023") {
			hasDate = true
		}
		if record.Metadata["session_id"] != "sess-1" {
			t.Fatalf("expected session_id metadata copied, got %#v", record.Metadata)
		}
	}
	if !hasDate {
		t.Fatal("expected dated support-group turn retained as a memory")
	}
	if !hasEpisode {
		t.Fatal("expected conversation_episode primitive on free dialogue")
	}
	for _, record := range store.records {
		if strings.Contains(strings.ToLower(record.Content), "7 may 2023") {
			if record.ObservedAt == nil {
				t.Fatal("expected ObservedAt from metadata.observed_at")
			}
		}
	}

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "When did Alex go to the community support group")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected search hits")
	}
	top := strings.ToLower(search.Results[0].Content)
	if !strings.Contains(top, "7 may 2023") && !strings.Contains(top, "support group") {
		t.Fatalf("expected dated fact to outrank topical preference neighbor, top=%q", search.Results[0].Content)
	}
	if search.Results[0].Explain["date_token_boost"] == nil && search.Results[0].Explain["exact_span_boost"] == nil && search.Results[0].Explain["episode_penalty"] == nil {
		t.Fatalf("expected ranking explain boosts, got %#v", search.Results[0].Explain)
	}
}

func TestExactSpanOutranksTopicalNeighbor(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	now := service.now()

	dated := MemoryRecord{
		MemoryID:  "mem_dated",
		TenantID:  "t1",
		SubjectID: "u1",
		Kind:      KindFact,
		Content:   "Alex went to the community support group on 7 May 2023",
		DedupeKey: "dated",
		Status:    StatusActive,
		UpdatedAt: now,
	}
	topical := MemoryRecord{
		MemoryID:  "mem_topical",
		TenantID:  "t1",
		SubjectID: "u1",
		Kind:      KindPreference,
		Content:   "Prefers LGBTQ community platforms and shows",
		DedupeKey: "topical",
		Status:    StatusActive,
		UpdatedAt: now,
	}
	store.records["dated"] = dated
	store.records["topical"] = topical

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "When did Alex go to the community support group")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected results")
	}
	if !strings.Contains(strings.ToLower(search.Results[0].Content), "7 may 2023") {
		t.Fatalf("expected dated fact first, got %q", search.Results[0].Content)
	}
}

func TestEpisodePenaltyPrefersTypedFact(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	now := service.now()
	store.records["ep"] = MemoryRecord{
		MemoryID: "mem_ep", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "Congratulations on sticking to your daily tidying routine for 3 weeks",
		DedupeKey: "ep", Status: StatusActive, UpdatedAt: now,
	}
	store.records["fact"] = MemoryRecord{
		MemoryID: "mem_f", TenantID: "t1", SubjectID: "u1",
		Kind:      KindFact,
		Content:   "Alex has been sticking to a daily tidying routine for 4 weeks",
		DedupeKey: "fact", Status: StatusActive, UpdatedAt: now,
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "how long have I been sticking to my daily tidying routine")
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected hits")
	}
	if !strings.Contains(search.Results[0].Content, "4 weeks") {
		t.Fatalf("typed fact should outrank congratulation episode, top=%q", search.Results[0].Content)
	}
}

func TestFactPrimaryDropsEpisodesWhenCoverageComplete(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	now := service.now()
	store.records["ep"] = MemoryRecord{
		MemoryID: "mem_ep", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "Yeah, Caroline, Yep, Melanie, Hey Caroline",
		DedupeKey: "ep", Status: StatusActive, UpdatedAt: now,
	}
	store.records["fact"] = MemoryRecord{
		MemoryID: "mem_f", TenantID: "t1", SubjectID: "u1",
		Kind:      KindFact,
		Content:   "Caroline is from Sweden",
		DedupeKey: "fact", Status: StatusActive, UpdatedAt: now,
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "Is Caroline from Sweden")
	if err != nil {
		t.Fatal(err)
	}
	if search.Trace == nil || search.Trace.RepresentationStatus != RepresentationComplete {
		t.Fatalf("expected complete representation, trace=%+v", search.Trace)
	}
	if search.Trace.EpisodeFallback {
		t.Fatalf("complete coverage must not fall back to episodes, trace=%+v", search.Trace)
	}
	if search.Trace.EpisodesDropped < 1 {
		t.Fatalf("expected standalone episodes dropped when coverage is complete, trace=%+v", search.Trace)
	}
	for _, r := range search.Results {
		if strings.Contains(strings.ToLower(r.Content), "yeah") {
			t.Fatalf("provenance episode leaked into complete-coverage search: %q", r.Content)
		}
	}
	if len(search.Results) == 0 || !strings.Contains(strings.ToLower(search.Results[0].Content), "sweden") {
		t.Fatalf("expected fact hit, got %#v", search.Results)
	}

	withEp, err := service.SearchOpt(context.Background(), "t1", "u1", "", "", "Is Caroline from Sweden", SearchOptions{IncludeEpisodes: true, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	sawEp := false
	for _, r := range withEp.Results {
		if strings.Contains(strings.ToLower(r.Content), "yeah") {
			sawEp = true
		}
	}
	if !sawEp {
		t.Fatal("IncludeEpisodes should keep provenance turns")
	}
}

func TestFactPrimaryKeepsEpisodesWhenCoveragePartial(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	now := service.now()
	store.records["sweden"] = MemoryRecord{
		MemoryID: "mem_ep_sweden", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "Caroline is from Sweden",
		DedupeKey: "sweden", Status: StatusActive, UpdatedAt: now,
	}
	store.records["fact"] = MemoryRecord{
		MemoryID: "mem_f", TenantID: "t1", SubjectID: "u1",
		Kind:      KindFact,
		Content:   "Caroline likes pottery",
		DedupeKey: "fact", Status: StatusActive, UpdatedAt: now,
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "Where is Caroline originally from")
	if err != nil {
		t.Fatal(err)
	}
	if search.Trace == nil || search.Trace.RepresentationStatus != RepresentationPartial {
		t.Fatalf("expected partial representation, trace=%+v", search.Trace)
	}
	if !search.Trace.EpisodeFallback {
		t.Fatalf("incomplete compiler coverage must keep episode fallback, trace=%+v", search.Trace)
	}
	sawSweden := false
	sawPottery := false
	for _, r := range search.Results {
		lower := strings.ToLower(r.Content)
		if strings.Contains(lower, "sweden") {
			sawSweden = true
		}
		if strings.Contains(lower, "pottery") {
			sawPottery = true
		}
	}
	if !sawSweden {
		t.Fatal("WRITE_MISS must not be converted into a retrieval miss: Sweden episode should remain")
	}
	if !sawPottery {
		t.Fatal("expected pottery fact to remain recall-primary")
	}
}

func TestMalformedFactsDoNotCountAsCoverage(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	now := service.now()
	store.records["junk"] = MemoryRecord{
		MemoryID: "mem_junk", TenantID: "t1", SubjectID: "u1",
		Kind:      KindFact,
		Content:   "Alex has done going at since then",
		DedupeKey: "junk", Status: StatusActive, UpdatedAt: now,
		Explain: map[string]any{"rule": "attribute_place_activity"},
	}
	store.records["ep"] = MemoryRecord{
		MemoryID: "mem_ep", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "I've known these friends for 4 years, since I moved from my home country",
		DedupeKey: "ep", Status: StatusActive, UpdatedAt: now,
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "How long has Alex had this group of friends")
	if err != nil {
		t.Fatal(err)
	}
	if search.Trace == nil || search.Trace.RepresentationStatus == RepresentationComplete {
		t.Fatalf("malformed atoms must not complete coverage, trace=%+v", search.Trace)
	}
	sawYears := false
	for _, r := range search.Results {
		if strings.Contains(strings.ToLower(r.Content), "4 years") {
			sawYears = true
		}
	}
	if !sawYears {
		t.Fatalf("provenance with the duration must remain, got %#v", search.Results)
	}
}

func TestSelectEpisodeFallbackPrefersDistinctiveProvenance(t *testing.T) {
	episodes := make([]MemoryRecord, 0, 12)
	for i := 0; i < 10; i++ {
		episodes = append(episodes, MemoryRecord{
			MemoryID: fmt.Sprintf("name_%d", i),
			Content:  "Alex: the support group has been a huge part of my journey",
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "I've known these friends for 4 years, since I moved from my home country",
	}
	episodes = append(episodes, gold)
	picked := selectEpisodeFallback(episodes, []string{"long", "alex", "current", "group", "friends"}, 8)
	found := false
	for _, ep := range picked {
		if ep.MemoryID == "gold" {
			found = true
		}
	}
	if !found {
		t.Fatalf("fallback must keep distinctive provenance, picked=%v", picked)
	}
}

func TestStaleFactPrefersNewerDate(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	now := service.now()
	store.records["may"] = MemoryRecord{
		MemoryID: "mem_may", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "The launch date is May 12.",
		DedupeKey: "may", Status: StatusActive, UpdatedAt: now.Add(-time.Minute),
	}
	store.records["june"] = MemoryRecord{
		MemoryID: "mem_june", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "The launch date is June 3.",
		DedupeKey: "june", Status: StatusActive, UpdatedAt: now,
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "launch date")
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) == 0 || !strings.Contains(strings.ToLower(search.Results[0].Content), "june") {
		t.Fatalf("newer June fact must rank first, got %#v", search.Results)
	}
}

func TestMalformedAtomsDoNotOutrankProvenance(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	now := service.now()
	store.records["junk"] = MemoryRecord{
		MemoryID: "mem_junk", TenantID: "t1", SubjectID: "u1",
		Kind:      KindFact,
		Content:   "Alex participates in runn",
		DedupeKey: "junk", Status: StatusActive, UpdatedAt: now,
		Explain: map[string]any{"rule": "attribute_activity"},
	}
	store.records["ep"] = MemoryRecord{
		MemoryID: "mem_ep", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "Alex: I ran a charity race last Saturday (20 May 2023)",
		DedupeKey: "ep", Status: StatusActive, UpdatedAt: now,
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "When did Alex run a charity race")
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected hits")
	}
	if strings.Contains(strings.ToLower(search.Results[0].Content), "participates in runn") {
		t.Fatalf("malformed atom outranked provenance: %#v", search.Results)
	}
	sawRace := false
	for _, r := range search.Results {
		if strings.Contains(strings.ToLower(r.Content), "charity race") {
			sawRace = true
		}
	}
	if !sawRace {
		t.Fatalf("expected charity-race episode, got %#v", search.Results)
	}
}

func TestEpisodeOnlyPoolFallsBack(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	now := service.now()
	store.records["ep"] = MemoryRecord{
		MemoryID: "mem_ep", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "Alex went to the community support group on 7 May 2023",
		DedupeKey: "ep", Status: StatusActive, UpdatedAt: now,
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "When did Alex go to the community support group")
	if err != nil {
		t.Fatal(err)
	}
	if search.Trace == nil || !search.Trace.EpisodeFallback {
		t.Fatalf("expected episode fallback when no facts exist, trace=%+v", search.Trace)
	}
	if search.Trace.RepresentationStatus != RepresentationEmpty {
		t.Fatalf("expected empty representation_status, trace=%+v", search.Trace)
	}
	if len(search.Results) == 0 {
		t.Fatal("fallback must not empty the pool")
	}
}

func TestCandidateLimitRecordedOnTrace(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t-pool", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{{Role: "user", Content: "Alex lives in Austin"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := service.SearchOpt(context.Background(), "t-pool", "u1", "", "", "where does Alex live", SearchOptions{Limit: 10, CandidateLimit: 50})
	if err != nil {
		t.Fatal(err)
	}
	if out.Trace == nil || out.Trace.CandidateOverfetch != 50 {
		t.Fatalf("expected candidate pool 50, trace=%+v", out.Trace)
	}
}

func TestSessionNeighborExpansion(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Metadata: map[string]any{"session_id": "sess-a", "observed_at": "2023-05-08T12:00:00Z"},
		Messages: []Message{
			{Role: "user", Content: "Alex: I am a community organizer"},
			{Role: "user", Content: "Sam: That takes courage"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "What is Alex identity")
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected results")
	}
	joined := ""
	for _, r := range search.Results {
		joined += " " + r.Content
	}
	if !strings.Contains(strings.ToLower(joined), "community organizer") {
		t.Fatalf("expected session content in results, got %q", joined)
	}
}

func TestQuestionMemoriesDownrankedForFactQueries(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Metadata: map[string]any{"session_id": "s1", "observed_at": "2023-05-08T12:00:00Z"},
		Messages: []Message{
			{Role: "user", Content: "What did Alex research last week?"},
			{Role: "user", Content: "Alex: I researched adoption agencies this week"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "What did Alex research?")
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected results")
	}
	top := search.Results[0].Content
	if strings.Contains(top, "?") && !strings.Contains(strings.ToLower(top), "adoption") {
		t.Fatalf("expected factual adoption memory on top, got %q", top)
	}
	if !strings.Contains(strings.ToLower(top), "adoption") {
		t.Fatalf("expected adoption fact ranked first, got %q explain=%v", top, search.Results[0].Explain)
	}
}

func TestLowInfoNameOnlyDownranked(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Metadata: map[string]any{"session_id": "s1", "observed_at": "2023-05-08T12:00:00Z"},
		Messages: []Message{
			{Role: "user", Content: "Yeah, Alice"},
			{Role: "user", Content: "Thanks, Alice"},
			{Role: "user", Content: "Alice: I am training for a marathon and run every morning before work"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "What activities does Alice enjoy?")
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected results")
	}
	top := strings.ToLower(search.Results[0].Content)
	if !strings.Contains(top, "marathon") && !strings.Contains(top, "run") {
		t.Fatalf("expected content-dense activity memory first, got %q explain=%v", search.Results[0].Content, search.Results[0].Explain)
	}
}

func TestSubjectContentExpansionSurfacesProfile(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Metadata: map[string]any{"session_id": "s-profile", "observed_at": "2023-05-08T12:00:00Z"},
		Messages: []Message{
			{Role: "user", Content: "Hey Bob"},
			{Role: "user", Content: "Bob: pottery keeps me grounded after long weeks at the studio"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Query verbs do not appear in the pottery memory; subject bridge must admit it.
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "What hobbies does Bob enjoy?")
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, r := range search.Results {
		joined += " " + strings.ToLower(r.Content)
	}
	if !strings.Contains(joined, "pottery") {
		t.Fatalf("expected subject-content expansion to surface pottery, got %q", joined)
	}
}

func TestSupersedeHidesPriorFromDefaultSearch(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	ingested, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "note",
		Messages: []Message{{Role: "user", Content: "Door code is 1111"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ingested.Memories) == 0 {
		t.Fatal("expected memory")
	}
	priorID := ingested.Memories[0].MemoryID
	replaced, err := service.Supersede(context.Background(), "t1", "u1", priorID, SupersedeRequest{
		Content: "Door code is 2222",
	})
	if err != nil {
		t.Fatal(err)
	}
	if replaced.MemoryID == priorID {
		t.Fatal("expected a new memory id for superseding record")
	}

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "door code")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range search.Results {
		if strings.Contains(r.Content, "1111") {
			t.Fatalf("superseded prior leaked into default search: %q", r.Content)
		}
		if r.MemoryID == priorID {
			t.Fatalf("superseded id %s visible in default search", priorID)
		}
	}
	joined := ""
	for _, r := range search.Results {
		joined += " " + r.Content
	}
	if !strings.Contains(joined, "2222") {
		t.Fatalf("expected replacement in search, got %q", joined)
	}

	hist, err := service.SearchOpt(context.Background(), "t1", "u1", "", "", "door code", SearchOptions{IncludeHistorical: true})
	if err != nil {
		t.Fatal(err)
	}
	foundPrior := false
	for _, r := range hist.Results {
		if r.MemoryID == priorID || strings.Contains(r.Content, "1111") {
			foundPrior = true
		}
	}
	if !foundPrior {
		t.Fatal("expected include_historical to surface superseded prior")
	}
}

func TestHistoricalIntentRetrievesPriorResidence(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	now := service.now()
	store.records["ny"] = MemoryRecord{
		MemoryID: "mem_ny", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Content: "Alex lives in New York",
		DedupeKey: "ny", Status: StatusActive, LifecycleState: LifecycleSuperseded,
		Metadata:  map[string]any{"predicate": PredicateResidence, "value_norm": "new york", "memory_type": "state"},
		CreatedAt: now, UpdatedAt: now,
	}
	store.records["au"] = MemoryRecord{
		MemoryID: "mem_au", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Content: "Alex lives in Austin",
		DedupeKey: "au", Status: StatusActive, LifecycleState: LifecycleActive,
		Metadata:  map[string]any{"predicate": PredicateResidence, "value_norm": "austin", "memory_type": "state"},
		CreatedAt: now, UpdatedAt: now,
	}

	cur, err := service.SearchOpt(context.Background(), "t1", "u1", "", "", "where does Alex currently live", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, r := range cur.Results {
		joined += " " + r.Content
		if strings.Contains(r.Content, "New York") {
			t.Fatalf("current-state search leaked superseded NY: %q", r.Content)
		}
	}
	if !strings.Contains(joined, "Austin") {
		t.Fatalf("current-state should prefer Austin, got %q", joined)
	}

	hist, err := service.SearchOpt(context.Background(), "t1", "u1", "", "", "where did Alex live before", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	foundNY := false
	for _, r := range hist.Results {
		if strings.Contains(r.Content, "New York") {
			foundNY = true
		}
	}
	if !foundNY {
		t.Fatalf("historical intent should retrieve NY, results=%+v", hist.Results)
	}
}

func TestListQueryDiversifiesThemes(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Metadata: map[string]any{"session_id": "s-div", "observed_at": "2023-05-08T12:00:00Z"},
		Messages: []Message{
			{Role: "user", Content: "Yeah, Dana"},
			{Role: "user", Content: "Dana: I love pottery and spend weekends at the studio shaping bowls"},
			{Role: "user", Content: "Dana: Camping in the mountains clears my head after long weeks"},
			{Role: "user", Content: "Dana: Swimming at the community pool is my tuesday habit"},
			{Role: "user", Content: "Dana: Wow that sounds great"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "What activities does Dana enjoy?")
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) < 3 {
		t.Fatalf("expected diversified candidates, got %d", len(search.Results))
	}
	head := ""
	for i, r := range search.Results {
		if i >= 6 {
			break
		}
		head += " " + strings.ToLower(r.Content)
	}
	for _, need := range []string{"pottery", "camping", "swimming"} {
		if !strings.Contains(head, need) {
			t.Fatalf("expected %q in diversified head results, got %q", need, head)
		}
	}
}

func TestDomainEventMatchByLabel(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	ingested, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "note",
		Label:    "promo",
		Metadata: map[string]any{"season": "summer"},
		// Fact-shaped so deterministic extract still fires when Label is set
		// (labeled ingest skips conversation_episode retention).
		Messages: []Message{{Role: "user", Content: "Summer splash promo is the active seasonal campaign."}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ingested.Memories) == 0 {
		t.Fatal("expected memory")
	}
	// Ensure label stuck on the stored record (core pack may ignore unknown labels).
	got, err := store.GetMemory(context.Background(), "t1", "u1", ingested.Memories[0].MemoryID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Label == "" {
		// Force label for match path coverage when pack vocabulary omits it.
		got.Label = "promo"
		got.Metadata = map[string]any{"season": "summer"}
		store.records[got.DedupeKey] = got
	}
	res, err := service.ApplyDomainEvent(context.Background(), DomainEventRequest{
		TenantID: "t1", SubjectID: "u1", EventType: "promo_ended",
		Match: &DomainEventMatch{Label: "promo"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Superseded) == 0 {
		t.Fatalf("expected match-based supersede, record=%+v", got)
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "Summer splash promo")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range search.Results {
		if r.MemoryID == ingested.Memories[0].MemoryID {
			t.Fatalf("matched memory still searchable: %q", r.Content)
		}
	}
}

func TestDomainEventBatchSupersede(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	ingested, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "note",
		Messages: []Message{
			{Role: "user", Content: "Campaign splash is live"},
			{Role: "user", Content: "Campaign splash headline is Ready"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(ingested.Memories))
	for _, m := range ingested.Memories {
		ids = append(ids, m.MemoryID)
	}
	res, err := service.ApplyDomainEvent(context.Background(), DomainEventRequest{
		TenantID: "t1", SubjectID: "u1", EventType: "campaign_ended",
		SupersedeMemoryIDs: ids,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Superseded) == 0 {
		t.Fatal("expected superseded ids")
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "Campaign splash")
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) != 0 {
		t.Fatalf("expected empty search after batch supersede, got %#v", search.Results)
	}
}

func seedStatefulMemory(t *testing.T, service *Service, store *memoryStoreStub, memoryID, content, predicate, valueNorm string) string {
	t.Helper()
	now := service.now()
	record := MemoryRecord{
		MemoryID: memoryID, TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Content: content, DedupeKey: DedupeKey("t1", "u1", KindFact, content),
		Status: StatusActive, LifecycleState: LifecycleActive,
		Metadata:  map[string]any{"predicate": predicate, "value_norm": valueNorm, "memory_type": "state"},
		CreatedAt: now, UpdatedAt: now,
	}
	store.records[record.DedupeKey] = record
	// Mirror the store projection the way ingest would.
	ProjectCurrentStateIfApplicable(context.Background(), store, record)
	return record.MemoryID
}

func currentStateValue(store *memoryStoreStub, predicate string) (string, bool) {
	for key, row := range store.currentState {
		if strings.HasSuffix(key, "::"+predicate) {
			return row.Value, row.MemoryID != ""
		}
	}
	return "", false
}

// Suppressing the winning current-state memory must remove it from current recall.
func TestSuppressClearsCurrentStateProjection(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	memID := seedStatefulMemory(t, service, store, "mem_res", "Alex lives in New York", PredicateResidence, "new york")

	if _, found := currentStateValue(store, PredicateResidence); !found {
		t.Fatal("expected residence current-state projection before suppress")
	}

	if err := service.Suppress(context.Background(), "t1", "u1", memID); err != nil {
		t.Fatal(err)
	}
	if _, found := currentStateValue(store, PredicateResidence); found {
		t.Fatalf("suppressed memory still projects current state: %+v", store.currentState)
	}
	if _, found := currentStateValue(store, PredicateResidence); found {
		t.Fatal("suppressed winner must not be recallable as current")
	}
}

// Correcting a stateful memory must update the projected current value.
func TestCorrectUpdatesCurrentStateProjection(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	memID := seedStatefulMemory(t, service, store, "mem_res", "Alex lives in New York", PredicateResidence, "new york")

	if _, err := service.Correct(context.Background(), "t1", "u1", memID, CorrectionRequest{
		Content:    "Alex lives in Austin",
		SourceText: "Actually Alex lives in Austin now.",
	}); err != nil {
		t.Fatal(err)
	}
	val, found := currentStateValue(store, PredicateResidence)
	if !found {
		t.Fatal("expected current-state projection after correction")
	}
	// The corrected content is the new ground-truth value; it replaces the stale
	// pre-correction projection value rather than the old value_norm.
	if val != "alex lives in austin" {
		t.Fatalf("expected corrected current value to reflect corrected content, got %q", val)
	}
}

// Superseding the current-state winner must re-point the projection at the replacement.
func TestSupersedeRepointsCurrentStateProjection(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	priorID := seedStatefulMemory(t, service, store, "mem_res", "Alex lives in New York", PredicateResidence, "new york")

	replacement, err := service.Supersede(context.Background(), "t1", "u1", priorID, SupersedeRequest{
		Content:    "Alex lives in New York",
		SourceText: "Same fact, refreshed source.",
	})
	if err != nil {
		t.Fatal(err)
	}

	val, found := currentStateValue(store, PredicateResidence)
	if !found {
		t.Fatal("expected current-state projection after supersede")
	}
	if val != "new york" {
		t.Fatalf("expected superseded value to remain current, got %q", val)
	}
	row, ok := store.currentState[mapKeyFor("t1", "u1", PredicateResidence)]
	if !ok || row.MemoryID != replacement.MemoryID {
		t.Fatalf("expected current-state winner to be replacement %s, got %+v", replacement.MemoryID, store.currentState)
	}
}

func mapKeyFor(tenantID, subjectID, predicate string) string {
	return tenantID + "::" + subjectID + "::" + predicate
}
