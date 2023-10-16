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

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/compress"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/gzip"
)

func main() {
	h := server.Default(server.WithHostPorts(":8081"))

	// Note: Using this middleware will hijack the response writer and may have an impact on other interfaces.
	// Therefore, it is only necessary to use this middleware on interfaces with streaming gzip requirements.
	h.GET("/ping", gzip.GzipStream(gzip.DefaultCompression), func(ctx context.Context, c *app.RequestContext) {
		for i := 0; i < 10; i++ {
			c.Write([]byte(fmt.Sprintf("chunk %d: %s\n", i, strings.Repeat("hi~", i)))) // nolint: errcheck
			c.Flush()                                                                   // nolint: errcheck
			time.Sleep(time.Second)
		}
	})
	go h.Spin()

	cli, err := client.NewClient(client.WithResponseBodyStream(true))
	if err != nil {
		panic(err)
	}

	req := protocol.AcquireRequest()
	res := protocol.AcquireResponse()

	req.SetMethod(consts.MethodGet)
	req.SetRequestURI("http://localhost:8081/ping")
	req.Header.Set("Accept-Encoding", "gzip")

	if err = cli.Do(context.Background(), req, res); err != nil {
		panic(err)
	}

	bodyStream := res.BodyStream()

	// size after firstChunk compression
	firstChunk := make([]byte, 34)
	_, err = bodyStream.Read(firstChunk)
	if err != nil {
		panic(err)
	}
	firstChunkData, err := compress.AppendGunzipBytes(nil, firstChunk)
	fmt.Println(fmt.Printf("%s", firstChunkData))

	// size after secondChunk compression
	secondChunk := make([]byte, 71)
	_, err = bodyStream.Read(secondChunk)
	if err != nil {
		panic(err)
	}
	secondChunkData, err := compress.AppendGunzipBytes(nil, secondChunk)
	fmt.Println(fmt.Printf("%s", secondChunkData))

	otherChunks, _ := ioutil.ReadAll(bodyStream)
	otherChunksData, err := compress.AppendGunzipBytes(nil, otherChunks)
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Printf("%s", otherChunksData))
}
