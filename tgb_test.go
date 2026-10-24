package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "tgb.yaml")
	if err := os.WriteFile(p, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadConfiguration(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := writeConfig(t, `
server:
  host: "localhost"
  port: 10001
  auth_token: "s3cret"
telegram:
  chatid: 42
  token: "bot-token"
`)
		c, err := loadConfiguration(p)
		if err != nil {
			t.Fatalf("loadConfiguration: %v", err)
		}
		if c.Server.Host != "localhost" || c.Server.Port != 10001 || c.Server.AuthToken != "s3cret" {
			t.Fatalf("unexpected server config: %+v", c.Server)
		}
		if c.Telegram.ChatID != 42 || c.Telegram.Token != "bot-token" {
			t.Fatalf("unexpected telegram config: %+v", c.Telegram)
		}
	})

	t.Run("default host", func(t *testing.T) {
		p := writeConfig(t, `
server:
  port: 10000
telegram:
  token: "bot-token"
`)
		c, err := loadConfiguration(p)
		if err != nil {
			t.Fatalf("loadConfiguration: %v", err)
		}
		if c.Server.Host != "127.0.0.1" {
			t.Fatalf("host = %q, want 127.0.0.1", c.Server.Host)
		}
	})

	for _, tc := range []struct {
		name string
		body string
		want string
	}{
		{"missing telegram token", "server:\n  port: 10000\n", "telegram.token"},
		{"port zero", "server:\n  port: 0\ntelegram:\n  token: t\n", "server.port"},
		{"port too large", "server:\n  port: 70000\ntelegram:\n  token: t\n", "server.port"},
		{"invalid yaml", "server: [\n", "parse"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := loadConfiguration(writeConfig(t, tc.body))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}

	if _, err := loadConfiguration(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Fatal("expected an error for a missing config file")
	}
}

func TestConfigPath(t *testing.T) {
	t.Setenv("TGB_CONFIG", "")
	if got := configPath(); got != defaultConfigPath {
		t.Fatalf("configPath() = %q, want %q", got, defaultConfigPath)
	}
	t.Setenv("TGB_CONFIG", "/tmp/custom-tgb.yaml")
	if got := configPath(); got != "/tmp/custom-tgb.yaml" {
		t.Fatalf("configPath() = %q, want /tmp/custom-tgb.yaml", got)
	}
}

func TestAuthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/send", nil)

	open := &handler{}
	if !open.authorized(req) {
		t.Fatal("empty auth token should allow requests")
	}

	h := &handler{authToken: "s3cret"}
	if h.authorized(req) {
		t.Fatal("request without credentials should be rejected")
	}

	req.Header.Set("Authorization", "Bearer wrong")
	if h.authorized(req) {
		t.Fatal("wrong bearer token should be rejected")
	}

	req.Header.Set("Authorization", "Bearer s3cret")
	if !h.authorized(req) {
		t.Fatal("correct bearer token should be accepted")
	}

	req.Header.Del("Authorization")
	req.Header.Set("X-Auth-Token", "s3cret")
	if !h.authorized(req) {
		t.Fatal("correct X-Auth-Token should be accepted")
	}
}

func send(t *testing.T, h *handler, body string, auth bool) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/send", strings.NewReader(body))
	if auth {
		req.Header.Set("Authorization", "Bearer s3cret")
	}
	rec := httptest.NewRecorder()
	h.sendMessage(rec, req)
	return rec
}

func TestSendMessage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var got string
		h := &handler{authToken: "s3cret", send: func(text string) error {
			got = text
			return nil
		}}
		rec := send(t, h, `{"sender":"monit","content":"disk hot"}`, true)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if got != "monit: disk hot" {
			t.Fatalf("sent = %q, want %q", got, "monit: disk hot")
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		called := false
		h := &handler{authToken: "s3cret", send: func(string) error { called = true; return nil }}
		rec := send(t, h, `{"sender":"x","content":"y"}`, false)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
		if called {
			t.Fatal("send must not be called without auth")
		}
	})

	t.Run("body too large", func(t *testing.T) {
		h := &handler{authToken: "s3cret", send: func(string) error { return nil }}
		body := `{"sender":"x","content":"` + strings.Repeat("a", maxBodyBytes) + `"}`
		rec := send(t, h, body, true)
		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("status = %d, want 413", rec.Code)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		h := &handler{authToken: "s3cret", send: func(string) error { return nil }}
		rec := send(t, h, `{`, true)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("missing fields", func(t *testing.T) {
		h := &handler{authToken: "s3cret", send: func(string) error { return nil }}
		rec := send(t, h, `{"sender":"x"}`, true)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("telegram failure", func(t *testing.T) {
		h := &handler{authToken: "s3cret", send: func(string) error {
			return errors.New("telegram is down")
		}}
		rec := send(t, h, `{"sender":"x","content":"y"}`, true)
		if rec.Code != http.StatusBadGateway {
			t.Fatalf("status = %d, want 502", rec.Code)
		}
	})

	t.Run("open mode", func(t *testing.T) {
		h := &handler{send: func(string) error { return nil }}
		rec := send(t, h, `{"sender":"x","content":"y"}`, false)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})
}

func TestRouterRejectsOtherMethods(t *testing.T) {
	h := &handler{send: func(string) error { return nil }}
	req := httptest.NewRequest(http.MethodGet, "/send", nil)
	rec := httptest.NewRecorder()
	newRouter(h).ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}
