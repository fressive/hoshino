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

package cron

import (
	"log/slog"

	"github.com/robfig/cron/v3"
	"rina.icu/hoshino/internal/k8s"
	"rina.icu/hoshino/store"
)

func InitContainerCron(s *store.Store, cm *k8s.ContainerManager) {
	slog.Info("Initializing container cron job")

	c := cron.New(cron.WithSeconds())

	c.AddFunc("@every 10s", func() {
		expiredContainers, err := s.GetExpiredRunningContainers()
		if err != nil {
			slog.Error("Failed to get expired containers: " + err.Error())
			return
		}

		for _, container := range expiredContainers {
			if err := cm.DisposeChallengeContainer(container.Challenge, container.Creator); err != nil {
				slog.Error("Failed to delete container: " + err.Error())
			}

			container.Status = store.ContainerStatusExpired
			if err := s.UpdateContainer(container); err != nil {
				slog.Error("Failed to delete container: " + err.Error())
			}
		}
	})

	c.Start()
}
