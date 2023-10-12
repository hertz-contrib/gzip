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
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/compress"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/stretchr/testify/assert"
)

const (
	testResponse = "Gzip Test Response"
)

func newServer() *route.Engine {
	router := route.NewEngine(config.NewOptions([]config.Option{}))
	router.Use(Gzip(DefaultCompression))
	router.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.Header("Content-Length", strconv.Itoa(len(testResponse)))
		c.String(200, testResponse)
	})
	return router
}

func TestGzip(t *testing.T) {
	request := ut.PerformRequest(newServer(), consts.MethodGet, "/", nil, ut.Header{
		Key: "Accept-Encoding", Value: "gzip",
	})
	w := request.Result()
	assert.Equal(t, w.StatusCode(), 200)
	assert.Equal(t, w.Header.Get("Vary"), "Accept-Encoding")
	assert.Equal(t, w.Header.Get("Content-Encoding"), "gzip")
	assert.NotEqual(t, w.Header.Get("Content-Length"), "0")
	assert.NotEqual(t, len(w.Body()), 19)
	assert.Equal(t, fmt.Sprint(len(w.Body())), w.Header.Get("Content-Length"))
}

func TestWildcard(t *testing.T) {
	request := ut.PerformRequest(newServer(), consts.MethodGet, "/", nil, ut.Header{
		Key: "Accept-Encoding", Value: "*",
	})
	w := request.Result()
	assert.Equal(t, w.StatusCode(), 200)
	assert.Equal(t, w.Header.Get("Vary"), "Accept-Encoding")
	assert.Equal(t, w.Header.Get("Content-Encoding"), "gzip")
	assert.NotEqual(t, w.Header.Get("Content-Length"), "0")
	assert.NotEqual(t, len(w.Body()), 19)
	assert.Equal(t, fmt.Sprint(len(w.Body())), w.Header.Get("Content-Length"))
}

func TestGzipPNG(t *testing.T) {
	router := route.NewEngine(config.NewOptions([]config.Option{}))
	router.Use(Gzip(DefaultCompression))
	router.GET("/image.png", func(ctx context.Context, c *app.RequestContext) {
		c.String(200, "this is a PNG!")
	})
	request := ut.PerformRequest(router, consts.MethodGet, "/image.png", nil, ut.Header{
		Key: "Accept-Encoding", Value: "gzip",
	})
	w := request.Result()
	assert.Equal(t, w.StatusCode(), 200)
	assert.Equal(t, w.Header.Get("Content-Encoding"), "")
	assert.Equal(t, w.Header.Get("Vary"), "")
	assert.Equal(t, string(w.Body()), "this is a PNG!")
}

func TestExcludedExtensions(t *testing.T) {
	router := route.NewEngine(config.NewOptions([]config.Option{}))
	router.Use(Gzip(DefaultCompression, WithExcludedExtensions([]string{".html"})))
	router.GET("/index.html", func(ctx context.Context, c *app.RequestContext) {
		c.String(200, "this is a HTML!")
	})
	request := ut.PerformRequest(router, consts.MethodGet, "/index.html", nil, ut.Header{
		Key: "Accept-Encoding", Value: "gzip",
	})
	w := request.Result()
	assert.Equal(t, http.StatusOK, w.StatusCode())
	assert.Equal(t, "", w.Header.Get("Content-Encoding"))
	assert.Equal(t, "", w.Header.Get("Vary"))
	assert.Equal(t, "this is a HTML!", string(w.Body()))
	assert.Equal(t, "15", w.Header.Get("Content-Length"))
}

func TestExcludedPaths(t *testing.T) {
	router := route.NewEngine(config.NewOptions([]config.Option{}))
	router.Use(Gzip(DefaultCompression, WithExcludedPaths([]string{"/api/"})))
	router.GET("/api/books", func(ctx context.Context, c *app.RequestContext) {
		c.String(200, "this is books!")
	})
	request := ut.PerformRequest(router, consts.MethodGet, "/api/books", nil, ut.Header{
		Key: "Accept-Encoding", Value: "gzip",
	})
	w := request.Result()
	assert.Equal(t, http.StatusOK, w.StatusCode())
	assert.Equal(t, "", w.Header.Get("Content-Encoding"))
	assert.Equal(t, "", w.Header.Get("Vary"))
	assert.Equal(t, "this is books!", string(w.Body()))
	assert.Equal(t, "14", w.Header.Get("Content-Length"))
}

