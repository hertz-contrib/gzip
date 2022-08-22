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
	"net/http"
	"regexp"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/compress"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var (
	DefaultExcludedExtensions = NewExcludedExtensions([]string{
		".png", ".gif", ".jpeg", ".jpg",
	})
	DefaultOptions = &Options{
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
	Option              func(*Options)
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
	if c.Request.Body() == nil {
		return
	}
	r, err := compress.AcquireGzipReader(bytes.NewReader(c.Request.Body()))
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(r)
	if err != nil {
		hlog.CtxErrorf(ctx, "buffer read error: %v", err.Error())
	}
	c.Request.Header.DelBytes([]byte("Content-Encoding"))
	c.Request.Header.DelBytes([]byte("Content-Length"))
	c.Request.SetBody(buf.Bytes())
}
