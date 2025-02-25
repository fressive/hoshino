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

type CreateCategoryPayload struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Visibility  bool   `json:"visibility"`
}

func CreateCategory(c echo.Context) error {
	ctx := c.(*context.CustomContext)
	user, _ := GetUserFromToken(&c)

	uuid := c.Param("game_uuid")
	game, err := ctx.Store.GetGameByUUID(uuid)
	if err != nil || !(game.Visibility || game.IsManager(user) || user.HasPrivilege(store.UserPrivilegeAdministrator)) {
		return Failed(&c, "Unable to fetch game")
	}

	if !(game.IsManager(user) || user.HasPrivilege(store.UserPrivilegeAdministrator)) {
		return PermissionDenied(&c)
	}

	var payload CreateCategoryPayload
	if err := c.Bind(&payload); err != nil {
		return Failed(&c, "Invalid payload")
	}

	categoryUUID := util.UUID()
	category := &store.Category{
		Name:        payload.Name,
		Description: payload.Description,
		Visibility:  payload.Visibility,
		UUID:        categoryUUID,
		Game:        game,
	}

	ctx.Store.CreateCategory(category)

	game.Categories = append(game.Categories, category)
	ctx.Store.UpdateGame(game)

	return OKWithData(&c, categoryUUID)
}

func GetCategory(c echo.Context) error {
	ctx := c.(*context.CustomContext)
	user, _ := GetUserFromToken(&c)

	uuid := c.Param("game_uuid")

	game, err := ctx.Store.GetGameByUUID(uuid)
	if err != nil || !(game.Visibility || game.IsManager(user) || user.HasPrivilege(store.UserPrivilegeAdministrator)) {
		return Failed(&c, "Unable to fetch game")
	}

	uuid = c.Param("category_uuid")
	category, err := ctx.Store.GetCategoryByUUID(uuid)
	if err != nil || !(game.Visibility || game.IsManager(user) || user.HasPrivilege(store.UserPrivilegeAdministrator)) {
		return Failed(&c, "Unable to fetch category")
	}

	return OKWithData(&c, category)
}

func GetCategories(c echo.Context) error {
	ctx := c.(*context.CustomContext)
	user, _ := GetUserFromToken(&c)

	uuid := c.Param("game_uuid")
	game, err := ctx.Store.GetGameByUUID(uuid)
	if err != nil || !(game.Visibility || game.IsManager(user) || user.HasPrivilege(store.UserPrivilegeAdministrator)) {
		return Failed(&c, "Unable to fetch game")
	}

	categories := ctx.Store.GetCategoriesByGame(game)

	return OKWithData(&c, categories)
}
