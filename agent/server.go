// Copyright 2025 Rina
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agent

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"rina.icu/hoshino/internal/rpc"
	"rina.icu/hoshino/proto"
)

type Agent struct {
	proto.UnimplementedAgentServer
	rpc.TokenAuth
}

func (*Agent) Ping(ctx context.Context, in *proto.PingRequest) (*proto.PongReply, error) {
	slog.Info(fmt.Sprintf("Received ping request: %v", in.Message))
	return &proto.PongReply{
		Message: in.Message,
	}, nil
}

func RunAgentServer(c *Config) {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", c.Addr, c.Port))

	if err != nil {
		panic(err)
	}

	var opts []grpc.ServerOption
	opts = append(opts, grpc.UnaryInterceptor(ServerInterceptorCheckToken(c)))
	slog.Info(fmt.Sprintf("%s %s", c.CertFile, c.PrivateKeyFile))
	cert, err := credentials.NewServerTLSFromFile(c.CertFile, c.PrivateKeyFile)
	if err != nil {
		panic(err)
	}
	opts = append(opts, grpc.Creds(cert))

	server := grpc.NewServer(opts...)
	proto.RegisterAgentServer(server, &Agent{})

	server.Serve(listen)
}
