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

type CreateTeamRequest struct {
	Name string `json:"name"`
}

func GetUserTeamIngame(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	user, _ := GetUserFromToken(&c)
	game, err := ctx.Store.GetGameByUUID(ctx.Param("game_uuid"))
	if err != nil || !(game.Visibility || user.HasPrivilege(store.UserPrivilegeAdministrator) || game.IsManager(user)) {
		return Failed(&c, "Unable to fetch game")
	}

	team := game.GetTeamByUser(ctx.Store, user)

	if team == nil {
		return Failed(&c, "You are not in a team")
	}

	return OKWithData(&c,
		team,
	)
}

func CreateTeam(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	var req CreateTeamRequest
	if err := ctx.Bind(&req); err != nil {
		return Failed(&c, "Invalid payload")
	}

	user, _ := GetUserFromToken(&c)
	game, err := ctx.Store.GetGameByUUID(ctx.Param("game_uuid"))
	if err != nil || !(game.Visibility || user.HasPrivilege(store.UserPrivilegeAdministrator) || game.IsManager(user)) {
		return Failed(&c, "Unable to fetch game")
	}

	if user.IsInTeam(ctx.Store, game) {
		return Failed(&c, "You are already in a team")
	}

	uuid := util.UUID()
	ctx.Store.CreateTeam(&store.Team{
		Name:     req.Name,
		UUID:     uuid,
		Game:     game,
		Creator:  user,
		Managers: []*store.User{user},
		Members:  []*store.User{user},
	})

	return OKWithData(&c, map[string]any{
		"team_uuid": uuid,
	})
}
