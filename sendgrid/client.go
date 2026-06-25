// Package sendgrid makes it easy to send emails via sendgrid provider. This package follows [sendgrid spec] strictly.
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
// [sendgrid spec]: https://www.twilio.com/docs/sendgrid/api-reference/mail-send/mail-send
package sendgrid

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

const endpoint = "https://api.sendgrid.com/v3/mail/send"

// EmailClient is sendgrid email client to interact with emails
type EmailClient struct {
	key    string
	client http.Client
}

// New creates a new sendgrid email client with given API key and http.Client
func New(c emailer.Config) (*EmailClient, error) {
	if strings.TrimSpace(c.Key) == "" {
		return nil, errors.New("sendgrid API key is blank")
	}
	e := &EmailClient{
		key:    c.Key,
		client: c.Client,
	}
	return e, nil
}

type emailObject struct {
	Email string `json:"email"`
}

type personalization struct {
	To  []emailObject `json:"to"`
	BCC []emailObject `json:"bcc,omitempty"`
	CC  []emailObject `json:"cc,omitempty"`
}

type content struct {
	// Type is MIME Type (e.g., text/plain or text/html)
	Type string `json:"type"`
	// Value is the actual content
	Value string `json:"value"`
}

// payload is a request that sendgrid uses to send email
type payload struct {
	Personalizations []personalization `json:"personalizations"`
	From             emailObject       `json:"from"`
	Subject          string            `json:"subject"`
	Content          []content         `json:"content,omitempty"`
}

type errorMessage struct {
	Errors []struct {
		Message string `json:"message"`
		Field   string `json:"field"`
		Help    any    `json:"help"`
	} `json:"errors"`
}

// Send sends a given email
func (c *EmailClient) Send(ctx context.Context, email emailer.Email) error {
	var p payload
	p.From.Email = email.From

	pers := personalization{}
	for _, e := range email.To {
		pers.To = append(pers.To, emailObject{Email: e})
	}
	for _, e := range email.BCC {
		pers.BCC = append(pers.BCC, emailObject{Email: e})
	}
	for _, e := range email.CC {
		pers.CC = append(pers.CC, emailObject{Email: e})
	}
	p.Personalizations = append(p.Personalizations, pers)
	p.Subject = email.Subject

	if email.TextContent != "" {
		p.Content = append(p.Content, content{Type: "text/plain", Value: email.TextContent})
	}
	if email.HTMLContent != "" {
		p.Content = append(p.Content, content{Type: "text/html", Value: email.HTMLContent})
	}

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
