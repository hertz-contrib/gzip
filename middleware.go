// The MIT License (MIT)
//
// Copyright (c) 2016 Bo-Yi Wu
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// This file may have been modified by CloudWeGo authors. All CloudWeGo
// Modifications are Copyright 2022 CloudWeGo Authors.

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