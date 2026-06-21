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
	reg, _ := pack.LoadRegistryFromDir("packs")
	return &Processor{
		store:     store,
		extractor: memory.NewExtractor(),
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

	for _, extracted := range p.extractor.Extract(job.Request) {
		p.metrics.RecordExtraction()
		now := p.now()
		record := memory.MemoryRecord{
			MemoryID:          p.id("mem"),
			TenantID:          job.Request.TenantID,
			SubjectID:         job.Request.SubjectID,
			Kind:              extracted.Kind,
			Content:           extracted.Content,
			SourceText:        extracted.SourceText,
			SourceType:        job.Request.SourceType,
			DedupeKey:         memory.DedupeKey(job.Request.TenantID, job.Request.SubjectID, extracted.Kind, extracted.Content),
			Status:            memory.StatusActive,
			Confidence:        extracted.Confidence,
			ExtractionVersion: "deterministic-v1",
			Explain:           extracted.Explain,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		memory.ApplyVerticalPack(&record, job.Request.Vertical, extracted.Kind, extracted.Content, p.packs)
		if _, err := p.store.UpsertMemory(ctx, record); err != nil {
			_ = p.store.FailExtractionJob(ctx, job.JobID, job.IngestID, err.Error())
			return true, err
		}
	}

	if err := p.store.CompleteExtractionJob(ctx, job.JobID, job.IngestID); err != nil {
		return true, err
	}
	return true, nil
}
