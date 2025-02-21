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

package middleware

import (
	"github.com/labstack/echo/v4"
	"rina.icu/hoshino/server/context"
	"rina.icu/hoshino/server/router"
	v1 "rina.icu/hoshino/server/router/api/v1"
)

func RequireLoginMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*context.CustomContext)

		if router.RequireLogin(c.Path()) {
			user, ok := v1.GetUserFromToken(&c)
			if !ok {
				return c.JSON(401, map[string]interface{}{
					"message": "Unauthorized",
					"status":  "error",
					"result":  false,
				})
			}

			if ctx.Store.GetSettingBool("need_email_verify") &&
				router.RequireEmailVerified(c.Path()) && !user.EmailVerified {
				return c.JSON(401, map[string]interface{}{
					"message": "Email not verified",
					"status":  "error",
					"result":  false,
				})
			}
		}
		return next(c)
	}
}
