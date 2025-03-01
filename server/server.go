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
	"k8s.io/client-go/kubernetes"

	"rina.icu/hoshino/internal/k8s"
	"rina.icu/hoshino/plugins/cron"
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
	k8sClient  *kubernetes.Clientset
}

func registerRouter(s *Server) {
	e := s.echoServer
	c := s.config

	g := e.Group("/api/v1")
	// Rate limitin
	g.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))
	// CORS
	g.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     c.CORS.AllowOrigins,
		AllowCredentials: true,
	}))
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
	userApi.GET("", v1.GetUserInfo).Name = "get-user-info"
	userApi.POST("/register", v1.UserRegister).Name = "user-register"
	userApi.GET("/register/check", v1.AllowRegister).Name = "check-register"
	userApi.POST("/login", v1.UserLogin).Name = "user-login"
	userApi.POST("/username/check", v1.CheckUsername).Name = "check-username"
	userApi.POST("/email/check", v1.CheckEmail).Name = "check-email"
	userApi.POST("/email/verify", v1.EmailVerify).Name = "verify-email"
	userApi.POST("/email/send", v1.SendVerificationEmail).Name = "send-email"

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
	teamApi.GET("", v1.GetUserTeamIngame).Name = "get-user-team-ingame"
	// teamApi.GET("", v1.GetTeams).Name = "get-teams"
	// teamApi.GET("/:uuid", v1.GetTeam).Name = "get-team"
	teamApi.POST("/create", v1.CreateTeam).Name = "create-team"
	// teamApi.POST("/:uuid/ban", v1.BanTeam).Name = "ban-team"

	// Challenge APIs
	challengeApi := gameApi.Group("/:game_uuid/challenge")
	challengeApi.GET("", v1.GetFullChallenges).Name = "get-challenges"
	challengeApi.GET("/:challenge_uuid", v1.GetFullChallenge).Name = "get-challenge"
	challengeApi.POST("/create", v1.CreateChallenge).Name = "create-challenge"
	challengeApi.POST("/create/container", v1.CreateContainer).Name = "create-test-container"
	challengeApi.DELETE("/create/container/:container_uuid", v1.DisposeContainer).Name = "dispose-test-container"
	challengeApi.POST("/:challenge_uuid/flag", v1.SubmitFlag).Name = "submit-flag"
	challengeApi.POST("/:challenge_uuid/container/create", v1.CreateChallengeContainer).Name = "create-challenge-container"

	// Container APIs
	containerApi := challengeApi.Group("/:challenge_uuid/container")
	containerApi.POST("/create", v1.CreateChallengeContainer).Name = "create-container"
	containerApi.POST("/dispose", v1.DisposeChallengeContainer).Name = "dispose-container"

	// Attachment APIs
	attachmentApi := challengeApi.Group("/:challenge_uuid/attachment")
	attachmentApi.POST("/", v1.UploadAttachment).Name = "upload-attachment"
	attachmentApi.GET("/:attachment_uuid", v1.GetAttachment).Name = "get-attachment"
	attachmentApi.GET("/", v1.GetAttachmentList).Name = "get-attachments"

	// Health check
	e.GET("/health", router.HealthService)
}

func NewServer(ctx context.Context, config *config.Config, clientSet *kubernetes.Clientset) (*Server, error) {
	s := &Server{
		config: config,
	}

	echoServer := echo.New()
	echoServer.Debug = true
	echoServer.HideBanner = true
	echoServer.HidePort = true

	echoServer.Use(middleware.RecoverWithConfig(
		middleware.RecoverConfig{
			LogErrorFunc: middleware.LogErrorFunc(func(c echo.Context, err error, stack []byte) error {
				slog.Error(fmt.Sprintf("Error: %v\nStack: %v", err, string(stack)))
				return err
			}),
		},
	))

	if config.IsDev() {
		echoServer.Use(middleware.Logger())
	}

	s.echoServer = echoServer
	s.k8sClient = clientSet

	store, err := store.GetStore(config)

	containerManager := &k8s.ContainerManager{
		K8SClient: clientSet,
		Store:     store,
	}

	echoServer.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := &cc.CustomContext{
				Context:          c,
				Config:           s.config,
				Store:            s.store,
				ContainerManager: containerManager,
			}
			return next(ctx)
		}
	})

	if err != nil {
		slog.Error("Failed to connect to database")
		panic(err)
	}

	s.store = store

	// Cron

	cron.InitContainerCron(store, containerManager)

	registerRouter(s)

	// TODO: frontend server register here

	slog.Info(fmt.Sprintf("Hoshino web server is running on %s:%d.", config.Address, config.Port))
	return s, nil
}

func (s *Server) Startup() error {
	return s.echoServer.Start(fmt.Sprintf("%s:%d", s.config.Address, s.config.Port))
}
