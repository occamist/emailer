// Package brevo makes it easy to send emails via brevo provider. This package follows [brevo spec] strictly.
//
// Example usage:
//
//	 email := emailer.Email{
//		From:        "skywalker@jedi.com",
//		To:          []string{"vindu@sith.com"},
//		Subject:     "peace",
//		TextContent: "peace was never an option",
//	 }
//	 c, err := New(emailer.Config{key: "api-key"})
//		if err != nil {
//			//check err
//		}
//	 c.Send(ctx, email)
//
// [brevo spec]: https://developers.brevo.com/reference/sendtransacemail
package brevo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"slices"
	"strings"

	"github.com/occamist/emailer"
)

const endpoint = "https://api.brevo.com/v3/smtp/email"

// EmailClient is brevo email client to interact with emails
type EmailClient struct {
	key    string
	client http.Client
}

// New creates a new brevo email client with given API key and http.Client
func New(c emailer.Config) (*EmailClient, error) {
	if strings.TrimSpace(c.Key) == "" {
		return nil, errors.New("brevo API key is blank")
	}

	e := &EmailClient{
		key:    c.Key,
		client: c.Client,
	}
	return e, nil
}

// Detail is additional info about the person such as email and name
type Detail struct {
	Email string `json:"email"`
}

// payload is a request that brevo uses to send email
type payload struct {
	Sender      Detail   `json:"sender"`
	To          []Detail `json:"to"`
	BCC         []Detail `json:"bcc"`
	CC          []Detail `json:"cc"`
	Subject     string   `json:"subject"`
	HTMLContent string   `json:"htmlContent,omitempty"`
	TextContent string   `json:"textContent,omitempty"`
}

// errorMessage is a response when brevo encounters a problem while sending email
type errorMessage struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Send sends a given email
func (c *EmailClient) Send(ctx context.Context, email emailer.Email) error {
	var p payload
	p.Sender.Email = email.From
	for _, e := range email.To {
		p.To = append(p.To, Detail{Email: e})
	}
	for _, e := range email.BCC {
		p.BCC = append(p.BCC, Detail{Email: e})
	}
	for _, e := range email.CC {
		p.CC = append(p.CC, Detail{Email: e})
	}
	p.Subject = email.Subject
	p.HTMLContent = email.HTMLContent
	p.TextContent = email.TextContent

	raw, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("json.Marshal(%v): %v", p, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(raw))
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext(): %v", err)
	}
	req.Header.Add("api-key", c.key)
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	resp, err := c.client.Do(req) //nolint:gosec //endpoint is already a constant
	if err != nil {
		return fmt.Errorf("client.Do(%v): %v", req, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if slices.Contains([]int{http.StatusAccepted, http.StatusCreated, http.StatusOK}, resp.StatusCode) {
		return nil
	}

	var m errorMessage
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		dump, _ := httputil.DumpResponse(resp, true)
		return fmt.Errorf("json.NewDecoder(%v).Decode(): %v", string(dump), err)
	}

	return fmt.Errorf("unsuccessful response with status code(%d): %v", resp.StatusCode, m)
}
