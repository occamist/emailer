# Emailer

[![Version](https://img.shields.io/github/tag/occamist/emailer.svg)](https://github.com/occamist/emailer/tags)
[![CI Build](https://github.com/occamist/emailer/actions/workflows/test.yaml/badge.svg)](https://github.com/occamist/emailer/actions/workflows/test.yaml)
[![GoDoc](https://godoc.org/github.com/occamist/emailer?status.svg)](https://godoc.org/github.com/occamist/emailer)
[![Report Card](https://goreportcard.com/badge/github.com/occamist/emailer)](https://goreportcard.com/report/github.com/occamist/emailer)
[![License](https://img.shields.io/github/license/occamist/emailer)](https://github.com/occamist/emailer/blob/main/LICENSE)
[![Codecov](https://codecov.io/github/occamist/emailer/graph/badge.svg?token=Q1CKXYDTBJ)](https://codecov.io/github/occamist/emailer)

## Purpose

Packaging SMTP(send only) APIs to a choosable option is the goal here, then to have a tiny server that picks 1 provider and serves
as HTTP endpoint so the other microservices that belong to you can make email requests.

## Supported providers

- [X] Brevo (highly recommended)
- [X] Resend
- [ ] Postmark
- [ ] Mailchimp
- [ ] Mailtrap
- [ ] Mailjet
- [ ] Mailgun
- [ ] Fastmail
- [X] Sendgrid

Note: Anything that only uses oauth2 like zoho does will not be implemented here for foreseeable future

## Getting Started

```shell
  export API_KEY=<SECRET_API_KEY>
  export PROVIDER=brevo
  go run ./cmd/emailer/main.go -debug
```

Kick the server by after having `PROVIDER` and `API_KEY` env variables then run `go run ./cmd/emailer/main.go`

or run with Docker

```shell
  docker run -d \
    -p 2345:5555 \
    -e API_KEY=<YOUR_SECRET_KEY> \
    -e PROVIDER=brevo \
    ghcr.io/occamist/emailer:latest
```

## API

This endpoint allows you to send an email by providing the necessary email details in the request body as JSON.

- Method: POST
- URL: /email
- Request: 
  - ```json
    {
      "from": "sender@example.com",
      "to": ["recipient1@example.com", "recipient2@example.com"],
      "bcc": ["bcc1@example.com", "bcc2@example.com"],
      "cc": ["cc1@example.com", "cc2@example.com"],
      "subject": "Test Email",
      "htmlContent": "<p>This is a test email in HTML format.</p>",
      "textContent": "This is a test email in plain text format."
    }
    ```
- Response:
  - 200 `Email successfully sent`
  - 400 `Encoding error` or `Failed to validate`
  - 500 `Failed to send email` (check logs something went wrong with the provider)

```shell
  curl -X POST http://localhost:5555/email \
  -H "Content-Type: application/json" \
  -d '{
    "from": "sender@example.com",
    "to": ["recipient@example.com"],
    "subject": "Test Email",
    "htmlContent": "<p>This is a test email in HTML format.</p>",
    "textContent": "This is a test email in plain text format."
  }'
```
