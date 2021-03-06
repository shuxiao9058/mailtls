// Copyright (c) 2016 The Mailtls Authors.
// Use of this source code is governed by an Expat-style
// MIT license that can be found in the LICENSE file.

// Package mailtls is used to send emails using SMTP/STARTTLS with
// SMTP plain authentication.
//
// Note: The API is presently experimental and may change.
package mailtls

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"
	"strings"
	"time"
)

// Email contains the information to send in an email
type Email struct {
	To      string
	From    string
	Subject string
	CC      []string
	BCC     []string
	Headers []string // headers other than To:, From:, Subject:, CC: and Date:
	Body    io.Reader
}

// Server contains SMTP server address and authentication information.
type Server struct {
	Address  string // e.g. smtp.example.com:587
	User     string
	Password string
}

// hostname returns a server address without the port number.
func hostname(address string) string {
	i := strings.LastIndex(address, ":")
	if i == -1 {
		return address
	}
	j := strings.LastIndex(address, "]")
	if j > i {
		return address // IPv6 without port number
	}
	return address[:i]
}

// Mail sends an email using STMP/STARTTLS with SMTP plain
// authentication. If no port number is given in s.Address, it
// defaults to 587. Mail will refuse to attempt authentication or send
// an email if TLS encryption to the sending server cannot be
// established.
func (s *Server) Mail(email *Email) error {
	// Connect to the remote SMTP server.
	address := s.Address
	if hostname(address) == address {
		address += ":587"
	}
	c, err := smtp.Dial(address)
	if err != nil {
		return err
	}

	// switch to TLS
	if err := c.StartTLS(
		&tls.Config{ServerName: hostname(address)}); err != nil {
		return err
	}

	// Set up authentication information.
	plainAuth := smtp.PlainAuth(
		"", s.User, s.Password, hostname(address))
	if err := c.Auth(plainAuth); err != nil {
		return err
	}

	// Set the sender and recipients first
	if err := c.Mail(email.From); err != nil {
		return err
	}
	if err := c.Rcpt(email.To); err != nil {
		return err
	}
	for _, to := range email.CC {
		if err := c.Rcpt(to); err != nil {
			return err
		}
	}
	for _, to := range email.BCC {
		if err := c.Rcpt(to); err != nil {
			return err
		}
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(wc,
		"To: %s\r\nFrom: %s\r\nSubject: %s\r\nDate: %s\r\n",
		email.To, email.From, email.Subject, time.Now().Format(time.RFC1123Z))
	if err != nil {
		return err
	}
	for _, to := range email.CC {
		_, err = fmt.Fprintf(wc, "CC: %s\r\n", to)
		if err != nil {
			return err
		}
	}
	for _, h := range email.Headers {
		_, err = fmt.Fprintf(wc, "%s\r\n", h)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(wc, "\r\n")
	if err != nil {
		return err
	}
	_, err = io.Copy(wc, email.Body)
	if err != nil {
		return err
	}
	err = wc.Close()
	if err != nil {
		return err
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		return err
	}
	return nil
}
