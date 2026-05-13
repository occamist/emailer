package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/occamist/emailer"
	"github.com/occamist/emailer/brevo"
	"github.com/occamist/emailer/resend"
	"github.com/occamist/emailer/sendgrid"
)

var debugEnabled = flag.Bool("debug", false, "in debug environment")

const (
	defaultPort = "5555"
	// Providers Listed below
	providerBrevo    = "brevo"
	providerResend   = "resend"
	providerSendgrid = "sendgrid"
)

func main() {
	flag.Parse()

	var logHandler slog.Handler
	if *debugEnabled {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		})
	} else {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
	slog.SetDefault(slog.New(logHandler))

	key, ok := os.LookupEnv("API_KEY")
	if !ok {
		slog.Error("API_KEY not found in env")
		os.Exit(1)
	}
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = defaultPort
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		slog.LogAttrs(context.Background(), slog.LevelError, fmt.Sprintf("strconv.Atoi(%q)", port), slog.String("err", err.Error()))
		os.Exit(1)
	}
	provider, ok := os.LookupEnv("PROVIDER")
	if !ok {
		provider = providerBrevo
	}

	c := retryablehttp.NewClient()
	c.RetryMax = 3
	httpClient := c.StandardClient()
	httpClient.Timeout = 10 * time.Second

	ctx := context.Background()
	var sender emailer.Sender
	cfg := emailer.Config{Key: key, Client: *httpClient}
	switch {
	case strings.EqualFold(provider, providerBrevo):
		slog.LogAttrs(ctx, slog.LevelDebug, "brevo.New()")
		sender, err = brevo.New(cfg)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "brevo.New()", slog.String("err", err.Error()))
		}
	case strings.EqualFold(provider, providerResend):
		slog.LogAttrs(ctx, slog.LevelDebug, "resend.New()")
		sender, err = resend.New(cfg)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "resend.New()", slog.String("err", err.Error()))
		}
	case strings.EqualFold(provider, providerSendgrid):
		slog.LogAttrs(ctx, slog.LevelDebug, "sendgrid.New()")
		sender, err = sendgrid.New(cfg)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "sendgrid.New()", slog.String("err", err.Error()))
		}
	default:
		slog.LogAttrs(ctx, slog.LevelError, "unknown provider", slog.String("provider", provider))
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /email", emailer.HandlerFunc(sender))
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", portNum),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		slog.Debug(fmt.Sprintf("server started at port %d", portNum))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.LogAttrs(context.Background(), slog.LevelError, "srv.ListenAndServe()", slog.String("err", err.Error()))
		}
	}()

	wait := make(chan os.Signal, 1)
	signal.Notify(wait, os.Interrupt)
	<-wait

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.LogAttrs(context.Background(), slog.LevelError, "srv.Shutdown()", slog.String("err", err.Error()))
	}
}
