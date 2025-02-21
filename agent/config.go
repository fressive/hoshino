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

package agent

import "os"

type Config struct {
	// prod, dev, demo
	Mode string `json:"mode"`

	// The binding address of the server
	Addr string `json:"address"`

	// The port of the server
	Port int `json:"port"`

	// Data directory
	DataDir string `json:"data_dir"`

	// The sha256ed secret key
	Secret string `json:"secret"`

	// The .pem file path
	CertFile string `json:"cert_file"`

	// The .key file path
	PrivateKeyFile string `json:"private_key_file"`
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
