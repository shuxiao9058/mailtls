// Copyright (c) 2016 The Mailtls Authors.
// Use of this source code is governed by a Expat-style
// MIT license that can be found in the LICENSE file.

package mailtls_test

import (
	"log"
	"strings"

	"xi2.org/x/mailtls"
)

func ExampleServer_Mail() {
	server := &mailtls.Server{
		Address:  "mail.example.com",
		User:     "myusername",
		Password: "mypassword",
	}
	err := server.Mail(&mailtls.Email{
		To:      "Someone <someone@example.com>",
		From:    "Me <me@example.com>",
		Subject: "A subject",
		Body:    strings.NewReader("Lol.\r\n"),
	})
	if err != nil {
		log.Fatal(err)
	}
}
