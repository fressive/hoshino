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

package internal

var DefaultSettings = map[string]string{
	"site_name":      "HoshinoCTF",
	"site_domain":    "hoshino.rina.icu",
	"site_desc":      "HoshinoCTF - A Lightweight CTF platform",
	"allow_register": "true",

	// useful when we create something flag-like
	"flag_prefix": "hoshino",

	// the regex for email validation, can be used to restrict the email's domain
	// for example, `.*@example\.com`
	"email_regex":       ".*",
	"need_email_verify": "true",

	"max_container_per_user":      "1",
	"max_container_renewal_times": "3",
	"container_expire_time":       "3600000",

	"node_domain": "node.hoshino.rina.icu",
}
