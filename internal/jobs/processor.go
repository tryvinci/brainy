package jobs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

	extracted, err := p.extractor.Extract(ctx, job.Request)
	if err != nil {
		// Fail before any upserts so a provider error cannot leave partial
		// enrichment or mutate the immutable raw_ingests payload.
		_ = p.store.FailExtractionJob(ctx, job.JobID, job.IngestID, err.Error())
		return true, err
	}

	for _, item := range extracted {
		p.metrics.RecordExtraction()
		record, err := memory.BuildMemoryRecord(p.id("mem"), p.now(), job.Request, item, p.packs)
		if err != nil {
			_ = p.store.FailExtractionJob(ctx, job.JobID, job.IngestID, err.Error())
			return true, err
		}
		upserted, err := p.store.UpsertMemory(ctx, record)
		if err != nil {
			_ = p.store.FailExtractionJob(ctx, job.JobID, job.IngestID, err.Error())
			return true, err
		}
		p.persistEmbedding(ctx, upserted.Record)
		p.persistEntityLinks(ctx, upserted.Record)
		p.persistEvidenceAndEvents(ctx, upserted.Record)
		// Match sync ingest: supersede older state, then project current_state
		// only when shouldReplaceCurrentState allows (no blind late-older wins).
		_ = memory.AutoSupersedePriorState(ctx, p.store, upserted.Record)
		fresh := upserted.Record
		if got, err := p.store.GetMemory(ctx, fresh.TenantID, fresh.SubjectID, fresh.MemoryID); err == nil {
			fresh = got
		}
		memory.ProjectCurrentStateIfApplicable(ctx, p.store, fresh)
	}

	if err := p.store.CompleteExtractionJob(ctx, job.JobID, job.IngestID); err != nil {
		return true, err
	}
	return true, nil
}

type embeddingWriter interface {
	UpsertEmbedding(ctx context.Context, memoryID, tenantID, subjectID string, values []float32) error
}

func (p *Processor) persistEmbedding(ctx context.Context, record memory.MemoryRecord) {
	writer, ok := p.store.(embeddingWriter)
	if !ok {
		return
	}
	embedder := p.embedder
	if embedder == nil {
		embedder = embedding.Default()
	}
	values, err := embedder.Embed(ctx, record.Content)
	if err != nil || len(values) == 0 {
		return
	}
	_ = writer.UpsertEmbedding(ctx, record.MemoryID, record.TenantID, record.SubjectID, values)
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
			_ = writer.UpsertMemoryEvent(ctx, record.TenantID, record.SubjectID, eventID, pred, val, record.Content, record.MemoryID, "", record.ObservedAt, record.Confidence, memory.ExtractEntities(record.Content))
		}
	}
}
