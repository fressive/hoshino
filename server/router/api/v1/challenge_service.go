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

package v1

import (
	"github.com/labstack/echo/v4"
	"rina.icu/hoshino/internal/util"
	"rina.icu/hoshino/server/context"
	"rina.icu/hoshino/store"
)

type CreateChallengePayload struct {
	Name                string   `json:"name" validate:"required"`
	Description         string   `json:"description" validate:"required"`
	Attachments         []string `json:"attachments"`
	CategoryUUID        string   `json:"category_uuid" validate:"required"`
	Tags                []string `json:"tags"`
	ExpireTime          int64    `json:"expire_time" validate:"required"`
	AfterExpiredOptions int32    `json:"after_expired_options" validate:"required"`
	DockerComposeFile   string   `json:"docker_compose_file"`
	NoContainer         bool     `json:"no_container"`
	DynamicFlag         bool     `json:"dynamic_flag"`
	Flag                string   `json:"flag"`
	Score               int      `json:"score"`
	ScoreMode           int      `json:"score_mode"`
	ScoreFormula        string   `json:"score_formula"`
	FakeFlag            []string `json:"fake_flag"`
	Hints               []string `json:"hints"`
}

func CreateChallenge(c echo.Context) error {
	ctx := c.(*context.CustomContext)
	user, _ := GetUserFromToken(&c)

	game_uuid := ctx.Param("game_uuid")
	game, err := ctx.Store.GetGameByUUID(game_uuid)
	if err != nil {
		return Failed(&c, "Unable to fetch game")
	}

	if !game.IsManager(user) || !user.HasPrivilege(store.UserPrivilegeAdministrator) {
		return PermissionDenied(&c)
	}

	var payload CreateChallengePayload
	if err := c.Bind(&payload); err != nil {
		return Failed(&c, "Invalid payload")
	}

	// TODO: Handle attachments

	category, err := ctx.Store.GetCategoryByUUID(payload.CategoryUUID)
	if err != nil {
		return Failed(&c, "Unable to fetch category")
	}

	uuid := util.UUID()
	challenge := &store.Challenge{
		Name:                   payload.Name,
		Description:            payload.Description,
		UUID:                   uuid,
		Game:                   game,
		Creator:                user,
		Category:               category,
		Tags:                   payload.Tags,
		ExpireTime:             payload.ExpireTime,
		AfterExpiredOperations: store.AfterExpireOp(payload.AfterExpiredOptions),
		DockerComposeFile:      payload.DockerComposeFile,
		NoContainer:            payload.NoContainer,
		DynamicFlag:            payload.DynamicFlag,
		Flag:                   payload.Flag,
		Score:                  payload.Score,
		ScoreMode:              store.ScoreMode(payload.ScoreMode),
		ScoreFormula:           payload.ScoreFormula,
		FakeFlag:               payload.FakeFlag,
		Hints:                  payload.Hints,
	}

	ctx.Store.CreateChallenge(challenge)
	game.Challenges = append(game.Challenges, challenge)
	ctx.Store.UpdateGame(game)

	if !payload.NoContainer {
		// TOOD: Prepare containers in agents here

	}

	return OKWithData(&c, uuid)
}

func GetFullChallenge(c echo.Context) error {
	ctx := c.(*context.CustomContext)
	user, _ := GetUserFromToken(&c)

	uuid := ctx.Param("challenge_uuid")
	challenge, err := ctx.Store.GetChallengeByUUID(uuid)
	if err != nil {
		return Failed(&c, "Unable to fetch challenge")
	}

	if !challenge.Game.Visibility && !user.HasPrivilege(store.UserPrivilegeAdministrator) && !challenge.Game.IsManager(user) {
		return PermissionDenied(&c)
	}

	return OKWithData(&c, challenge)
}

func GetFullChallenges(c echo.Context) error {
	ctx := c.(*context.CustomContext)
	user, _ := GetUserFromToken(&c)

	game_uuid := ctx.Param("game_uuid")
	game, err := ctx.Store.GetGameByUUID(game_uuid)
	if err != nil {
		return Failed(&c, "Unable to fetch game")
	}

	if !game.Visibility && !user.HasPrivilege(store.UserPrivilegeAdministrator) && !game.IsManager(user) {
		return PermissionDenied(&c)
	}

	return OKWithData(&c, game.Challenges)
}
