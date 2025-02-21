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

package config

import "os"

type Config struct {
	// prod, dev, demo
	Mode string `json:"mode" mapstructure:"mode"`

	// The binding address of the server
	Address string `json:"address" mapstructure:"address"`

	// The port of the server
	Port int `json:"port" mapstructure:"port"`

	// Data directory
	DataDir string `json:"data_dir" mapstructure:"data_dir"`

	// DSN
	DSN string `json:"dsn" mapstructure:"dsn"`

	// The database driver
	// sqlite, mysql
	Driver string `json:"driver" mapstructure:"driver"`

	// The version of the server
	Version string `json:"version" mapstructure:"version"`

	// The secret key to generate JWT token
	Secret string `json:"secret" mapstructure:"secret"`

	// The SMTP configuration
	SMTP SMTP `json:"smtp" mapstructure:"smtp"`

	// The container services
	Agents []Agent `json:"agents" mapstructure:"agents"`
}

type Agent struct {
	Name     string `json:"name" mapstructure:"name"`
	Address  string `json:"address" mapstructure:"address"`
	Port     int    `json:"port" mapstructure:"port"`
	Token    string `json:"token" mapstructure:"token"`
	CertFile string `json:"cert_file" mapstructure:"cert_file"`

	// the weight is used to set load-balancing when multiple agents available
	Weight int
}

type SMTP struct {
	Host     string `json:"host" mapstructure:"host"`
	Port     int    `json:"port" mapstructure:"port"`
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`
}

func (c *Config) IsDev() bool {
	return c.Mode != "prod"
}

func (c *Config) Validate() error {
	if c.Mode != "demo" && c.Mode != "dev" && c.Mode != "prod" {
		c.Mode = "demo"
	}

	// create data dir if not exists
	_, err := os.Stat(c.DataDir)
	if os.IsNotExist(err) {
		os.MkdirAll(c.DataDir, os.ModePerm)
	}

	if c.Secret == "" {
		panic("Secret key is required")
	}

	return nil
}
