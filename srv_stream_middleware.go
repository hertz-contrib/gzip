/*
 * Copyright 2023 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * The MIT License (MIT)
 *
 * Copyright (c) 2016 Bo-Yi Wu
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
* This file may have been modified by CloudWeGo authors. All CloudWeGo
* Modifications are Copyright 2022 CloudWeGo Authors.
*/

package gzip

import (
	"context"
	"strings"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/compress"
	"github.com/cloudwego/hertz/pkg/network"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/http1/ext"
	"github.com/cloudwego/hertz/pkg/protocol/http1/resp"
)

type gzipChunkedWriter struct {
	sync.Once
	level          int
	originalSize   int
	compressedSize int
	wroteHeader    bool
	finalizeErr    error
	r              *protocol.Response
	w              network.Writer
}

func (g *gzipChunkedWriter) Write(p []byte) (n int, err error) {
	gzipBytes := compress.AppendGzipBytesLevel(nil, p, g.level)

	if !g.wroteHeader {
		g.r.Header.SetContentLength(-1)
		g.r.Header.Set("Content-Encoding", "gzip")
		g.r.Header.Set("Vary", "Accept-Encoding")
		if err = resp.WriteHeader(&g.r.Header, g.w); err != nil {
			return
		}
		g.wroteHeader = true
	}

	if err = ext.WriteChunk(g.w, gzipBytes, false); err != nil {
		return
	}

	g.originalSize += len(p)
	g.compressedSize += len(gzipBytes)

	return len(gzipBytes), nil
}

func (g *gzipChunkedWriter) Flush() error {
	return g.w.Flush()
}

func (g *gzipChunkedWriter) Finalize() error {
	g.Do(func() {
		// in case no actual data from user
		if !g.wroteHeader {
			g.r.Header.SetContentLength(-1)
			g.r.Header.Set("Content-Encoding", "gzip")
			g.r.Header.Set("Vary", "Accept-Encoding")
			if g.finalizeErr = resp.WriteHeader(&g.r.Header, g.w); g.finalizeErr != nil {
				return
			}
			g.wroteHeader = true
		}
		g.finalizeErr = ext.WriteChunk(g.w, nil, true)
		if g.finalizeErr != nil {
			return
		}
		g.finalizeErr = ext.WriteTrailer(g.r.Header.Trailer(), g.w)
	})
	return g.finalizeErr
}

func newGzipChunkedWriter(r *protocol.Response, w network.Writer, level int) network.ExtWriter {
	extWriter := new(gzipChunkedWriter)
	extWriter.r = r
	extWriter.w = w
	extWriter.Once = sync.Once{}
	extWriter.level = level
	return extWriter
}

func (g *gzipSrvMiddleware) SrvStreamMiddleware(ctx context.Context, c *app.RequestContext) {
	if fn := g.DecompressFn; fn != nil && strings.EqualFold(c.Request.Header.Get("Content-Encoding"), "gzip") {
		fn(ctx, c)
	}
	if !g.shouldCompress(&c.Request) {
		return
	}

	w := newGzipChunkedWriter(&c.Response, c.GetWriter(), g.level)
	c.Response.HijackWriter(w)

	c.Next(ctx)
}
