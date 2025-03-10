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
	"time"

	"github.com/labstack/echo/v4"
	"rina.icu/hoshino/internal/k8s"
	"rina.icu/hoshino/internal/util"
	"rina.icu/hoshino/server/context"
	"rina.icu/hoshino/store"
)

func CreateChallengeContainer(c echo.Context) error {
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

	// the container is still allowed to be created after the challenge is solved
	// if challenge.IsSolvedBy(team, ctx.Store) {
	// 	return Failed(&c, "You have solved this challenge.")
	// }

	flag := fmt.Sprintf("%s{%s}", challenge.Game.FlagPrefix, util.GenerateFlagContent(challenge.FlagFormat, user.UUID))

	if !ctx.Store.CanCreateContainer(user) {
		return Failed(&c, "You have reached the maximum number of containers.")
	}

	uuid, nodeDomain, err := ctx.ContainerManager.CreateChallengeContainer(challenge, user, flag, ctx.Store)
	containerModel := &store.Container{
		Creator:          user,
		Challenge:        challenge,
		UUID:             uuid,
		Status:           store.ContainerStatusRunning,
		ExpireTime:       time.Now().UnixMilli() + ctx.Store.GetSettingInt64("container_expire_time"),
		LeftRenewalTimes: ctx.Store.GetSettingInt("max_container_renewal_times"),
		Identifier:       challenge.UUID + user.UUID,
		NodeDomain:       nodeDomain,
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
		"uuid":               uuid,
		"entrance":           fmt.Sprintf("%s.%s", uuid, nodeDomain),
		"expire":             containerModel.ExpireTime,
		"left_renewal_times": containerModel.LeftRenewalTimes,
	})
}

func DisposeChallengeContainer(c echo.Context) error {
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

	ctx.ContainerManager.DisposeChallengeContainer(challenge, user)
	container.Status = store.ContainerStatusStopped
	ctx.Store.UpdateContainer(container)

	return OK(&c)
}

func CreateContainer(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	user, _ := GetUserFromToken(&c)
	game, _ := ctx.Store.GetGameByUUID(c.Param("game_uuid"))

	if user.Privilege < store.UserPrivilegeAdministrator && !game.IsManager(user) {
		return PermissionDenied(&c)
	}

	payload := new(k8s.ContainerCreatePayload)
	if err := c.Bind(payload); err != nil {
		return Failed(&c, "Bad request")
	}

	identifier := "test-" + util.SHA256(util.UUID())[:16]

	flag := fmt.Sprintf("%s{%s}", game.FlagPrefix, util.GenerateFlagContent(c.QueryParams().Get("flag_format"), user.UUID))
	// this is a test container, so we don't need to create the flag model

	info := &k8s.ContainerInfo{
		Identifier: identifier,
		Labels: map[string]string{
			"test": "true",
			"game": game.UUID,
			"user": user.UUID,
		},
		Flag: flag,
	}

	uuid, nodeDomain, err := ctx.ContainerManager.CreateContainer(payload, info, ctx.Store)

	if err != nil {
		return Failed(&c, "Unable to create container.")
	}

	container := &store.Container{
		Creator:          user,
		UUID:             uuid,
		Status:           store.ContainerStatusRunning,
		ExpireTime:       time.Now().UnixMilli() + ctx.Store.GetSettingInt64("container_expire_time"),
		LeftRenewalTimes: ctx.Store.GetSettingInt("max_container_renewal_times"),
		Identifier:       identifier,
		NodeDomain:       nodeDomain,
	}

	ctx.Store.CreateContainer(container)

	return OKWithData(&c, map[string]interface{}{
		"identifier": identifier,
		"uuid":       uuid,
		"entrance":   fmt.Sprintf("%s.%s", uuid, nodeDomain),
		"flag":       flag,
	})
}

func DisposeContainer(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	user, _ := GetUserFromToken(&c)
	game, _ := ctx.Store.GetGameByUUID(c.Param("game_uuid"))

	if user.Privilege < store.UserPrivilegeAdministrator && !game.IsManager(user) {
		return PermissionDenied(&c)
	}

	container, err := ctx.Store.GetContainerByUUID(c.Param("container_uuid"))

	if err != nil {
		return Failed(&c, "Unable to fetch container")
	}

	ctx.ContainerManager.DisposeContainer(container.Identifier)

	container.Status = store.ContainerStatusStopped
	ctx.Store.UpdateContainer(container)

	return OK(&c)
}

func GetChallengeRunningContainer(c echo.Context) error {
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

	if container.Status != store.ContainerStatusRunning {
		return Failed(&c, "Container is not running")
	}

	return OKWithData(&c, map[string]interface{}{
		"uuid":               container.UUID,
		"entrance":           fmt.Sprintf("%s.%s", container.UUID, container.NodeDomain),
		"expire":             container.ExpireTime,
		"left_renewal_times": container.LeftRenewalTimes,
	})
}
