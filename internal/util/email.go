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

package util

import (
	"bytes"
	"fmt"
	"log/slog"
	"text/template"

	"gopkg.in/gomail.v2"
	"rina.icu/hoshino/server/config"
	"rina.icu/hoshino/store"
)

var (
	dialer     *gomail.Dialer
	smtpConfig *config.SMTP
)

func InitSMTP(c *config.SMTP) error {
	dialer = gomail.NewDialer(c.Host, c.Port, c.Username, c.Password)
	if _, err := dialer.Dial(); err != nil {
		slog.Error("Error dialing SMTP server.", slog.Any("error", err))
		panic(err)
	}

	smtpConfig = c

	return nil
}

func SendEmail(s *store.Store, email string, templateFile string, title string, data map[string]interface{}) error {
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		slog.Error("Error parsing email template.", slog.Any("error", err))
		panic(err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		slog.Error("Error executing email template.", slog.Any("error", err))
		panic(err)
	}

	// Send email
	mail := gomail.NewMessage()

	mail.SetHeader("From", smtpConfig.Username)
	mail.SetHeader("To", email)
	mail.SetHeader("Subject", title)
	mail.SetBody("text/html", body.String())

	return dialer.DialAndSend(mail)
}

func SendVerificationEmail(s *store.Store, email string, nickname string, token string) error {
	title := fmt.Sprintf("%s - Email Verification", s.GetSettingString("site_name"))
	data := map[string]interface{}{
		"Title":    title,
		"Nickname": nickname,
		"Code":     token,
		"SiteName": s.GetSettingString("site_name"),
	}

	return SendEmail(s, email, "template/email/verification.html", title, data)
}
