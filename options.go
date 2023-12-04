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
	"net/http"
	"regexp"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/compress"
	"github.com/cloudwego/hertz/pkg/protocol"
)

var (
	DefaultExcludedExtensions = NewExcludedExtensions([]string{
		".png", ".gif", ".jpeg", ".jpg",
	})
	DefaultOptions = &Options{
		ExcludedExtensions: DefaultExcludedExtensions,
	}
	DefaultClientExcludedExtensions = NewExcludedExtensions([]string{
		".png", ".gif", ".jpeg", ".jpg",
	})
	DefaultClientOptions = &ClientOptions{
		ExcludedExtensions: DefaultExcludedExtensions,
	}
)

type (
	Options struct {
		ExcludedExtensions  ExcludedExtensions
		ExcludedPaths       ExcludedPaths
		ExcludedPathRegexes ExcludedPathRegexes
		DecompressFn        app.HandlerFunc
	}
	ClientOptions struct {
		ExcludedExtensions    ExcludedExtensions
		ExcludedPaths         ExcludedPaths
		ExcludedPathRegexes   ExcludedPathRegexes
		DecompressFnForClient client.Middleware
	}
	Option       func(*Options)
	ClientOption func(*ClientOptions)

	ExcludedExtensions  map[string]bool
	ExcludedPaths       []string
	ExcludedPathRegexes []*regexp.Regexp
)

// WithExcludedExtensions customize excluded extensions
func WithExcludedExtensions(args []string) Option {
	return func(o *Options) {
		o.ExcludedExtensions = NewExcludedExtensions(args)
	}
}

// WithExcludedPathRegexes customize paths' regexes
func WithExcludedPathRegexes(args []string) Option {
	return func(o *Options) {
		o.ExcludedPathRegexes = NewExcludedPathRegexes(args)
	}
}

// WithExcludedPathsRegexs customize path's regexes
// NOTE: WithExcludedPathRegexs is exactly same as WithExcludedPathRegexes, this just for aligning with gin
func WithExcludedPathsRegexs(args []string) Option {
	return func(o *Options) {
		o.ExcludedPathRegexes = NewExcludedPathRegexes(args)
	}
}

func WithExcludedPaths(args []string) Option {
	return func(o *Options) {
		o.ExcludedPaths = NewExcludedPaths(args)
	}
}

func WithDecompressFn(decompressFn app.HandlerFunc) Option {
	return func(o *Options) {
		o.DecompressFn = decompressFn
	}
}

func WithDecompressFnForClient(decompressFnForClient client.Middleware) ClientOption {
	return func(o *ClientOptions) {
		o.DecompressFnForClient = decompressFnForClient
	}
}

// WithExcludedExtensionsForClient customize excluded extensions
func WithExcludedExtensionsForClient(args []string) ClientOption {
	return func(o *ClientOptions) {
		o.ExcludedExtensions = NewExcludedExtensions(args)
	}
}

// WithExcludedPathRegexesForClient customize paths' regexes
func WithExcludedPathRegexesForClient(args []string) ClientOption {
	return func(o *ClientOptions) {
		o.ExcludedPathRegexes = NewExcludedPathRegexes(args)
	}
}

func WithExcludedPathsForClient(args []string) ClientOption {
	return func(o *ClientOptions) {
		o.ExcludedPaths = NewExcludedPaths(args)
	}
}

func NewExcludedPaths(paths []string) ExcludedPaths {
	return ExcludedPaths(paths)
}

func NewExcludedExtensions(extensions []string) ExcludedExtensions {
	res := make(ExcludedExtensions)
	for _, e := range extensions {
		res[e] = true
	}
	return res
}

func NewExcludedPathRegexes(regexes []string) ExcludedPathRegexes {
	result := make([]*regexp.Regexp, len(regexes))
	for i, reg := range regexes {
		result[i] = regexp.MustCompile(reg)
	}
	return result
}

func (e ExcludedPathRegexes) Contains(requestURI string) bool {
	for _, reg := range e {
		if reg.MatchString(requestURI) {
			return true
		}
	}
	return false
}

func (e ExcludedExtensions) Contains(target string) bool {
	_, ok := e[target]
	return ok
}

func (e ExcludedPaths) Contains(requestURI string) bool {
	for _, path := range e {
		if strings.HasPrefix(requestURI, path) {
			return true
		}
	}
	return false
}

func DefaultDecompressHandle(ctx context.Context, c *app.RequestContext) {
	if len(c.Request.Body()) <= 0 {
		return
	}
	gunzipBytes, err := compress.AppendGunzipBytes(nil, c.Request.Body())
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.Request.Header.DelBytes([]byte("Content-Encoding"))
	c.Request.Header.DelBytes([]byte("Content-Length"))
	c.Request.SetBody(gunzipBytes)
}

func DefaultDecompressMiddlewareForClient(next client.Endpoint) client.Endpoint {
	return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) (err error) {
		if len(resp.Body()) <= 0 {
			return
		}
		gunzipBytes, err := compress.AppendGunzipBytes(nil, resp.Body())
		if err != nil {
			return err
		}
		resp.Header.DelBytes([]byte("Content-Encoding"))
		resp.Header.DelBytes([]byte("Content-Length"))
		resp.SetBody(gunzipBytes)
		return next(ctx, req, resp)
	}
}
