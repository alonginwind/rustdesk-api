package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// gzipMinLen is the smallest body worth compressing; tiny payloads burn CPU for
// little or negative gain.
const gzipMinLen = 1024

// gzipPool recycles gzip.Writer instances. Each deflater allocates a large
// internal window/buffer, so pooling keeps the GC load down under high QPS.
var gzipPool = sync.Pool{
	New: func() any {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return w
	},
}

// gzipWriter wraps gin.ResponseWriter and deflates the body on the fly. Only
// Write/WriteString/Size/Written/Flush/WriteHeaderNow/Unwrap are overridden;
// the promoted methods keep operating on the underlying writer.
type gzipWriter struct {
	gin.ResponseWriter
	gz      *gzip.Writer
	decided bool
	rawSize int    // bytes written by the handler, before compression
	buf     []byte // first-chunk buffer, used only while undecided
}

// compressible reports whether a content type benefits from gzip. Unknown or
// unset types are left alone, and already-compressed binaries (images, fonts,
// media) are skipped.
func compressible(contentType string) bool {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	// SSE relies on unbuffered delivery; gzip buffering would break realtime events.
	if ct == "text/event-stream" {
		return false
	}
	if strings.HasPrefix(ct, "text/") {
		return true
	}
	switch ct {
	case "application/javascript", "application/x-javascript", "application/json",
		"application/xml", "application/rss+xml", "application/atom+xml",
		"application/wasm", "application/manifest+json", "application/ld+json":
		return true
	}
	return false
}

// decide runs once, before the body is flushed, and arms compression only for a
// 200 response with a compressible body that is not already encoded. firstLen
// is the size of the initial write; when no Content-Length is declared (gin's
// JSON/String renderers never set it) a small first chunk means compression is
// skipped, so tiny payloads are not deflated.
func (g *gzipWriter) decide(status, firstLen int) {
	g.decided = true
	if g.ResponseWriter.Written() {
		return
	}
	if status != http.StatusOK {
		return
	}
	h := g.ResponseWriter.Header()
	if h.Get("Content-Encoding") != "" {
		return
	}
	if !compressible(h.Get("Content-Type")) {
		return
	}
	if cl := h.Get("Content-Length"); cl != "" {
		n, err := strconv.Atoi(cl)
		if err != nil {
			// Malformed Content-Length: be conservative and skip compression.
			return
		}
		if n < gzipMinLen {
			return
		}
	} else if firstLen < gzipMinLen {
		return
	}
	// Create the gzip writer first; only arm compression on success to avoid
	// advertising Content-Encoding: gzip while emitting a plain body.
	gz, _ := gzipPool.Get().(*gzip.Writer)
	if gz == nil {
		return
	}
	gz.Reset(g.ResponseWriter)
	h.Set("Content-Encoding", "gzip")
	// Add, not Set: an upstream Vary (e.g. Origin from CORS) must be preserved.
	// Check every value, not just the first, to avoid duplicate entries.
	hasVary := false
	for _, v := range h.Values("Vary") {
		if strings.Contains(strings.ToLower(v), "accept-encoding") {
			hasVary = true
			break
		}
	}
	if !hasVary {
		h.Add("Vary", "Accept-Encoding")
	}
	// The compressed length differs, so the plain Content-Length must go.
	h.Del("Content-Length")
	g.gz = gz
}

// WriteHeaderNow flushes the pending first chunk before the headers go out, so
// nothing is lost when something other than Write triggers the response.
func (g *gzipWriter) WriteHeaderNow() {
	g.flushBuf()
	g.ResponseWriter.WriteHeaderNow()
}

// flushBuf writes out the buffered first chunk. When compression has been armed
// the chunk goes through the gzip writer; otherwise it is passed through as-is.
// A nil buf means the current write is being handled directly by the caller.
func (g *gzipWriter) flushBuf() {
	if g.buf == nil {
		return
	}
	b := g.buf
	g.buf = nil
	if g.gz != nil {
		if _, err := g.gz.Write(b); err == nil {
			g.rawSize += len(b)
		}
		return
	}
	_, _ = g.ResponseWriter.Write(b)
}

func (g *gzipWriter) write(b []byte) (int, error) {
	n := len(b)
	if !g.decided {
		g.decide(g.ResponseWriter.Status(), n)
		if g.gz == nil {
			// Not compressing: pass through directly. buf is only ever filled
			// on the compressed path below, so there is nothing to flush here.
			return g.ResponseWriter.Write(b)
		}
		if n < gzipMinLen {
			// Compression armed but the first chunk is small: buffer it; the
			// rest of the body (if any) streams straight into the deflater.
			g.buf = append(g.buf[:0], b...)
			return n, nil
		}
	}
	if g.buf != nil {
		g.flushBuf()
	}
	if g.gz != nil {
		if _, err := g.gz.Write(b); err != nil {
			return 0, err
		}
		g.rawSize += n
		return n, nil
	}
	return g.ResponseWriter.Write(b)
}

func (g *gzipWriter) Write(b []byte) (int, error) {
	return g.write(b)
}

// WriteString must be overridden too: gin's render.WriteString would otherwise
// bypass the gzip.Writer and emit a raw body under a gzip Content-Encoding.
func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.write([]byte(s))
}

// Size reports the uncompressed bytes written by the handler, keeping the
// gin.ResponseWriter contract meaningful for downstream logging/audit
// middleware. Before the first write it defers to the underlying writer (-1).
func (g *gzipWriter) Size() int {
	if g.rawSize == 0 && g.buf == nil {
		return g.ResponseWriter.Size()
	}
	return g.rawSize + len(g.buf)
}

// Written reports whether any handler-side bytes have been produced, matching
// the semantics of Size.
func (g *gzipWriter) Written() bool {
	return g.rawSize > 0 || g.buf != nil || g.ResponseWriter.Written()
}

func (g *gzipWriter) Flush() {
	g.flushBuf()
	if g.gz != nil {
		_ = g.gz.Flush()
	}
	g.ResponseWriter.Flush()
}

// Unwrap exposes the underlying ResponseWriter. It is the Go 1.20+ convention
// for writer wrappers, letting http.ResponseController (and thus SSE / streaming
// handlers that type-assert http.Flusher or http.Pusher) reach the real writer.
func (g *gzipWriter) Unwrap() http.ResponseWriter {
	return g.ResponseWriter
}

// Gzip deflates responses for clients that accept it. It is a stdlib-only
// replacement for gin-contrib/gzip, so no extra dependency is pulled in.
// Register it as the innermost middleware so error paths (429/500) never pay
// the wrapping cost.
func Gzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip when the client does not accept gzip, for Range requests (a 206
		// body must not be re-encoded), and for WebSocket upgrades (hijacked).
		if !strings.Contains(strings.ToLower(c.GetHeader("Accept-Encoding")), "gzip") ||
			c.GetHeader("Range") != "" ||
			strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
			c.Next()
			return
		}
		gw := &gzipWriter{ResponseWriter: c.Writer}
		c.Writer = gw
		// Closing via defer guarantees the gzip trailer (CRC32/ISIZE) is
		// emitted even when a downstream handler panics; without it clients
		// would see a truncated stream and report unexpected EOF.
		defer func() {
			// Flush the pending first chunk before closing, otherwise a body
			// that fits entirely in buf would be lost (empty gzip stream) when
			// the handler returns or panics without further writes.
			gw.flushBuf()
			if gw.gz != nil {
				_ = gw.gz.Close()
				gzipPool.Put(gw.gz)
				gw.gz = nil
			}
			c.Writer = gw.ResponseWriter
		}()
		c.Next()
	}
}
