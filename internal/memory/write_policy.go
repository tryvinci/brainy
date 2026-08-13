package memory

import "strings"

// WriteMutationMode is the ingest write policy for provider extract ops.
// Conversational/core memory is append-only (Mem0 ADAPT): NONE/UPDATE/DELETE
// must not drop or mutate prior records. Governed verticals keep #94 ops.
type WriteMutationMode string

const (
	WriteModeAppendOnly WriteMutationMode = "append_only"
	WriteModeGoverned   WriteMutationMode = "governed"
)

// WriteMutationModeOf returns append-only for core/empty vertical and governed
// for any non-core vertical (marketing, support, and other pack-backed writes).
func WriteMutationModeOf(req IngestRequest) WriteMutationMode {
	vertical := strings.TrimSpace(req.Vertical)
	if vertical != "" && !strings.EqualFold(vertical, VerticalCore) {
		return WriteModeGoverned
	}
	return WriteModeAppendOnly
}

func rewriteProviderEventAsAdd(extracted *ExtractedMemory) {
	if extracted.Explain == nil {
		extracted.Explain = map[string]any{}
	}
	extracted.Explain["memory_event"] = MemoryEventAdd
	delete(extracted.Explain, "supersedes_memory_id")
}
