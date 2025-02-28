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
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"rina.icu/hoshino/server/context"
	"rina.icu/hoshino/store"
)

func CreateContainer(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	challenge, err := ctx.Store.GetChallengeByUUID(c.Param("challenge_uuid"))

	if err != nil {
		return Failed(&c, "Unable to fetch challenge")
	}

	if challenge.NoContainer ||
		challenge.State != store.ChallengeStateVisible ||

		(challenge.Expired() && challenge.AfterExpiredOperations&store.AfterExpireCreateContainer == 0) {
		return Failed(&c, "Unable to create container.")
	}

	user, _ := GetUserFromToken(&c)
	team := challenge.Game.GetTeamByUser(ctx.Store, user)

	if team == nil {
		return Failed(&c, "You are not in a team.")
	}

	flag := challenge.Flag
	if challenge.DynamicFlag {
		flag = fmt.Sprintf("%s{%s}", challenge.Game.FlagPrefix, uuid.New().String())
	}

	if !ctx.Store.CanCreateContainer(user) {
		return Failed(&c, "You have reached the maximum number of containers.")
	}

	// TODO: create container here
	slog.Info("Creating container for challenge %s", slog.Any("challenge", challenge))

	uuid, err := ctx.ContainerManager.CreateChallengeContainer(challenge, user, flag)
	containerModel := &store.Container{
		Creator:          user,
		Challenge:        challenge,
		UUID:             uuid,
		Status:           store.ContainerStatusRunning,
		ExpireTime:       time.Now().Unix() + cast.ToInt64(ctx.Store.GetSettingInt("container_expire_time")),
		LeftRenewalTimes: ctx.Store.GetSettingInt("max_container_renewal_times"),
	}

	flagModel := &store.Flag{
		Challenge: challenge,
		Flag:      flag,
		Container: containerModel,
		Team:      team,
	}

	ctx.Store.CreateContainer(containerModel)
	ctx.Store.CreateFlag(flagModel)

	if err != nil {
		return Failed(&c, "Unable to create container.")
	}

	return OKWithData(&c, map[string]interface{}{
		"uuid":     uuid,
		"entrance": fmt.Sprintf("%s.%s", uuid, ctx.Store.GetSettingString("node_domain")),
		"expire":   containerModel.ExpireTime,
	})
}

func DisposeContainer(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	challenge, err := ctx.Store.GetChallengeByUUID(c.Param("challenge_uuid"))

	if err != nil {
		return Failed(&c, "Unable to fetch challenge")
	}

	user, _ := GetUserFromToken(&c)
	container, err := ctx.Store.GetContainerByChallengeAndUser(challenge, user)

	if err != nil {
		return Failed(&c, "Unable to fetch container")
	}

	if container.CreatorID != user.ID {
		return PermissionDenied(&c)
	}

	if container.Status != store.ContainerStatusRunning {
		return Failed(&c, "Container is not running")
	}

	ctx.ContainerManager.DeleteChallengeContainer(challenge, user)
	container.Status = store.ContainerStatusStopped
	ctx.Store.UpdateContainer(container)

	return OK(&c)
}