func TestNoGzip(t *testing.T) {
	request := ut.PerformRequest(newServer(), consts.MethodGet, "/", nil)
	w := request.Result()
	assert.Equal(t, w.StatusCode(), 200)
	assert.Equal(t, w.Header.Get("Content-Encoding"), "")
	assert.Equal(t, w.Header.Get("Content-Length"), "18")
	assert.Equal(t, string(w.Body()), testResponse)
}

func TestDecompressGzip(t *testing.T) {
	buf := &bytes.Buffer{}
	gz := compress.AcquireStacklessGzipWriter(buf, gzip.DefaultCompression)
	if _, err := gz.Write([]byte(testResponse)); err != nil {
		gz.Close()
		t.Fatal(err)
	}
	gz.Close()
	router := route.NewEngine(config.NewOptions([]config.Option{}))
	router.Use(Gzip(DefaultCompression, WithDecompressFn(DefaultDecompressMiddleware)))
	router.POST("/", func(ctx context.Context, c *app.RequestContext) {
		if v := c.Request.Header.Get("Content-Encoding"); v != "" {
			t.Errorf("unexpected `Content-Encoding`: %s header", v)
		}
		if v := c.Request.Header.Get("Content-Length"); v != "" {
			t.Errorf("unexpected `Content-Length`: %s header", v)
		}
		data := c.GetRawData()
		c.Data(200, "text/plain", data)
	})
	request := ut.PerformRequest(router, consts.MethodPost, "/", &ut.Body{Body: buf, Len: buf.Len()}, ut.Header{
		Key: "Content-Encoding", Value: "gzip",
	})
	w := request.Result()
	assert.Equal(t, http.StatusOK, w.StatusCode())
	assert.Equal(t, "", w.Header.Get("Content-Encoding"))
	assert.Equal(t, "", w.Header.Get("Vary"))
	assert.Equal(t, testResponse, string(w.Body()))
	assert.Equal(t, "18", w.Header.Get("Content-Length"))
}

func TestDecompressGzipWithEmptyBody(t *testing.T) {
	router := route.NewEngine(config.NewOptions([]config.Option{}))
	router.Use(Gzip(DefaultCompression, WithDecompressFn(DefaultDecompressMiddleware)))
	router.POST("/", func(ctx context.Context, c *app.RequestContext) {
		c.String(200, "ok")
	})

	request := ut.PerformRequest(router, consts.MethodPost, "/", nil,
		ut.Header{Key: "Content-Encoding", Value: "gzip"})
	w := request.Result()
	assert.Equal(t, http.StatusOK, w.StatusCode())
	assert.Equal(t, "", w.Header.Get("Content-Encoding"))
	assert.Equal(t, "", w.Header.Get("Vary"))
	assert.Equal(t, "ok", string(w.Body()))
	assert.Equal(t, "2", w.Header.Get("Content-Length"))
}

func TestDecompressGzipWithSkipFunc(t *testing.T) {
	router := route.NewEngine(config.NewOptions([]config.Option{}))
	router.Use(Gzip(DefaultCompression, WithDecompressFn(DefaultDecompressMiddleware)))
	router.POST("/", func(ctx context.Context, c *app.RequestContext) {
		c.SetStatusCode(200)
	})

	request := ut.PerformRequest(router, consts.MethodPost, "/", nil,
		ut.Header{Key: "Accept-Encoding", Value: "gzip"})
	w := request.Result()
	assert.Equal(t, http.StatusOK, w.StatusCode())
	assert.Equal(t, "gzip", w.Header.Get("Content-Encoding"))
	assert.Equal(t, "Accept-Encoding", w.Header.Get("Vary"))
	assert.Equal(t, "", string(w.Body()))
	assert.Equal(t, "0", w.Header.Get("Content-Length"))
}

