package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/emersion/go-smtp"
	"go.bengfort.dev/wilson/pkg"
	"go.bengfort.dev/wilson/pkg/config"
	"go.rtnl.ai/x/probez"
	"go.rtnl.ai/x/rlog"
	"go.rtnl.ai/x/rlog/console"
)

type Server struct {
	sync.RWMutex
	probez.Handler
	conf    config.Config
	http    *http.Server
	smtp    *smtp.Server
	errc    chan error
	started time.Time
}

func New() (s *Server, err error) {
	// Create a new server instance
	s = &Server{errc: make(chan error, 1)}

	// Load the configuration from the environment
	if s.conf, err = config.Get(); err != nil {
		return nil, err
	}

	// Configure logging
	// TODO: Move to telemetry
	rlog.SetLevel(s.conf.GetLogLevel())
	opts := &console.Options{
		HandlerOptions: rlog.MergeWithCustomLevels(rlog.WithGlobalLevel(nil)),
	}

	// Create the stdout handler: text or JSON.
	var stdout slog.Handler
	if s.conf.ConsoleLog {
		stdout = console.New(os.Stdout, opts)
	} else {
		stdout = slog.NewJSONHandler(os.Stdout, opts.HandlerOptions)
	}

	// Create the root logger and set it as the default rlog logger.
	rlog.SetDefault(rlog.New(slog.New(stdout)))

	// Configure the SMTP server
	s.smtp = smtp.NewServer(s)
	s.smtp.Addr = s.conf.BindAddr
	s.smtp.Domain = s.conf.Domain
	s.smtp.ReadTimeout = s.conf.ReadTimeout
	s.smtp.WriteTimeout = s.conf.WriteTimeout
	s.smtp.MaxMessageBytes = s.conf.MaxMessageSize
	s.smtp.MaxRecipients = s.conf.MaxRecipients

	if s.conf.UseTLS {
		var certs tls.Certificate
		if certs, err = tls.LoadX509KeyPair(s.conf.TLSCerts, s.conf.TLSKey); err != nil {
			return nil, fmt.Errorf("failed to load TLS certificates: %w", err)
		}

		s.smtp.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{certs},
			MinVersion:   tls.VersionTLS12,
		}

		// Explicitly demand TLS if you want to restrict incoming mail to secure-only,
		// though note that some old senders might fail if you strictly forbid unencrypted HELO.
		s.smtp.EnableREQUIRETLS = true
	}

	// Configure the HTTP server
	s.http = &http.Server{
		Addr:              ":80",
		Handler:           s,
		ErrorLog:          nil,
		ReadHeaderTimeout: 4 * time.Second,
		WriteTimeout:      16 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	return s, nil
}

func (s *Server) Serve() (err error) {
	// Catch OS signals for graceful shutdowns
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		s.errc <- s.Shutdown()
	}()

	// After rlog.Fatal, run graceful shutdown then exit.
	// NOTE: this is done after the server is fully initialized.
	rlog.SetFatalHook(func() {
		if err := s.Shutdown(); err != nil {
			rlog.ErrorAttrs(context.Background(), "error(s) during fatal shutdown of endeavor server", slog.Any("err", err))
		}
		os.Exit(1)
	})

	s.started = time.Now()
	s.Healthy()

	// Listen for HTTP request and handle them
	go func() {
		if serr := s.http.ListenAndServe(); !errors.Is(serr, http.ErrServerClosed) {
			s.errc <- serr
		} else {
			s.errc <- nil
		}
	}()

	// Listen for SMTP connections and handle them
	go func() {
		if serr := s.smtp.ListenAndServe(); !errors.Is(serr, smtp.ErrServerClosed) {
			s.errc <- serr
		} else {
			s.errc <- nil
		}
	}()

	// Now the server is live and ready
	s.Ready()

	rlog.Info("wilson service started",
		slog.String("version", pkg.Version(false)),
	)

	// Wait for an error to occur
	return <-s.errc
}

func (s *Server) Shutdown() (err error) {
	rlog.Info("gracefully shutting down wilson service")
	s.NotReady()
	defer s.Unhealthy()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if serr := s.smtp.Shutdown(ctx); serr != nil {
		err = errors.Join(err, fmt.Errorf("failed to shutdown SMTP server: %w", serr))
	}

	s.http.SetKeepAlivesEnabled(false)
	if serr := s.http.Shutdown(ctx); serr != nil {
		err = errors.Join(err, fmt.Errorf("failed to shutdown HTTP server: %w", serr))
	}

	return err
}

// NewSession is called when an SMTP client connects to the server.
func (s *Server) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{}, nil
}
