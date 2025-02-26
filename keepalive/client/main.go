/*
 *
 * Copyright 2019 gRPC authors.
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
 *
 */

// Binary client demonstrates how to configure keepalive pings to maintain
// connectivity and detect stale connections.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "google.golang.org/grpc/examples/features/proto/echo"
	"google.golang.org/grpc/keepalive"
)

var addr = flag.String("addr", "localhost:50052", "the address to connect to")
var idleThreads = flag.Int("idle-threads", 0, "count of idle threads")
var sleep = flag.Duration("sleep", 1, "seconds to sleep between launching threads")
var pingTime = flag.Duration("time", 60, "inactivity ping period")
var pingTimeout = flag.Duration("timeout", 10, "ping timeout")

func main() {
	flag.Parse()

	for range *idleThreads {
		time.Sleep(*sleep * time.Second)
		go idle_conn()
	}

	select {} // Block forever; run with GODEBUG=http2debug=2 to observe ping frames and GOAWAYs due to idleness.
}

func idle_conn() {
	var kacp = keepalive.ClientParameters{
		Time:                *pingTime * time.Second,    // send pings if there is no activity
		Timeout:             *pingTimeout * time.Second, // wait 10 second for ping ack before considering the connection dead
		PermitWithoutStream: true,                       // send pings even without active streams
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		log.Fatalf("idle_conn: did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewEchoClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("idle_conn: Performing unary request")
	res, err := c.UnaryEcho(ctx, &pb.EchoRequest{Message: "keepalive idle"})
	if err != nil {
		log.Fatalf("unexpected error from UnaryEcho: %v", err)
	}
	fmt.Println("idle_conn: RPC response:", res)
	fmt.Println("idle_conn: Going idle")
	select {} // Block forever
}
