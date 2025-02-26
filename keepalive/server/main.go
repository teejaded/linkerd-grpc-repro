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

// Binary server demonstrates how to enforce keepalive settings and manage idle
// connections to maintain active client connections.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	pb "google.golang.org/grpc/examples/features/proto/echo"
)

var port = flag.Int("port", 50052, "port number")
var sleep = flag.Duration("sleep", 30*time.Second, "Duration to sleep between before echo response")
var maxConnectionIdle = flag.Duration("max-connection-idle", 5*time.Second, "If a client is idle for this duration, send a GOAWAY")
var maxConnectionAge = flag.Duration("max-connection-age", 5*time.Second, "If any connection is alive for more than this duration, send a GOAWAY")
var maxConnectionAgeGrace = flag.Duration("max-connection-age-grace", 60*time.Second, "Allow this duration for pending RPCs to complete before forcibly closing connections")
var pingTime = flag.Duration("time", 60*time.Second, "Ping the client if it is idle for this duration to ensure the connection is still active")
var pingTimeout = flag.Duration("timeout", 10*time.Second, "Wait this duration for the ping ack before assuming the connection is dead")

var kaep = keepalive.EnforcementPolicy{
	MinTime:             10 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
	PermitWithoutStream: true,             // Allow pings even when there are no active streams
}

// server implements EchoServer.
type server struct {
	pb.UnimplementedEchoServer
}

func (s *server) UnaryEcho(_ context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	time.Sleep(*sleep)
	return &pb.EchoResponse{Message: req.Message}, nil
}

func main() {
	flag.Parse()

	address := fmt.Sprintf(":%v", *port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var kasp = keepalive.ServerParameters{
		MaxConnectionIdle:     *maxConnectionIdle,     // If a client is idle for X seconds, send a GOAWAY
		MaxConnectionAge:      *maxConnectionAge,      // If any connection is alive for more than X seconds, send a GOAWAY
		MaxConnectionAgeGrace: *maxConnectionAgeGrace, // Allow X seconds for pending RPCs to complete before forcibly closing connections
		Time:                  *pingTime,              // Ping the client if it is idle for X seconds to ensure the connection is still active
		Timeout:               *pingTimeout,           // Wait X second for the ping ack before assuming the connection is dead
	}

	s := grpc.NewServer(grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp))
	pb.RegisterEchoServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
