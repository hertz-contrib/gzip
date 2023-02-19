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
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol"
)

func TestGzipForClient(t *testing.T) {
	h := server.Default(server.WithHostPorts(":2333"))
	h.Use(Gzip(gzip.DefaultCompression))
	h.GET("/ping", func(ctx context.Context, c *app.RequestContext) {
		c.String(http.StatusOK, "pong "+fmt.Sprint(time.Now().Unix()))
	})
	go h.Spin()
	time.Sleep(time.Second)

	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	/*	cli.Use(GzipForClient(DefaultCompression))*/
	cli.Use(GzipForClient(DefaultCompression, WithDecompressFnForClient(DefaultDecompressFn4Client)))
	req := protocol.AcquireRequest()
	res := protocol.AcquireResponse()
	req.SetRequestURI("http://localhost:2333/ping")
	cli.Do(context.Background(), req, res)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	/*
		if g, e := resp.StatusCode(), backendStatus; g != e {
			t.Errorf("got res.StatusCode %d; expected %d", g, e)
		}
	*/

}
