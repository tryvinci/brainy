package jobs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"brainy/internal/embedding"
	"brainy/internal/memory"
	"brainy/internal/observability"
	"brainy/internal/pack"
)

type Processor struct {
	store     memory.Store
	extractor memory.Extractor
	embedder  embedding.Embedder
	packs     *pack.Registry
	metrics   *observability.Metrics
	now       func() time.Time
	id        func(prefix string) string
}

func NewProcessor(store memory.Store, metrics *observability.Metrics) *Processor {
	return NewProcessorWithExtractor(store, metrics, memory.NewDeterministicExtractor())
}

func NewProcessorWithExtractor(store memory.Store, metrics *observability.Metrics, extractor memory.Extractor) *Processor {
	reg, _ := pack.LoadRegistryFromDir("packs")
	if extractor == nil {
		extractor = memory.NewDeterministicExtractor()
	}
	return &Processor{
		store:     store,
		extractor: extractor,
		embedder:  embedding.Default(),
		packs:     reg,
		metrics:   metrics,
		now:       time.Now().UTC,
		id: func(prefix string) string {
			return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
		},
	}
}

func (p *Processor) WithEmbedder(embedder embedding.Embedder) *Processor {
	if embedder != nil {
		p.embedder = embedder
	}
	return p
}

func (p *Processor) ProcessNext(ctx context.Context) (bool, error) {
	job, ok, err := p.store.ClaimNextExtractionJob(ctx)
	if err != nil || !ok {
		return ok, err
	}

	memory.NormalizeIngestRequest(&job.Request)
	// Extract may exceed the initial lease (provider calls up to 45s vs a 30s
	// default lease). Renew the lease for the current owner so a live worker is
	// not reclaimed mid-extraction.
	if fencer, ok := p.store.(memory.LeaseFencer); ok {
		_ = fencer.HeartbeatExtractionJob(ctx, job.JobID, job.LeaseOwner)
		defer func() {
			_ = fencer.HeartbeatExtractionJob(ctx, job.JobID, job.LeaseOwner)
		}()
	}
	extracted, err := p.extractor.Extract(ctx, job.Request)
	if err != nil {
		// Fail before any upserts so a provider error cannot leave partial
		// enrichment or mutate the immutable raw_ingests payload.
		_ = p.failJob(ctx, job, err.Error())
		return true, err
	}

	mode := memory.WriteMutationModeOf(job.Request)
	memory.PersistIngestAliases(ctx, p.store, job.Request)
	for _, item := range extracted {
		p.metrics.RecordExtraction()
		if memory.MemoryEventOf(item) == memory.MemoryEventDelete {
			_ = memory.ApplyDeleteMemoryEvent(ctx, p.store, job.Request.TenantID, job.Request.SubjectID, item, mode)
			if mode == memory.WriteModeGoverned {
				continue
			}
		}
		if !memory.PrepareExtractedForPersist(&item, mode) {
			continue
		}
		record, err := memory.BuildMemoryRecord(p.id("mem"), p.now(), job.Request, item, p.packs)
		if err != nil {
			_ = p.failJob(ctx, job, err.Error())
			return true, err
		}
		upserted, err := p.store.UpsertMemory(ctx, record)
		if err != nil {
			_ = p.failJob(ctx, job, err.Error())
			return true, err
		}
		if err := p.persistEmbedding(ctx, upserted.Record); err != nil {
			_ = p.failJob(ctx, job, err.Error())
			return true, err
		}
		p.persistEntityLinks(ctx, upserted.Record)
		p.persistEvidenceAndEvents(ctx, upserted.Record)
		// Match sync ingest: supersede older state, then project current_state
		// only when shouldReplaceCurrentState allows (no blind late-older wins).
		_ = memory.AutoSupersedePriorState(ctx, p.store, upserted.Record)
		_ = memory.ApplyIngestSupersession(ctx, p.store, upserted.Record)
		fresh := upserted.Record
		if got, err := p.store.GetMemory(ctx, fresh.TenantID, fresh.SubjectID, fresh.MemoryID); err == nil {
			fresh = got
		}
		memory.ProjectCurrentStateIfApplicable(ctx, p.store, fresh)
	}

	if err := p.completeJob(ctx, job); err != nil {
		return true, err
	}
	p.persistRuntime(ctx)
	return true, nil
}

