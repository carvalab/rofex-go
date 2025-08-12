package rofex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type testServerState struct {
	loginCalls int32
	goodToken  atomic.Value // string
}

func newTestServer(t *testing.T) (*httptest.Server, *testServerState) {
	st := &testServerState{}
	st.goodToken.Store("token-initial")

	h := http.NewServeMux()
	h.HandleFunc("/auth/getToken", func(w http.ResponseWriter, r *http.Request) {
		// Always succeed and increment token
		calls := atomic.AddInt32(&st.loginCalls, 1)
		var token string
		if calls == 1 {
			token = "token-1"
		} else {
			token = "token-2"
		}
		st.goodToken.Store(token)
		w.Header().Set("X-Auth-Token", token)
		w.WriteHeader(http.StatusOK)
	})
	h.HandleFunc("/rest/segment/all", func(w http.ResponseWriter, r *http.Request) {
		want := st.goodToken.Load().(string)
		if r.Header.Get("X-Auth-Token") != want {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "OK",
			"segments": []map[string]any{
				{"marketSegmentId": "DDA", "marketId": "ROFX"},
				{"marketSegmentId": "DUAL", "marketId": "ROFX"},
			},
		})
	})

	ts := httptest.NewServer(h)
	return ts, st
}

func TestLoginAndSegments_With401Refresh(t *testing.T) {
	// constants
	loginPath := "/auth/getToken"
	segmentsPath := "/rest/segment/all"
	unauthCode := http.StatusUnauthorized

	ts, st := newTestServer(t)
	defer ts.Close()

	c, err := NewClient(
		WithBaseURL(ts.URL+"/"),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Login sets PasswordAuth and token-1
	if err := c.Login(ctx, Credentials{Username: "u", Password: "p"}); err != nil {
		t.Fatalf("login: %v", err)
	}

	// Force server to expect token-2 so the next GET causes 401 and triggers refresh
	st.goodToken.Store("must-refresh")

	res, err := c.Segments(ctx)
	if err != nil {
		t.Fatalf("segments after refresh: %v", err)
	}
	// Basic check (typed)
	if len(res.Segments) == 0 {
		t.Fatalf("no segments returned")
	}

	// Ensure login was called at least twice (initial + refresh)
	if atomic.LoadInt32(&st.loginCalls) < 2 {
		t.Fatalf("expected at least 2 login calls, got %d", st.loginCalls)
	}

	_ = loginPath
	_ = segmentsPath
	_ = unauthCode
}
