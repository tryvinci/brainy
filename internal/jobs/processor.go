package jobs

import (
	"context"
	"fmt"
	"time"

	"brainy/internal/memory"
	"brainy/internal/observability"
	"brainy/internal/pack"
)

type Processor struct {
	store     memory.Store
	extractor memory.Extractor
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
		packs:     reg,
		metrics:   metrics,
		now:       time.Now().UTC,
		id: func(prefix string) string {
			return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
		},
	}
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
	// Local hash embedder — same path as sync Service.
	_ = writer.UpsertEmbedding(ctx, record.MemoryID, record.TenantID, record.SubjectID, nil)
}
