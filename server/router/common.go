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

package router

func RequireLogin(path string) bool {
	return (path != "/api/v1/user/login" &&
		path != "/api/v1/user/register/check" &&
		path != "/api/v1/user/register" &&
		path != "/api/v1/user/username/check" &&
		path != "/api/v1/user/email/check")
}

func RequireEmailVerified(path string) bool {
	return (path != "/api/v1/user/email/verify" &&
		path != "/api/v1/user/email/send" &&
		path != "/api/v1/user" &&
		RequireLogin(path))
}