func TestDecompressGzipWithIncorrectData(t *testing.T) {
	router := route.NewEngine(config.NewOptions([]config.Option{}))
	router.Use(Gzip(DefaultCompression, WithDecompressFn(DefaultDecompressMiddleware)))
	router.POST("/", func(ctx context.Context, c *app.RequestContext) {
		c.String(200, "ok")
	})
	reader := bytes.NewReader([]byte(testResponse))
	request := ut.PerformRequest(router, consts.MethodPost, "/", &ut.Body{Body: reader, Len: reader.Len()},
		ut.Header{Key: "Content-Encoding", Value: "gzip"})
	w := request.Result()
	assert.Equal(t, http.StatusBadRequest, w.StatusCode())
}

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

	err = cli.Do(context.Background(), req, res)
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
	cli.Use(GzipForClient(DefaultCompression, WithExcludedExtensionsForClient([]string{".png"})))

	req := protocol.AcquireRequest()
	res := protocol.AcquireResponse()

	req.SetBodyString("bar")
	req.SetRequestURI("http://127.0.0.1:2334/image.png")

	err = cli.Do(context.Background(), req, res)
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
	cli.Use(GzipForClient(DefaultCompression, WithExcludedExtensionsForClient([]string{".html"})))

	req := protocol.AcquireRequest()
	res := protocol.AcquireResponse()

	req.SetBodyString("bar")
	req.SetRequestURI("http://127.0.0.1:3333/index.html")

	err = cli.Do(context.Background(), req, res)
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
	cli.Use(GzipForClient(DefaultCompression, WithExcludedPathsForClient([]string{"/api/"})))

	req := protocol.AcquireRequest()
	res := protocol.AcquireResponse()

	req.SetBodyString("bar")
	req.SetRequestURI("http://127.0.0.1:2336/api/books")

	err = cli.Do(context.Background(), req, res)
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

	err = cli.Do(context.Background(), req, res)
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
	h.Use(Gzip(DefaultCompression, WithDecompressFn(DefaultDecompressMiddleware)))
	go h.Spin()

	time.Sleep(time.Second)

	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(GzipForClient(DefaultCompression, WithDecompressFnForClient(DefaultDecompressMiddlewareForClient)))

	req := protocol.AcquireRequest()
	res := protocol.AcquireResponse()

	req.SetBodyString("bar")
	req.SetRequestURI("http://127.0.0.1:2338/")

	err = cli.Do(context.Background(), req, res)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	assert.Equal(t, res.StatusCode(), 200)
	assert.Equal(t, res.Header.Get("Content-Encoding"), "")
	assert.Equal(t, res.Header.Get("Vary"), "")
	assert.Equal(t, testResponse, string(res.Body()))
	assert.Equal(t, "18", res.Header.Get("Content-Length"))
}

func TestStreamGzip(t *testing.T) {
	data := `chunk 0: 
chunk 1: hi~
chunk 2: hi~hi~
chunk 3: hi~hi~hi~
chunk 4: hi~hi~hi~hi~
chunk 5: hi~hi~hi~hi~hi~
chunk 6: hi~hi~hi~hi~hi~hi~
chunk 7: hi~hi~hi~hi~hi~hi~hi~
chunk 8: hi~hi~hi~hi~hi~hi~hi~hi~
chunk 9: hi~hi~hi~hi~hi~hi~hi~hi~hi~
`
	h := server.Default(server.WithHostPorts("127.0.0.1:2339"))

	h.Use(GzipStream(DefaultCompression))
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		for i := 0; i < 10; i++ {
			c.Write([]byte(fmt.Sprintf("chunk %d: %s\n", i, strings.Repeat("hi~", i)))) // nolint: errcheck
			c.Flush()                                                                   // nolint: errcheck
			time.Sleep(200 * time.Millisecond)
		}
	})

	go h.Spin()

	time.Sleep(time.Second)

	c, _ := client.NewClient(client.WithResponseBodyStream(true))

	req := &protocol.Request{}
	resp := &protocol.Response{}
	defer func() {
		protocol.ReleaseRequest(req)
		protocol.ReleaseResponse(resp)
	}()

	req.SetMethod(consts.MethodGet)
	req.SetRequestURI("http://127.0.0.1:2339/")
	req.Header.Set("Accept-Encoding", "gzip")

	err := c.Do(context.Background(), req, resp)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	bodyStream := resp.BodyStream()
	compressedData, _ := ioutil.ReadAll(bodyStream)
	gunzipBytes, err := compress.AppendGunzipBytes(nil, compressedData)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))
	assert.Equal(t, "chunked", resp.Header.Get("Transfer-Encoding"))
	assert.Equal(t, "Accept-Encoding", resp.Header.Get("Vary"))
	assert.Equal(t, data, string(gunzipBytes))
}
