package memory

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestEnrichImageTextNoopWithoutURLs(t *testing.T) {
	in := []Message{{Role: "user", Content: "Riley: hello"}}
	out := EnrichImageText(context.Background(), in)
	if len(out) != 1 || out[0].Content != "Riley: hello" {
		t.Fatalf("unexpected %#v", out)
	}
}

func TestTitleFromVisibleTextPrefersFunctionWordPhrase(t *testing.T) {
	got, ok := titleFromVisibleText("ANN AUTHOR THE QUIET ORCHARD FOREWORD BY PAT MAXIMIZE YOUR RESULTS")
	if !ok || !strings.Contains(strings.ToLower(got), "quiet orchard") {
		t.Fatalf("got %q ok=%v", got, ok)
	}
}

func TestNeedsImageOCRDeixis(t *testing.T) {
	if !needsImageOCR(Message{
		Content:   "Riley: This book I read last year reminds me to keep going.",
		ImageURLs: []string{"https://example.com/cover.jpg"},
	}) {
		t.Fatal("expected OCR for deictic book + image")
	}
	if needsImageOCR(Message{
		Content:   `Riley: I loved reading "The Little Prince" as a kid.`,
		ImageURLs: []string{"https://example.com/cover.jpg"},
	}) {
		t.Fatal("quoted title should skip OCR")
	}
	if needsImageOCR(Message{Content: "Riley: pottery class", ImageURLs: []string{"https://example.com/bowl.jpg"}}) {
		t.Fatal("unrelated image should skip OCR")
	}
}

func TestFixtureCoverOCR(t *testing.T) {
	path := strings.TrimSpace(os.Getenv("BRAINY_OCR_JPEG"))
	if path == "" {
		t.Skip("set BRAINY_OCR_JPEG to run cover OCR")
	}
	if _, err := exec.LookPath("tesseract"); err != nil {
		t.Skip("tesseract not installed")
	}
	img, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	crop := cropUpperCenterPNG(img)
	got := ocrTesseract(context.Background(), crop, "11") + " " + ocrTesseract(context.Background(), img, "11")
	title, ok := titleFromVisibleText(got)
	if !ok || !strings.Contains(strings.ToLower(title), "impossible") {
		t.Fatalf("ocr=%q title=%q ok=%v", got, title, ok)
	}
}
