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
	"fmt"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	"rina.icu/hoshino/server/context"
	"rina.icu/hoshino/store"
)

type (
	SubmitFlagPayloads struct {
		TeamUUID string `json:"team_uuid" validate:"required"`
		Flag     string `json:"flag" validate:"required"`
	}
)

type CheatReason int

const (
	CheatReasonNone CheatReason = iota
	CheatReasonSharingFlag
	CheatReasonFakeFlag
)

func anticheatCheck(_ *echo.Context, s *store.Store,
	flag string,
	team *store.Team,
	challenge *store.Challenge,
) (bool, CheatReason) {
	// anti-cheat

	if challenge.DynamicFlag {
		// check if the flag was shared by multiple teams
		flag, err := s.GetFlagByChallenge(flag, challenge)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to get flag: %s ", err.Error()))
			return false, CheatReasonNone
		}

		if flag.Team.ID != team.ID {
			event := store.GameEvent{
				Content:      fmt.Sprintf("Team `%s` shared the flag `%s` with team `%s`", team.Name, flag.Flag, flag.Team.Name),
				Game:         challenge.Game,
				Challenge:    challenge,
				RelatedTeams: []*store.Team{team, flag.Team},
				Visibility:   true,
				Type:         store.GameEventTypeCheatDetected,
			}

			flag.State = store.FlagCheated
			s.UpdateFlag(flag)

			if challenge.Game.AutoBan {
				// ban the team instantly
				team.Banned = true
				s.UpdateTeam(team)
			} else {
				// log the cheat silently
				event.Visibility = false
			}

			s.CreateGameEvent(&event)
			return false, CheatReasonSharingFlag
		}
	}

	for _, fake := range challenge.FakeFlag {
		if fake == flag {

			event := store.GameEvent{
				Content:      fmt.Sprintf("Team `%s` submitted a fake flag `%s`", team.Name, flag),
				Game:         challenge.Game,
				Challenge:    challenge,
				RelatedTeams: []*store.Team{team},
				Visibility:   true,
				Type:         store.GameEventTypeCheatDetected,
			}

			if challenge.Game.AutoBan {
				// ban the team instantly
				team.Banned = true
				s.UpdateTeam(team)
			} else {
				// log the cheat silently
				event.Visibility = false
			}

			s.CreateGameEvent(&event)
			return false, CheatReasonFakeFlag
		}
	}

	return true, CheatReasonNone
}

func SubmitFlag(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	payload := new(SubmitFlagPayloads)
	if err := c.Bind(payload); err != nil {
		return Failed(&c, "Bad request")
	}

	user, _ := GetUserFromToken(&c)
	if user.Privilege < store.UserPrivilegeNormal {
		// make sure that the user has the privilege to submit flags
		return Unauthorized(&c)
	}

	team, err := ctx.Store.GetTeamByUUID(payload.TeamUUID)

	if err != nil {
		return Failed(&c, "Failed to submit the flag")
	}

	challenge, err := ctx.Store.GetChallengeByUUID(c.Param("challenge_uuid"))
	if err != nil {
		return Failed(&c, "Failed to submit the flag")
	}

	// the game should be active
	if challenge.Game.Status == store.GameStatusInactive {
		return Failed(&c, "Failed to submit the flag")
	}

	// lets check the challenge's stuff
	if challenge.State != store.ChallengeStateVisible ||
		(challenge.Expired() && challenge.AfterExpiredOperations&store.AfterExpireSubmitFlag == 0) {
		return Failed(&c, "Failed to submit the flag")
	}

	if !challenge.DynamicFlag {
		// static flag case, create a flag first
		// Create a new flag object here
		f := store.Flag{
			Challenge: challenge,
			Team:      team,
			Flag:      challenge.Flag,
			State:     store.FlagUnsolved,
			SolvedAt:  time.Now().Unix(),
		}

		ctx.Store.CreateFlag(&f)
	}

	// get the specified flag object by challenge and team, for there may be fixed flag cases
	storedFlag, err := ctx.Store.GetFlagByChallengeAndTeam(challenge, team)

	if err != nil {
		return Failed(&c, "Failed to submit the flag")
	}

	if storedFlag.State > store.FlagUnsolved {
		// state in [solved, cheated]
		return Failed(&c, "Flag has already been solved")
	}

	// anti-cheat
	ok, _ := anticheatCheck(&c, ctx.Store, payload.Flag, team, challenge)
	if !ok {
		storedFlag.State = store.FlagCheated
		ctx.Store.UpdateFlag(storedFlag)

		if challenge.Game.AutoBan {
			return Failed(&c, "Cheat detected")
		}
		// don't return if auto-ban is disabled
		// and calculate the score normally
		// let admins decide whether to ban the team later
	}

	if storedFlag.Flag == payload.Flag {
		if storedFlag.State == store.FlagUnsolved {
			// we'll not update the flag state if it was cheated
			storedFlag.State = store.FlagSolved
		}

		storedFlag.SolvedAt = time.Now().Unix()

		// TODO: calculate the score of the flag for some game modes
		if !storedFlag.Challenge.Expired() {
			if storedFlag.Challenge.ScoreMode == store.ScoreModeStatic {
				// only set the score when the flag is static

			}
		}
		ctx.Store.UpdateFlag(storedFlag)
	} else {
		ctx.Store.UpdateFlag(storedFlag)
		return Failed(&c, "Flag is incorrect")
	}
	return OK(&c)
}
