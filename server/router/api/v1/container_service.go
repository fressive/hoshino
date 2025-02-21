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

// This file contains the implementation of the container service.

package v1

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"rina.icu/hoshino/server/context"
	"rina.icu/hoshino/store"
)

type (
	CreateContainerPayload struct {
		ChallengeUUID string `json:"challenge_uuid"`
	}
)

func CreateContainer(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	var payload CreateContainerPayload
	if err := c.Bind(&payload); err != nil {
		return Failed(&c, "Invalid payload")
	}

	challengeUUID := payload.ChallengeUUID

	challenge, err := ctx.Store.GetChallengeByUUID(challengeUUID)

	if err != nil {
		return Failed(&c, "Unable to fetch challenge")
	}

	if challenge.NoContainer ||
		challenge.State != store.ChallengeStateVisible ||

		(challenge.Expired() && challenge.AfterExpiredOperations&store.AfterExpireCreateContainer == 0) {
		return Failed(&c, "Unable to create container.")
	}

	flag := challenge.Flag
	if challenge.DynamicFlag {
		flag = fmt.Sprintf("%s{%s}", challenge.Game.FlagPrefix, uuid.New().String())
	}

	// TODO: create container here
	ctx.Store.CreateFlag(&store.Flag{
		Challenge: challenge,
		Flag:      flag,

		// TODO: set container
	})

	return OK(&c)
}
