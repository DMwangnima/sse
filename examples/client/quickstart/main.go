/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * The MIT License (MIT)
 *
 * Copyright (c) 2014 Manuel Martínez-Almeida
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/cloudwego/hertz/pkg/protocol"

	"github.com/hertz-contrib/sse"
)

func main() {
	hertzCli, err := client.NewClient(client.WithDialer(standard.NewDialer()))
	if err != nil {
		hlog.Errorf("create Hertz Client failed, err: %v", err)
		return
	}
	c, err := sse.NewClientWithOptions(sse.WithHertzClient(hertzCli))
	if err != nil {
		hlog.Errorf("create SSE Client failed, err: %v", err)
		return
	}

	// touch off when connected to the server
	c.SetOnConnectCallback(func(ctx context.Context, client *sse.Client) {
		hlog.Infof("client1 connect to server %s success with %s method", c.GetURL(), c.GetMethod())
	})

	// touch off when the connection is shutdown
	c.SetDisconnectCallback(func(ctx context.Context, client *sse.Client) {
		hlog.Infof("client1 disconnect to server %s success with %s method", c.GetURL(), c.GetMethod())
	})

	events := make(chan *sse.Event)
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		req := &protocol.Request{}
		req.SetRequestURI("http://127.0.0.1:8888/sse")
		cErr := c.SubscribeWithContext(ctx, func(msg *sse.Event) {
			if msg.Data != nil {
				events <- msg
				return
			}
		}, sse.WithRequest(req))
		errChan <- cErr
	}()
	go func() {
		time.Sleep(5 * time.Second)
		cancel()
		fmt.Println("client1 subscribe cancel")
	}()
	for {
		select {
		case e := <-events:
			hlog.Info(e)
		case err := <-errChan:
			hlog.CtxErrorf(context.Background(), "err = %s", err.Error())
			return
		}
	}
}
