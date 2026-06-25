// Package resend makes it easy to send emails via resend provider. This package follows [resend spec] strictly.
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
// [resend spec]: https://resend.com/docs/api-reference/emails/send-email
package resend

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

const endpoint = "https://api.resend.com/emails"

// EmailClient is resend email client to interact with emails
type EmailClient struct {
	key    string
	client http.Client
}

// New creates a new resend email client with given API key and http.Client
func New(c emailer.Config) (*EmailClient, error) {
	if strings.TrimSpace(c.Key) == "" {
		return nil, errors.New("resend API key is blank")
	}
	e := &EmailClient{
		key:    c.Key,
		client: c.Client,
	}
	return e, nil
}

// payload is a request that resend uses to send email
type payload struct {
	From        string   `json:"from"`
	To          []string `json:"to"`
	BCC         []string `json:"bcc"`
	CC          []string `json:"cc"`
	Subject     string   `json:"subject"`
	HTMLContent string   `json:"html,omitempty"`
	TextContent string   `json:"text,omitempty"`
}

type errorMessage struct {
	Message    string `json:"message"`
	Name       string `json:"name"`
	StatusCode int    `json:"statusCode"`
}

// Send sends a given email
func (c *EmailClient) Send(ctx context.Context, email emailer.Email) error {
	var p payload
	p.From = email.From
	p.To = append(p.To, email.To...)
	p.BCC = append(p.BCC, email.BCC...)
	p.CC = append(p.CC, email.CC...)
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
	req.Header.Add("Authorization", "Bearer "+c.key)
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	resp, err := c.client.Do(req)
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
