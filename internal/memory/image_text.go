package memory

import (
	"bytes"
	"context"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	visibleTextMarker = "[visible text:"
	maxImageBytes     = 5 << 20
	maxOCRImages      = 3
)

var imageHTTPClient = &http.Client{
	Timeout: 8 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return http.ErrUseLastResponse
		}
		return nil
	},
}

// EnrichImageText appends OCR visible-text from message image URLs when the
// utterance refers to attached media without naming it. Idempotent.
func EnrichImageText(ctx context.Context, messages []Message) []Message {
	if len(messages) == 0 {
		return messages
	}
	if ctx == nil {
		ctx = context.Background()
	}
	out := make([]Message, len(messages))
	copy(out, messages)
	for i := range out {
		if !needsImageOCR(out[i]) {
			continue
		}
		var chunks []string
		n := 0
		for _, raw := range out[i].ImageURLs {
			if n >= maxOCRImages {
				break
			}
			text := visibleTextFromURL(ctx, raw)
			if text == "" {
				continue
			}
			chunks = append(chunks, text)
			n++
		}
		if len(chunks) == 0 {
			continue
		}
		joined := sanitizeVisibleText(strings.Join(chunks, "\n"))
		if joined == "" {
			continue
		}
		marker := visibleTextMarker + " " + joined + "]"
		if strings.TrimSpace(out[i].Content) == "" {
			out[i].Content = marker
			continue
		}
		out[i].Content = strings.TrimSpace(out[i].Content) + " " + marker
	}
	return out
}

func needsImageOCR(msg Message) bool {
	if len(msg.ImageURLs) == 0 {
		return false
	}
	if strings.Contains(msg.Content, visibleTextMarker) {
		return false
	}
	lower := strings.ToLower(msg.Content)
	return strings.Contains(lower, "this book") || strings.Contains(lower, "this novel") ||
		strings.Contains(lower, "this title")
}

func visibleTextFromURL(ctx context.Context, raw string) string {
	img, err := fetchImage(ctx, raw)
	if err != nil || len(img) < 64 {
		return ""
	}
	parts := make([]string, 0, 2)
	if crop := cropUpperCenterPNG(img); len(crop) > 64 {
		if text := ocrTesseract(ctx, crop, "11"); text != "" {
			parts = append(parts, text)
			if _, ok := titleFromVisibleText(text); ok {
				return strings.Join(parts, "\n")
			}
		}
	}
	if text := ocrTesseract(ctx, img, "11"); text != "" {
		parts = append(parts, text)
	}
	return strings.Join(parts, "\n")
}

func fetchImage(ctx context.Context, raw string) ([]byte, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return nil, io.EOF
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, io.EOF
	}
	if !imageURLHostAllowed(u.Hostname()) {
		return nil, io.EOF
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "brainy-ingest/1.0")
	resp, err := imageHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, io.EOF
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxImageBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxImageBytes {
		return nil, io.EOF
	}
	return data, nil
}

func imageURLHostAllowed(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "" {
		return false
	}
	loopbackOK := strings.TrimSpace(os.Getenv("BRAINY_IMAGE_FETCH_LOOPBACK")) != ""
	if h == "localhost" || h == "127.0.0.1" || h == "::1" {
		return loopbackOK
	}
	if h == "169.254.169.254" || strings.HasSuffix(h, ".local") {
		return false
	}
	ips, err := net.LookupIP(h)
	if err != nil || len(ips) == 0 {
		return false
	}
	for _, ip := range ips {
		if ip.IsLoopback() {
			if !loopbackOK {
				return false
			}
			continue
		}
		if ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return false
		}
	}
	return true
}

func cropUpperCenterPNG(img []byte) []byte {
	src, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		return nil
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 80 || h < 80 {
		return nil
	}
	x0 := b.Min.X + w*22/100
	y0 := b.Min.Y + h*10/100
	x1 := x0 + w*55/100
	y1 := y0 + h*42/100
	if x1 <= x0+20 || y1 <= y0+20 {
		return nil
	}
	rect := image.Rect(x0, y0, x1, y1)
	dst := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(dst, dst.Bounds(), src, rect.Min, draw.Src)
	scaled := scaleRGBA2x(dst)
	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return nil
	}
	return buf.Bytes()
}

func scaleRGBA2x(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewRGBA(image.Rect(0, 0, w*2, h*2))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := src.RGBAAt(b.Min.X+x, b.Min.Y+y)
			out.SetRGBA(x*2, y*2, c)
			out.SetRGBA(x*2+1, y*2, c)
			out.SetRGBA(x*2, y*2+1, c)
			out.SetRGBA(x*2+1, y*2+1, c)
		}
	}
	return out
}

func ocrTesseract(ctx context.Context, img []byte, psm string) string {
	bin, err := exec.LookPath("tesseract")
	if err != nil {
		return ""
	}
	dir, err := os.MkdirTemp("", "brainy-ocr-")
	if err != nil {
		return ""
	}
	defer os.RemoveAll(dir)
	ext := ".jpg"
	if len(img) >= 8 && string(img[:8]) == "\x89PNG\r\n\x1a\n" {
		ext = ".png"
	}
	path := filepath.Join(dir, "img"+ext)
	if err := os.WriteFile(path, img, 0o600); err != nil {
		return ""
	}
	ocrCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if psm == "" {
		psm = "11"
	}
	cmd := exec.CommandContext(ocrCtx, bin, path, "stdout", "--psm", psm)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return sanitizeVisibleText(string(out))
}

func sanitizeVisibleText(s string) string {
	s = strings.ReplaceAll(s, "]", " ")
	s = strings.ReplaceAll(s, "[", " ")
	s = whitespaceRE.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	if len(s) > 800 {
		s = strings.TrimSpace(s[:800])
	}
	return s
}