// completeJob uses the owner-fenced completion when the store supports it, so a
// reclaimed job cannot be committed by its old worker.
func (p *Processor) completeJob(ctx context.Context, job memory.ExtractionJob) error {
	if fencer, ok := p.store.(memory.LeaseFencer); ok {
		return fencer.CompleteExtractionJobFenced(ctx, job.JobID, job.IngestID, job.LeaseOwner)
	}
	return p.store.CompleteExtractionJob(ctx, job.JobID, job.IngestID)
}

// failJob uses the owner-fenced failure path when the store supports it.
func (p *Processor) failJob(ctx context.Context, job memory.ExtractionJob, reason string) error {
	p.persistRuntime(ctx)
	if fencer, ok := p.store.(memory.LeaseFencer); ok {
		return fencer.FailExtractionJobFenced(ctx, job.JobID, job.IngestID, job.LeaseOwner, reason)
	}
	return p.store.FailExtractionJob(ctx, job.JobID, job.IngestID, reason)
}

// ProcessAvailable runs up to concurrency parallel ProcessNext calls.
// Used for LME-scale queue drain; concurrency<=1 keeps prior serial behavior.
func (p *Processor) ProcessAvailable(ctx context.Context, concurrency int) (int, error) {
	if concurrency <= 1 {
		ok, err := p.ProcessNext(ctx)
		if ok {
			return 1, err
		}
		return 0, err
	}
	type result struct {
		ok  bool
		err error
	}
	ch := make(chan result, concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			ok, err := p.ProcessNext(ctx)
			ch <- result{ok: ok, err: err}
		}()
	}
	processed := 0
	var firstErr error
	for i := 0; i < concurrency; i++ {
		r := <-ch
		if r.ok {
			processed++
		}
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
	}
	return processed, firstErr
}

type embeddingWriter interface {
	UpsertEmbedding(ctx context.Context, memoryID, tenantID, subjectID string, values []float32) error
}

func (p *Processor) persistEmbedding(ctx context.Context, record memory.MemoryRecord) error {
	embedder := p.embedder
	if embedder == nil {
		embedder = embedding.Default()
	}
	values, err := embedder.Embed(ctx, record.Content)
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return nil
	}
	rec := embedding.RecordFromEmbedder(embedder, record.MemoryID, record.TenantID, record.SubjectID, values)
	if writer, ok := p.store.(interface {
		WriteEmbedding(context.Context, embedding.Record) error
	}); ok {
		return writer.WriteEmbedding(ctx, rec)
	}
	if writer, ok := p.store.(embeddingWriter); ok {
		return writer.UpsertEmbedding(ctx, record.MemoryID, record.TenantID, record.SubjectID, values)
	}
	return nil
}

func (p *Processor) PersistRuntime(ctx context.Context) {
	p.persistRuntime(ctx)
}

func (p *Processor) persistRuntime(ctx context.Context) {
	sink, ok := p.store.(interface {
		UpsertProviderRuntime(context.Context, string, []byte) error
	})
	if !ok {
		return
	}
	embedID := embedding.IdentityOf(p.embedder)
	embedStats := embedding.StatsOf(p.embedder)
	extractID := memory.ExtractorIdentityOf(p.extractor)
	extractStats := memory.ExtractorStatsOf(p.extractor)
	payload, err := json.Marshal(map[string]any{
		"embedder":            embedID,
		"embedder_stats":      embedStats,
		"embedder_signature":  embedID.Signature(),
		"embedder_fallbacks":  embedStats.Fallbacks,
		"extractor":           extractID,
		"extractor_stats":     extractStats,
		"extractor_signature": extractID.Signature(),
		"extractor_fallbacks": extractStats.Fallbacks,
	})
	if err != nil {
		return
	}
	_ = sink.UpsertProviderRuntime(ctx, "worker", payload)
}

