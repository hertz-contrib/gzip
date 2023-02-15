package gzip

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/compress"
	"github.com/cloudwego/hertz/pkg/protocol"
)

type gzipMiddleware struct {
	*Options
	level int
}

func newGzipMiddleware(level int, opts ...Option) *gzipMiddleware {
	middleware := &gzipMiddleware{
		Options: DefaultOptions,
		level:   level,
	}
	for _, fn := range opts {
		fn(middleware.Options)
	}
	return middleware
}

func (g *gzipMiddleware) Middleware(next client.Endpoint) client.Endpoint {
	return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) (err error) {
		if fn := g.DecompressFnForClient; fn != nil && req.Header.Get("Content-Encoding") == "gzip" {
			fn(next)
		}
		if !g.shouldCompress(req) {
			return
		}
		req.SetHeader("Content-Encoding", "gzip")
		req.SetHeader("Vary", "Accept-Encoding")
		gzipBytes := compress.AppendGzipBytesLevel(nil, resp.Body(), g.level) // 拿到gzip数据
		resp.SetBodyStream(bytes.NewBuffer(gzipBytes), len(gzipBytes))
		return next(ctx, req, resp)
	}
}

func (g *gzipMiddleware) shouldCompress(req *protocol.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") ||
		strings.Contains(req.Header.Get("Connection"), "Upgrade") ||
		strings.Contains(req.Header.Get("Accept"), "text/event-stream") {
		return false
	}

	path := string(req.URI().RequestURI())

	extension := filepath.Ext(path)
	if g.ExcludedExtensions.Contains(extension) {
		return false
	}

	if g.ExcludedPaths.Contains(path) {
		return false
	}
	if g.ExcludedPathRegexes.Contains(path) {
		return false
	}

	return true
}
