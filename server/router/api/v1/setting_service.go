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
	"rina.icu/hoshino/server/context"
	"rina.icu/hoshino/store"
)

type (
	UpdateSettingRequest struct {
		Value string `json:"value"`
	}
)

func UpdateSetting(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	user, _ := GetUserFromToken(&c)
	// We can get the user successfully cuz the middleware has verified the token
	if user.Privilege != store.UserPrivilegeHost {
		return PermissionDenied(&c)
	}

	var req UpdateSettingRequest
	if err := c.Bind(&req); err != nil {
		return Failed(&c, "Invalid payload")
	}

	ctx.Store.SetSetting(c.Param("key"), req.Value)

	return OK(&c)
}

func GetSetting(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	user, _ := GetUserFromToken(&c)
	if user.Privilege != store.UserPrivilegeHost {
		return PermissionDenied(&c)
	}

	setting, _ := ctx.Store.GetSetting(c.Param("key"))

	return c.JSON(200, map[string]interface{}{
		"message": "OK",
		"status":  "success",
		"result":  setting.Value,
	})
}

func GetAllSettings(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	user, _ := GetUserFromToken(&c)
	if user.Privilege != store.UserPrivilegeHost {
		return PermissionDenied(&c)
	}

	settings, _ := ctx.Store.GetAllSettings()

	resultMap := make(map[string]string)
	for _, setting := range settings {
		resultMap[setting.Key] = setting.Value
	}

	return c.JSON(200, map[string]interface{}{
		"message": "OK",
		"status":  "success",
		"result":  resultMap,
	})
}
