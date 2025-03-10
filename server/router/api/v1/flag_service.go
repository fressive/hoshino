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
	"math"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/labstack/echo/v4"
	"rina.icu/hoshino/internal/util"
	"rina.icu/hoshino/server/context"
	"rina.icu/hoshino/store"
)

type (
	SubmitFlagPayloads struct {
		Flag string `json:"flag" validate:"required"`
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
		} else {
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
	} else {
		attachments, err := s.GetAttachmentsByChallenge(challenge)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to get attachments: %s ", err.Error()))
			return true, CheatReasonNone
		}

		checked := make(map[string]bool)
		for _, attachment := range attachments {
			if checked[attachment.Name] {
				continue
			}

			if attachment.Multiple {
				ma, _ := s.GetAttachmentsByName(attachment.Name)
				index := util.SHA256Uint64(team.UUID+challenge.UUID+attachment.Name) % uint64(len(ma))
				for i, a := range ma {
					if a.Flag == flag && i == int(index) {
						return true, CheatReasonNone
					}

					if a.Flag == flag {
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
			}

			checked[attachment.Name] = true
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

func updateScore(s *store.Store, challenge *store.Challenge) {
	// update the score of the challenge
	solvedFlags, err := s.GetSolvedFlagsByChallenge(challenge)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get solved flags: %s ", err.Error()))
		return
	}

	solvedRate := float64(len(solvedFlags)) / float64(len(challenge.Game.GetTeams(s)))
	lossRate := float64(len(solvedFlags)-1) / float64(len(challenge.Game.GetTeams(s)))
	exponentialScore := float64(challenge.Score) * math.Exp((float64(challenge.Difficulty)-2)*lossRate)
	parameters := make(map[string]interface{})
	parameters["original_score"] = challenge.Score
	parameters["solved_count"] = len(solvedFlags)
	parameters["team_count"] = len(challenge.Game.GetTeams(s))
	parameters["difficulty"] = challenge.Difficulty
	parameters["loss_rate"] = lossRate
	parameters["solved_rate"] = solvedRate
	parameters["unsolved_rate"] = 1 - solvedRate
	parameters["linear_score"] = float64(challenge.Score) * (1 - lossRate)
	parameters["exponential_score"] = exponentialScore

	slog.Info(fmt.Sprintf("Original score: %d, Solved count: %d, Team count: %d, Difficulty: %f, Loss rate: %f, Solved rate: %f, Unsolved rate: %f, Linear score: %f, Exponential score: %f", challenge.Score, len(solvedFlags), len(challenge.Game.GetTeams(s)), challenge.Difficulty, lossRate, solvedRate, 1-solvedRate, float64(challenge.Score)*(1-lossRate), exponentialScore))

	functions := map[string]govaluate.ExpressionFunction{
		"max": func(args ...interface{}) (interface{}, error) {
			return math.Max(args[0].(float64), args[1].(float64)), nil
		},
		"min": func(args ...interface{}) (interface{}, error) {
			return math.Min(args[0].(float64), args[1].(float64)), nil
		},
		"exponential_score_with_top3_bonus": func(args ...interface{}) (interface{}, error) {
			order := args[0].(uint32)
			rate1 := args[1].(float64)
			rate2 := args[2].(float64)
			rate3 := args[3].(float64)
			if order == 1 {
				return exponentialScore * rate1, nil
			} else if order == 2 {
				return exponentialScore * rate2, nil
			} else if order == 3 {
				return exponentialScore * rate3, nil
			}
			return exponentialScore, nil
		},
	}

	expression, err := govaluate.NewEvaluableExpressionWithFunctions(challenge.ScoreFormula, functions)

	actualOrder := 1
	for _, flag := range solvedFlags {
		if flag.State == 2 && challenge.Game.AutoBan {
			continue
		}

		parameters["order"] = actualOrder

		result, _ := expression.Evaluate(parameters)
		score := math.Round(result.(float64))
		flag.Score = int(score)
		s.UpdateFlag(flag)
		actualOrder++
	}
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
		return PermissionDenied(&c)
	}

	challenge, err := ctx.Store.GetChallengeByUUID(c.Param("challenge_uuid"))
	if err != nil {
		return Failed(&c, "Failed to submit the flag")
	}

	team := challenge.Game.GetTeamByUser(ctx.Store, user)

	if team == nil {
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
			Flag:      challenge.FlagFormat,
			State:     store.FlagUnsolved,
			SolvedAt:  time.Now().UnixMilli(),
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
		storedFlag.SolvedAt = time.Now().UnixMilli()
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

		storedFlag.SolvedAt = time.Now().UnixMilli()

		ctx.Store.UpdateFlag(storedFlag)
	} else {
		ctx.Store.UpdateFlag(storedFlag)
		return Failed(&c, "Flag is incorrect")
	}

	go updateScore(ctx.Store, challenge)

	return OK(&c)
}
