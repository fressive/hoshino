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

type CreateGamePayload struct {
	Name               string `json:"name" validate:"required"`
	Description        string `json:"description" validate:"required"`
	Visibility         bool   `json:"visibility" validate:"required"`
	FlagPrefix         string `json:"flag_prefix" validate:"required"`
	StartTime          int64  `json:"start_time" validate:"required"`
	EndTime            int64  `json:"end_time" validate:"required"`
	MaxTeamSize        int    `json:"max_team_size" validate:"required"`
	EnableChangeMember bool   `json:"enable_change_member" validate:"required"`
	AutoBan            bool   `json:"auto_ban" validate:"required"`
}

func GetGame(c echo.Context) error {
	ctx := c.(*context.CustomContext)
	user, _ := GetUserFromToken(&c)

	uuid := c.Param("game_uuid")
	game, err := ctx.Store.GetGameByUUID(uuid)
	if err != nil || !(game.Visibility || user.HasPrivilege(store.UserPrivilegeAdministrator) || game.IsManager(user)) {
		return Failed(&c, "Unable to fetch game")
	}

	return OKWithData(&c, game)
}

func GetGames(c echo.Context) error {
	ctx := c.(*context.CustomContext)
	// Get all games
	user, _ := GetUserFromToken(&c)

	var games []*store.Game
	var err error

	if user.HasPrivilege(store.UserPrivilegeAdministrator) {
		games, err = ctx.Store.GetGames()
	} else {
		games, err = ctx.Store.GetGamesWithoutSecrets()
	}
	if err != nil {
		return Failed(&c, "Unable to fetch games")
	} else {
		filtered := make([]*store.Game, 0)
		for _, game := range games {
			if game.Visibility || user.HasPrivilege(store.UserPrivilegeAdministrator) || game.IsManager(user) {
				filtered = append(filtered, game)
			}
		}
		return OKWithData(&c, filtered)
	}
}

func CreateGame(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	user, _ := GetUserFromToken(&c)

	if !user.HasPrivilege(store.UserPrivilegeAdministrator) {
		return Unauthorized(&c)
	}

	var payload CreateGamePayload
	if err := c.Bind(&payload); err != nil {
		return Failed(&c, "Invalid payload")
	}

	game := store.Game{
		UUID:               util.UUID(),
		Name:               payload.Name,
		Description:        payload.Description,
		FlagPrefix:         payload.FlagPrefix,
		Visibility:         payload.Visibility,
		StartTime:          payload.StartTime,
		EndTime:            payload.EndTime,
		MaxTeamSize:        payload.MaxTeamSize,
		EnableChangeMember: payload.EnableChangeMember,
		AutoBan:            payload.AutoBan,
		Creator:            user,
		Managers:           []*store.User{user},
	}

	if err := ctx.Store.CreateGame(&game); err != nil {
		return Failed(&c, "Unable to create game")
	}

	return OK(&c)
}
