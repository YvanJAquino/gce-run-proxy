package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"google.golang.org/api/idtoken"
)

type HealthChecker struct {
	h http.Handler
}

func NewHealthChecker(h http.Handler) *HealthChecker {
	return &HealthChecker{
		h: h,
	}
}

func (c HealthChecker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/health/check" {
		c.checkHealth(w, r)
	} else {
		c.h.ServeHTTP(w, r)
	}
}

func (c HealthChecker) checkHealth(w http.ResponseWriter, r *http.Request) {

	err := json.NewEncoder(w).Encode(StatusOK)
	if err != nil {
		log.Println(err)
	}
	w.WriteHeader(http.StatusOK)
}

type AuthorizationVerifier struct {
	// rp *httputil.ReverseProxy
	h  http.Handler
	v  *idtoken.Validator
	vs map[string]any
}

func NewAuthorizationVerifier(ctx context.Context, h http.Handler) (*AuthorizationVerifier, error) {
	pm := new(AuthorizationVerifier)
	v, err := idtoken.NewValidator(ctx)
	if err != nil {
		return nil, err
	}

	// pm.rp = rp
	pm.h = h
	pm.v = v
	pm.vs = make(map[string]any)
	return pm, nil
}

func (pm *AuthorizationVerifier) VerifyClaim(claim string, value any) {
	pm.vs[claim] = value
}

func (pm *AuthorizationVerifier) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	header := r.Header.Get(AuthorizationHeader)
	if header == "" {
		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(Forbidden)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	token := pm.extractToken(header)
	ok, err := pm.checkToken(r.Context(), token, "")
	if err != nil {
		log.Println(err)
		return
	}
	if !ok {
		log.Println("things didn't match bro")
		return
	}
	// pm.rp.ServeHTTP(w, r)
	pm.h.ServeHTTP(w, r)

}

func (pm AuthorizationVerifier) extractToken(header string) string {
	authHeaderParts := strings.Split(header, " ")
	return authHeaderParts[1]
}

func (pm AuthorizationVerifier) checkToken(ctx context.Context, token, aud string) (bool, error) {
	payload, err := pm.v.Validate(ctx, token, aud)
	if err != nil {
		return false, err
	}
	ok := true
	for key, val := range payload.Claims {
		vval, vok := pm.vs[key]
		if vok {
			log.Printf("Key: %s, Match?: %v, Claim Value: %v, Verifier Value, %v\n", key, val == vval, val, vval)
		}
		if vok && vval != val {
			ok = false
			return ok, nil
		}
	}
	return ok, nil
}
