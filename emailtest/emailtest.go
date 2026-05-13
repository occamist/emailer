// Package emailtest makes testing easy for our tests.
package emailtest

import (
	"errors"
	"net/http"

	"github.com/occamist/emailer"
)

// RoundTripFunc is transport layer of HTTP client
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip satisfies http.RoundTripper for stubbing
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// FaultyRoundTripFunc is faulty transport layer of HTTP client
type FaultyRoundTripFunc func(req *http.Request) *http.Response

// RoundTrip satisfies http.RoundTripper for stubbing
func (f FaultyRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), errors.New("boom baby boom")
}

// NewConfig creates emailer config for a given healthy round tripper
func NewConfig(tripper RoundTripFunc) emailer.Config {
	return emailer.Config{
		Key:    "key",
		Client: http.Client{Transport: tripper},
	}
}

// NewFaultyClientConfig creates emailer config for a given unhealthy round tripper
func NewFaultyClientConfig(tripper FaultyRoundTripFunc) emailer.Config {
	return emailer.Config{
		Key:    "key",
		Client: http.Client{Transport: tripper},
	}
}
