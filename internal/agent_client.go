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

package internal

import (
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"rina.icu/hoshino/internal/rpc"
	"rina.icu/hoshino/proto"
	"rina.icu/hoshino/server/config"
)

type AgentClient struct {
	Config *config.Agent
	Client proto.AgentClient
}

func NewClient(conf *config.Agent) *AgentClient {
	cert, err := credentials.NewClientTLSFromFile(conf.CertFile, conf.Name)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create client credentials for server %s:%d", conf.Address, conf.Port))
		panic(err)
	}

	auth := rpc.TokenAuth{
		TokenValidParam: proto.TokenValidParam{
			Token: conf.Token,
		},
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(cert),
		grpc.WithPerRPCCredentials(&auth),
	}

	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", conf.Address, conf.Port), opts...)

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create agent client for server %s:%d", conf.Address, conf.Port))
		panic(err)
	}

	c := proto.NewAgentClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	r, err := c.Ping(ctx, &proto.PingRequest{
		Message: "hoshino-agent",
	})

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to ping agent server[%s:%d].", conf.Address, conf.Port))
		panic(err)
	}

	if r.Message != "hoshino-agent" {
		slog.Error(fmt.Sprintf("The agent server[%s:%d] answered incorrectly.", conf.Address, conf.Port))
		return nil
	}

	return &AgentClient{
		Config: conf,
		Client: c,
	}
}

func (c *AgentClient) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Client.Ping(ctx, &proto.PingRequest{
		Message: "hoshino-agent",
	})

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to ping agent server[%s:%d].", c.Config.Address, c.Config.Port))
		panic(err)
	}

	if r.Message != "hoshino-agent" {
		slog.Error(fmt.Sprintf("The agent server[%s:%d] answered incorrectly.", c.Config.Address, c.Config.Port))
	}
}
