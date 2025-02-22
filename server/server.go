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

package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"rina.icu/hoshino/server/config"
	cc "rina.icu/hoshino/server/context"
	hmw "rina.icu/hoshino/server/middleware"
	"rina.icu/hoshino/server/router"
	v1 "rina.icu/hoshino/server/router/api/v1"
	"rina.icu/hoshino/store"
)

type Server struct {
	config *config.Config

	echoServer *echo.Echo
	store      *store.Store
}

func registerRouter(s *Server) {
	e := s.echoServer
	c := s.config

	g := e.Group("/api/v1")
	// Rate limitin
	g.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))
	// Body limitin
	g.Use(middleware.BodyLimitWithConfig(
		middleware.BodyLimitConfig{
			Limit: "2M",
			Skipper: func(c echo.Context) bool {
				// TODO: skip some upload cases
				return false
			},
		},
	))
	// JWT
	g.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(c.Secret),
		Skipper: func(c echo.Context) bool {
			return !router.RequireLogin(c.Path())
		},
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwt.MapClaims)
		},
	}))
	// Permission check
	g.Use(hmw.RequireLoginMiddleware)

	// User APIs
	userApi := g.Group("/user")
	userApi.POST("/register", v1.UserRegister).Name = "user-register"
	userApi.POST("/login", v1.UserLogin).Name = "user-login"
	userApi.POST("/username/check", v1.CheckUsername).Name = "check-username"
	userApi.POST("/email/check", v1.CheckEmail).Name = "check-email"
	userApi.POST("/email/verify", v1.EmailVerify).Name = "verify-email"
	userApi.GET("/email/resend", v1.ResendVerificationEmail).Name = "resend-email"

	// Container APIs
	containersApi := g.Group("/container")
	containersApi.POST("/create", v1.CreateContainer).Name = "create-container"

	// Setting APIs
	settingApi := g.Group("/setting")
	settingApi.GET("", v1.GetAllSettings).Name = "get-all-settings"
	settingApi.GET("/:key", v1.GetSetting).Name = "get-setting"
	settingApi.POST("/:key", v1.UpdateSetting).Name = "set-setting"

	// Game APIs
	gameApi := g.Group("/game")
	gameApi.GET("", v1.GetGames).Name = "get-games"
	gameApi.GET("/:game_uuid", v1.GetGame).Name = "get-game"
	gameApi.POST("/create", v1.CreateGame).Name = "create-game"

	// Team APIs
	teamApi := gameApi.Group("/:game_uuid/team")
	// teamApi.GET("", v1.GetTeams).Name = "get-teams"
	// teamApi.GET("/:uuid", v1.GetTeam).Name = "get-team"
	teamApi.POST("/create", v1.CreateTeam).Name = "create-team"
	// teamApi.POST("/:uuid/ban", v1.BanTeam).Name = "ban-team"

	// Category APIs
	categoryApi := gameApi.Group("/:game_uuid/category")
	categoryApi.GET("", v1.GetCategories).Name = "get-categories"
	categoryApi.GET("/:category_uuid", v1.GetCategory).Name = "get-category"
	categoryApi.POST("/create", v1.CreateCategory).Name = "create-category"

	// Challenge APIs
	challengeApi := gameApi.Group("/:game_uuid/challenge")
	challengeApi.GET("", v1.GetChallenges).Name = "get-challenges"
	challengeApi.GET("/:challenge_uuid", v1.GetChallenge).Name = "get-challenge"
	challengeApi.POST("/create", v1.CreateChallenge).Name = "create-challenge"
	challengeApi.POST("/:challenge_uuid/flag", v1.SubmitFlag).Name = "submit-flag"
	challengeApi.POST("/:challenge_uuid/container/create", v1.CreateContainer).Name = "create-challenge-container"

	// Health check
	e.GET("/health", router.HealthService)
}

func NewServer(ctx context.Context, config *config.Config) (*Server, error) {
	s := &Server{
		config: config,
	}

	echoServer := echo.New()
	echoServer.Debug = true
	echoServer.HideBanner = true
	echoServer.HidePort = true

	echoServer.Use(middleware.Recover())
	if config.IsDev() {
		echoServer.Use(middleware.Logger())
	}
	echoServer.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := &cc.CustomContext{
				Context: c,
				Config:  s.config,
				Store:   s.store,
			}
			return next(ctx)
		}
	})

	s.echoServer = echoServer

	store, err := store.GetStore(config)

	if err != nil {
		slog.Error("Failed to connect to database")
		panic(err)
	}

	s.store = store

	registerRouter(s)

	// TODO: frontend server register here

	slog.Info(fmt.Sprintf("Hoshino web server is running on %s:%d.", config.Address, config.Port))
	return s, nil
}

func (s *Server) Startup() error {
	return s.echoServer.Start(fmt.Sprintf("%s:%d", s.config.Address, s.config.Port))
}
