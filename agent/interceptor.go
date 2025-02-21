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

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"rina.icu/hoshino/internal/util"
)

func ServerInterceptorCheckToken(c *Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		err := CheckToken(ctx, c)
		if err != nil {
			slog.Warn(fmt.Sprintf("Agent check token failed: %s", err.Error()))
			return nil, err
		}

		return handler(ctx, req)
	}
}

func CheckToken(ctx context.Context, c *Config) error {
	data, b := metadata.FromIncomingContext(ctx)
	if !b {
		return status.Error(codes.InvalidArgument, "Agent check token failed. No metadata found.")
	}

	var token string
	tokenInfo, err := data["token"]
	if !err || len(tokenInfo) == 0 {
		return status.Error(codes.InvalidArgument, "Agent check token failed. No token found.")
	}

	token = util.SHA256(tokenInfo[0])
	if token != c.Secret {
		return status.Error(codes.PermissionDenied, "Agent check token failed. Token is invalid.")
	}

	return nil
}
