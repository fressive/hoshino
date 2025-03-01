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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"rina.icu/hoshino/internal/util"
	"rina.icu/hoshino/server/context"
	"rina.icu/hoshino/store"
)

func UploadAttachment(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	user, _ := GetUserFromToken(&c)
	game, err := ctx.Store.GetGameByUUID(c.Param("game_uuid"))
	if err != nil {
		return Failed(&c, "Unable to fetch game")
	}

	if !game.IsManager(user) && !user.HasPrivilege(store.UserPrivilegeAdministrator) {
		return PermissionDenied(&c)
	}

	challenge, err := ctx.Store.GetChallengeByUUID(c.Param("challenge_uuid"))
	if err != nil {
		return Failed(&c, "Unable to fetch challenge")
	}

	name := c.FormValue("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to get file from request"})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to open the file"})
	}
	defer src.Close()

	uuid := util.UUID()
	filename := filepath.Clean(filepath.Base(uuid))
	dstPath := filepath.Join(ctx.Config.DataDir, "attachments", filename)

	if !strings.HasPrefix(dstPath, filepath.Join(ctx.Config.DataDir, "attachments")) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid file path"})
	}

	dst, err := os.Create(dstPath)
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to copy the file"})
	}

	downloadName := c.FormValue("download_name")
	if downloadName == "" {
		downloadName = uuid + filepath.Ext(file.Filename)
	}

	attachment := &store.Attachment{
		UUID:         uuid,
		SavePath:     dstPath,
		Name:         name,
		DownloadName: downloadName,
		Challenge:    challenge,
		Uploader:     user,
	}

	ctx.Store.CreateAttachment(attachment)

	return c.JSON(http.StatusOK, map[string]string{"message": "File uploaded successfully"})
}

func DeleteAttachment(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	user, _ := GetUserFromToken(&c)
	game, err := ctx.Store.GetGameByUUID(c.Param("game_uuid"))
	if err != nil {
		return Failed(&c, "Unable to fetch game")
	}

	if !game.IsManager(user) && !user.HasPrivilege(store.UserPrivilegeAdministrator) {
		return PermissionDenied(&c)
	}

	challenge, err := ctx.Store.GetChallengeByUUID(c.Param("challenge_uuid"))
	if err != nil {
		return Failed(&c, "Unable to fetch challenge")
	}

	attachment, err := ctx.Store.GetAttachmentByUUID(c.Param("attachment_uuid"))
	if err != nil {
		return Failed(&c, "Unable to fetch attachment")
	}

	if attachment.Challenge.UUID != challenge.UUID {
		return Failed(&c, "Attachment does not belong to the challenge")
	}

	if err := os.Remove(attachment.SavePath); err != nil {
		return Failed(&c, "Failed to delete the attachment")
	}

	ctx.Store.DeleteAttachment(attachment)

	return OK(&c)
}

func GetAttachment(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	user, _ := GetUserFromToken(&c)
	if user.Privilege < store.UserPrivilegeNormal {
		return PermissionDenied(&c)
	}

	game, err := ctx.Store.GetGameByUUID(c.Param("game_uuid"))
	if err != nil {
		return Failed(&c, "Unable to fetch game")
	}

	if !game.Visibility && !game.IsManager(user) {
		return PermissionDenied(&c)
	}

	challenge, err := ctx.Store.GetChallengeByUUID(c.Param("challenge_uuid"))
	if err != nil {
		return Failed(&c, "Unable to fetch challenge")
	}

	if !challenge.Game.Visibility && !game.IsManager(user) && !challenge.Ongoing() {
		return PermissionDenied(&c)
	}

	attachment, err := ctx.Store.GetAttachmentByUUID(c.Param("attachment_uuid"))
	if err != nil {
		return Failed(&c, "Unable to fetch attachment")
	}

	return c.Attachment(attachment.SavePath, attachment.DownloadName)
}

func GetAttachmentList(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	user, _ := GetUserFromToken(&c)
	if user.Privilege < store.UserPrivilegeNormal {
		return PermissionDenied(&c)
	}

	game, err := ctx.Store.GetGameByUUID(c.Param("game_uuid"))
	if err != nil {
		return Failed(&c, "Unable to fetch game")
	}

	team := game.GetTeamByUser(ctx.Store, user)
	if team == nil {
		return Failed(&c, "Unable to fetch team")
	}

	if !game.Visibility && !game.IsManager(user) {
		return PermissionDenied(&c)
	}

	challenge, err := ctx.Store.GetChallengeByUUID(c.Param("challenge_uuid"))
	if err != nil {
		return Failed(&c, "Unable to fetch challenge")
	}

	if !challenge.Game.Visibility && !game.IsManager(user) && !challenge.Ongoing() {
		return PermissionDenied(&c)
	}

	attachments, err := ctx.Store.GetAttachmentsByChallenge(challenge)

	if err != nil {
		return Failed(&c, "Unable to fetch attachments")
	}

	result := []map[string]interface{}{}
	exists := map[string]bool{}
	for _, attachment := range attachments {
		if exists[attachment.Name] {
			continue
		}

		if attachment.Multiple {
			ma, _ := ctx.Store.GetAttachmentsByName(attachment.Name)
			index := util.SHA256Uint64(team.UUID+challenge.UUID+attachment.Name) % uint64(len(ma))
			result = append(result, map[string]interface{}{
				"uuid": ma[index].UUID,
				"name": ma[index].Name,
			})

			exists[attachment.Name] = true
		} else {
			result = append(result, map[string]interface{}{
				"uuid": attachment.UUID,
				"name": attachment.Name,
			})

			exists[attachment.Name] = true
		}
	}

	return OKWithData(&c, attachments)
}