func (p *Processor) persistEntityLinks(ctx context.Context, record memory.MemoryRecord) {
	if linker, ok := p.store.(memory.EntityLinker); ok {
		ents := memory.ExtractEntities(record.Content + " " + record.SourceText)
		if record.Metadata != nil {
			if raw, ok := record.Metadata["entities"]; ok {
				if list, ok := raw.([]string); ok && len(list) > 0 {
					ents = list
				}
			}
		}
		if len(ents) > 0 {
			_ = linker.LinkMemoryEntities(ctx, record.TenantID, record.SubjectID, record.MemoryID, ents)
		}
	}
	memory.PersistCanonicalEntity(ctx, p.store, record)
	memory.PersistDialogueAliases(ctx, p.store, record.TenantID, record.SubjectID, "", record.Content)

	// Mirror Service.persistEntityLinks: typed extract must land on atoms for
	// async LoCoMo/LME (default eval path), not only sync /ingest.
	if indexer, ok := p.store.(memory.AtomIndexer); ok {
		pred, _ := record.Explain["predicate"].(string)
		val, _ := record.Explain["value_norm"].(string)
		if pred == "" && record.Metadata != nil {
			if v, ok := record.Metadata["predicate"].(string); ok {
				pred = v
			}
			if v, ok := record.Metadata["value_norm"].(string); ok {
				val = v
			}
		}
		if pred != "" && val != "" {
			_ = indexer.UpsertMemoryAtom(ctx, record.TenantID, record.SubjectID, pred, val, record.MemoryID, record.ObservedAt)
		}
	}
	if rel, ok := memory.ProjectMemoryRelation(record); ok {
		if indexer, ok := p.store.(memory.RelationIndexer); ok {
			_ = indexer.UpsertMemoryRelation(ctx, rel)
		}
	}
}

func (p *Processor) persistEvidenceAndEvents(ctx context.Context, record memory.MemoryRecord) {
	if _, hasRaw := p.store.(memory.RawEvidenceWriter); !hasRaw {
		if writer, ok := p.store.(memory.EvidenceWriter); ok {
			srcType := "conversation"
			session := ""
			if record.Metadata != nil {
				if v, ok := record.Metadata["source_type"].(string); ok && v != "" {
					srcType = v
				}
				if v, ok := record.Metadata["session_id"].(string); ok {
					session = v
				}
			}
			_ = writer.ShadowWriteEvidence(ctx, record.TenantID, record.SubjectID, srcType, record.MemoryID, session, record.Content, record.MemoryID, record.ObservedAt, record.Metadata)
		}
	}
	if writer, ok := p.store.(memory.EventWriter); ok && record.Metadata != nil {
		pred, _ := record.Metadata["predicate"].(string)
		val, _ := record.Metadata["value_norm"].(string)
		if pred == "" {
			return
		}
		// Stateful current_state projection happens after supersede in
		// ProcessNext (guarded). Events still append here.
		if memory.IsStatefulPredicate(pred) && val != "" {
			return
		}
		if memory.PredicatePolicy(pred) == memory.PolicyAppendOnlyEvent || pred == memory.PredicateEvent || pred == memory.PredicateActivity {
			sum := sha256.Sum256([]byte(record.TenantID + pred + val + record.MemoryID))
			eventID := "evt_" + hex.EncodeToString(sum[:12])
			evidenceID := ""
			if record.Metadata != nil {
				if v, ok := record.Metadata["evidence_id"].(string); ok {
					evidenceID = v
				}
			}
			_ = writer.UpsertMemoryEvent(ctx, record.TenantID, record.SubjectID, eventID, pred, val, record.Content, record.MemoryID, evidenceID, record.ObservedAt, record.Confidence, memory.ExtractEntities(record.Content))
		}
	}
}
