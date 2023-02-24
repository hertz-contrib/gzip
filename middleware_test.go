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
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/stretchr/testify/assert"
)

func TestGzipForClient(t *testing.T) {
	h := server.Default(server.WithHostPorts("127.0.0.1:2333"))

	h.GET("/ping", func(ctx context.Context, c *app.RequestContext) {
		c.Header("Content-Length", strconv.Itoa(len(testResponse)))
		c.String(200, testResponse)
	})
	go h.Spin()
	time.Sleep(time.Second)

	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(GzipForClient(DefaultCompression))

	req := protocol.AcquireRequest()
	res := protocol.AcquireResponse()

	req.SetBodyString("bar")
	req.SetRequestURI("http://127.0.0.1:2333/ping")

	cli.Do(context.Background(), req, res)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	assert.Equal(t, res.StatusCode(), 200)
	assert.Equal(t, req.Header.Get("Vary"), "Accept-Encoding")
	assert.Equal(t, req.Header.Get("Content-Encoding"), "gzip")
	assert.NotEqual(t, req.Header.Get("Content-Length"), "0")
	assert.NotEqual(t, fmt.Sprint(len(req.Body())), req.Header.Get("Content-Length"))
}

func TestGzipPNGForClient(t *testing.T) {
	h := server.Default(server.WithHostPorts("127.0.0.1:2334"))

	h.GET("/image.png", func(ctx context.Context, c *app.RequestContext) {
		c.Header("Content-Length", strconv.Itoa(len(testResponse)))
		c.String(200, testResponse)
	})
	go h.Spin()
	time.Sleep(time.Second)

	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(GzipForClient(DefaultCompression, WithExcludedExtensions([]string{".png"})))

	req := protocol.AcquireRequest()
	res := protocol.AcquireResponse()

	req.SetBodyString("bar")
	req.SetRequestURI("http://127.0.0.1:2334/image.png")

	cli.Do(context.Background(), req, res)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	assert.Equal(t, res.StatusCode(), 200)
	assert.Equal(t, req.Header.Get("Vary"), "")
	assert.Equal(t, req.Header.Get("Content-Encoding"), "")
}

func TestExcludedExtensionsForClient(t *testing.T) {
	h := server.Default(server.WithHostPorts("127.0.0.1:3333"))

	h.GET("/index.html", func(ctx context.Context, c *app.RequestContext) {
		c.Header("Content-Length", strconv.Itoa(len(testResponse)))
		c.String(200, testResponse)
	})
	go h.Spin()
	time.Sleep(time.Second)

	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(GzipForClient(DefaultCompression, WithExcludedExtensions([]string{".html"})))

	req := protocol.AcquireRequest()
	res := protocol.AcquireResponse()

	req.SetBodyString("bar")
	req.SetRequestURI("http://127.0.0.1:3333/index.html")

	cli.Do(context.Background(), req, res)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	assert.Equal(t, res.StatusCode(), 200)
	assert.Equal(t, req.Header.Get("Vary"), "")
	assert.Equal(t, req.Header.Get("Content-Encoding"), "")
}

func TestExcludedPathsForClient(t *testing.T) {
	h := server.Default(server.WithHostPorts("127.0.0.1:2336"))

	h.GET("/api/books", func(ctx context.Context, c *app.RequestContext) {
		c.Header("Content-Length", strconv.Itoa(len(testResponse)))
		c.String(200, testResponse)
	})
	go h.Spin()
	time.Sleep(time.Second)

	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(GzipForClient(DefaultCompression, WithExcludedPaths([]string{"/api/"})))

	req := protocol.AcquireRequest()
	res := protocol.AcquireResponse()

	req.SetBodyString("bar")
	req.SetRequestURI("http://127.0.0.1:2336/api/books")

	cli.Do(context.Background(), req, res)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	assert.Equal(t, res.StatusCode(), 200)
	assert.Equal(t, req.Header.Get("Vary"), "")
	assert.Equal(t, req.Header.Get("Content-Encoding"), "")
}

func TestNoGzipForClient(t *testing.T) {
	h := server.Default(server.WithHostPorts("127.0.0.1:2337"))

	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.Header("Content-Length", strconv.Itoa(len(testResponse)))
		c.String(200, testResponse)
	})
	go h.Spin()

	time.Sleep(time.Second)

	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	req := protocol.AcquireRequest()
	res := protocol.AcquireResponse()

	req.SetBodyString("bar")
	req.SetRequestURI("http://127.0.0.1:2337/")

	cli.Do(context.Background(), req, res)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	assert.Equal(t, res.StatusCode(), 200)
	assert.Equal(t, req.Header.Get("Content-Encoding"), "")
	assert.Equal(t, req.Header.Get("Content-Length"), "3")
}

func TestDecompressGzipForClient(t *testing.T) {
	h := server.Default(server.WithHostPorts("127.0.0.1:2338"))

	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.Header("Content-Length", strconv.Itoa(len(testResponse)))
		c.String(200, testResponse)
	})
	h.Use(Gzip(DefaultCompression, WithDecompressFn(DefaultDecompressHandle)))
	go h.Spin()

	time.Sleep(time.Second)

	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(GzipForClient(DefaultCompression, WithDecompressFnForClient(DefaultDecompressFn4Client)))

	req := protocol.AcquireRequest()
	res := protocol.AcquireResponse()

	req.SetBodyString("bar")
	req.SetRequestURI("http://127.0.0.1:2338/")

	cli.Do(context.Background(), req, res)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	assert.Equal(t, res.StatusCode(), 200)
	assert.Equal(t, res.Header.Get("Content-Encoding"), "")
	assert.Equal(t, res.Header.Get("Vary"), "")
	assert.Equal(t, testResponse, string(res.Body()))
	assert.Equal(t, "18", res.Header.Get("Content-Length"))
}
