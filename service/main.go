package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/api/idtoken"
)

var (
	PORT      = os.Getenv("PORT")
	UPSTREAM  = os.Getenv("UPSTREAM")
	PRINCIPAL = os.Getenv("PRINCIPAL")
)

func main() {
	// Create a root context
	ctx := context.Background()

	// Create dummy x509 key and certificate
	tlsMgr, tlsCloser := NewSelfSignedTLS()
	defer tlsCloser()

	// Parse the upstream URL into it's *url.URL format
	URL, err := url.Parse(UPSTREAM)
	if err != nil {
		log.Fatal(err)
	}

	// Create a token Source based on the Upstream's URL
	// URL.string() is used here to ensure consistency
	ts, err := idtoken.NewTokenSource(ctx, URL.String())
	if err != nil {
		log.Fatal(err)
	}

	// Create a ReverseProxy instance.
	reverseProxy := httputil.NewSingleHostReverseProxy(URL)
	// Store the director for reuse
	director := reverseProxy.Director
	// Apply post-processing to the original directory
	reverseProxy.Director = func(r *http.Request) {
		director(r)

		// Set the Host Header and the Host property for Cloud Run
		// Cloud Run requires this Host-based routing.
		r.Header.Set("X-Forwarded-Proto", "https")
		r.Header.Set("Host", URL.Host)
		r.Host = URL.Host

		// Create a new token
		token, err := ts.Token()
		if err != nil {
			log.Fatal(err)
		}
		// Apply the new token
		token.SetAuthHeader(r)
	}

	// Create a New Verifier.
	verifier, err := NewAuthorizationVerifier(ctx, reverseProxy)
	if err != nil {
		log.Fatal(err)
	}
	// Verify the calling principal.  You can add more verifiers
	// which will inspect the claims for matching purposes.
	verifier.VerifyClaim("email", PRINCIPAL)

	// HealthChecker middleware.  Allows for health checks over HTTP(S)
	// specify the path /health/check for health checks.
	checker := NewHealthChecker(verifier)

	// Create a notification channel and listen for SIGTERM, SIGINT
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	// Create a new HTTP(S) server
	server := &http.Server{
		Addr:        ":" + PORT,
		Handler:     checker,
		BaseContext: func(l net.Listener) context.Context { return ctx },
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}
	log.Printf("Proxying traffic for %v", URL)

	// Run the server on a separate goroutine so that main
	// can listen for signals.
	go func() {
		log.Printf("Listening and serving HTTP on %s\n", ":"+PORT)
		err := server.ListenAndServeTLS(tlsMgr.CertFilename(), tlsMgr.KeyFilename())
		if err != nil && err != http.ErrServerClosed {
			log.Println(err)
		}
	}()

	// Listen for "termination" signals.
	select {
	case <-ctx.Done():
		log.Println("CONTEXT DONE")
	case sig := <-signals:
		log.Printf("%s signal received\n", sig)
		defer tlsCloser()
		shutCtx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		err = server.Shutdown(shutCtx)
		if err != nil && err != http.ErrServerClosed {
			log.Println(err)
		} else if err == http.ErrServerClosed {
			os.Exit(0)
		}

	}
}
