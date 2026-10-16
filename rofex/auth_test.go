package rofex

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Many goroutines hitting the API at once (e.g. a bot polling several symbols)
// must share one login instead of each one logging in.
func TestPasswordAuth_ConcurrentRequestsLoginOnce(t *testing.T) {
	var logins atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/getToken" {
			logins.Add(1)
			time.Sleep(50 * time.Millisecond) // make the race window realistic
			w.Header().Set("X-Auth-Token", "tok")
			return
		}
		if r.Header.Get("X-Auth-Token") != "tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "OK"})
	}))
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL+"/"), WithAuth(NewPasswordAuth(Credentials{Username: "u", Password: "p"})))
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			if _, err := c.Segments(context.Background()); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if n := logins.Load(); n != 1 {
		t.Errorf("logins = %d, want 1", n)
	}
}

// Wrong credentials must surface as an error, not loop or hang.
func TestPasswordAuth_BadCredentials(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL+"/"), WithAuth(NewPasswordAuth(Credentials{Username: "u", Password: "bad"})))
	_, err := c.Accounts(context.Background())
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("want HTTP 401 error, got %v", err)
	}
}
