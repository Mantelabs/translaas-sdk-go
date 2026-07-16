package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/acuencadev/translaas-sdk-go/models"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cli, err := New(Options{
		APIKey:  "test-api-key",
		BaseURL: srv.URL,
	}, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return cli
}

func TestNew_ValidationError(t *testing.T) {
	t.Parallel()
	_, err := New(Options{})
	var cfgErr *models.ConfigurationError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("expected ConfigurationError, got %T", err)
	}
}

func TestGetEntry_Success(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "test-api-key" {
			t.Fatalf("X-Api-Key = %q", r.Header.Get("X-Api-Key"))
		}
		if accept := r.Header.Get("Accept"); accept != "text/plain" {
			t.Fatalf("Accept = %q", accept)
		}
		if !strings.HasSuffix(r.URL.Path, "/sdk/v1/translations/text") {
			t.Fatalf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("group") != "ui" || q.Get("entry") != "greeting" || q.Get("lang") != "en" {
			t.Fatalf("query = %v", q)
		}
		w.Header().Set("ETag", `W/"abc"`)
		_, _ = io.WriteString(w, "Hello, World!")
	})

	got, err := cli.GetEntry(context.Background(), "ui", "greeting", "en")
	if err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}
	if got != "Hello, World!" {
		t.Fatalf("got %q", got)
	}
}

func TestGetEntry_WithNumber(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("N") != "5" {
			t.Fatalf("N = %q", r.URL.Query().Get("N"))
		}
		_, _ = io.WriteString(w, "5 items")
	})

	got, err := cli.GetEntry(context.Background(), "ui", "items", "en", WithNumber(5))
	if err != nil || got != "5 items" {
		t.Fatalf("GetEntry() = (%q, %v)", got, err)
	}
}

func TestGetEntry_WithDecimalNumber(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("N") != "1.31" {
			t.Fatalf("N = %q", r.URL.Query().Get("N"))
		}
		_, _ = io.WriteString(w, "ok")
	})

	_, err := cli.GetEntry(context.Background(), "ui", "items", "en", WithNumber(1.31))
	if err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}
}

func TestGetEntry_WithParameters(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("userName") != "John Doe" {
			t.Fatalf("userName = %q", q.Get("userName"))
		}
		_, _ = io.WriteString(w, "hi")
	})

	_, err := cli.GetEntry(context.Background(), "ui", "greeting", "en", WithParameters(map[string]string{
		"userName": "John Doe",
	}))
	if err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}
}

func TestGetEntry_WithRequestContext(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("channel") != "beta" || q.Get("v") != "snap-1" || q.Get("project") != "proj-1" {
			t.Fatalf("query = %v", q)
		}
		if r.Header.Get("If-None-Match") != `"etag-1"` {
			t.Fatalf("If-None-Match = %q", r.Header.Get("If-None-Match"))
		}
		_, _ = io.WriteString(w, "ok")
	})

	ctx := &models.RequestContext{
		Channel:     "beta",
		Version:     "snap-1",
		Project:     "proj-1",
		IfNoneMatch: `"etag-1"`,
	}
	_, err := cli.GetEntry(context.Background(), "ui", "greeting", "en", WithRequestContext(ctx))
	if err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}
	if ctx.ResponseETag != "" && ctx.NotModified {
		t.Fatal("unexpected not modified on 200")
	}
}

func TestGetEntry_NoContent(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	got, err := cli.GetEntry(context.Background(), "ui", "missing.entry", "en")
	if err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}
	if got != "missing.entry" {
		t.Fatalf("got %q, want entry key fallback", got)
	}
}

func TestGetEntry_NotModified(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `W/"xyz"`)
		w.WriteHeader(http.StatusNotModified)
	})

	reqCtx := &models.RequestContext{}
	got, err := cli.GetEntry(context.Background(), "ui", "greeting", "en", WithRequestContext(reqCtx))
	if err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}
	if got != "" {
		t.Fatalf("got %q, want empty string", got)
	}
	if !reqCtx.NotModified {
		t.Fatal("expected NotModified")
	}
	if reqCtx.ResponseETag != `W/"xyz"` {
		t.Fatalf("ResponseETag = %q", reqCtx.ResponseETag)
	}
}

func TestGetEntry_APIErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		status     int
		body       string
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "unauthorized",
			status:     http.StatusUnauthorized,
			body:       `{"message":"invalid key","code":"AUTH"}`,
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "[AUTH] invalid key",
		},
		{
			name:       "server error",
			status:     http.StatusInternalServerError,
			body:       "Internal Server Error",
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "API request failed with status code 500.",
		},
		{
			name:       "not found",
			status:     http.StatusNotFound,
			body:       "Not Found",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			})

			_, err := cli.GetEntry(context.Background(), "ui", "entry", "en")
			var apiErr *models.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected APIError, got %T (%v)", err, err)
			}
			if apiErr.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", apiErr.StatusCode, tt.wantStatus)
			}
			if tt.wantMsg != "" && apiErr.Message != tt.wantMsg {
				t.Fatalf("message = %q, want %q", apiErr.Message, tt.wantMsg)
			}
		})
	}
}

func TestGetEntry_ValidationErrors(t *testing.T) {
	t.Parallel()
	cli := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected HTTP call")
	})

	_, err := cli.GetEntry(context.Background(), "", "entry", "en")
	if err == nil || !strings.Contains(err.Error(), "group") {
		t.Fatalf("expected group error, got %v", err)
	}
}

func TestGetEntry_ContextCanceled(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = io.WriteString(w, "late")
	}))
	t.Cleanup(srv.Close)

	cli, err := New(Options{
		APIKey:  "test-api-key",
		BaseURL: srv.URL,
		Timeout: time.Second,
	}, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = cli.GetEntry(ctx, "ui", "entry", "en")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestGetEntry_Timeout(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		_, _ = io.WriteString(w, "late")
	}))
	t.Cleanup(srv.Close)

	cli, err := New(Options{
		APIKey:  "test-api-key",
		BaseURL: srv.URL,
		Timeout: 50 * time.Millisecond,
	}, WithHTTPClient(&http.Client{Timeout: 50 * time.Millisecond}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = cli.GetEntry(context.Background(), "ui", "entry", "en")
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusRequestTimeout {
		t.Fatalf("status = %d, want 408", apiErr.StatusCode)
	}
}

func TestGetEntry_DefaultProjectID(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("project") != "default-proj" {
			t.Fatalf("project = %q", r.URL.Query().Get("project"))
		}
		_, _ = io.WriteString(w, "ok")
	}))
	t.Cleanup(srv.Close)

	cli, err := New(Options{
		APIKey:           "test-api-key",
		BaseURL:          srv.URL,
		DefaultProjectID: "default-proj",
	}, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = cli.GetEntry(context.Background(), "ui", "entry", "en")
	if err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}
}

func TestHandleAPIError_Envelope(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(models.TranslaasError{Message: "missing", Code: "NOT_FOUND"})
	err := handleAPIError(http.StatusNotFound, body)
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError")
	}
	if apiErr.Message != "[NOT_FOUND] missing" {
		t.Fatalf("message = %q", apiErr.Message)
	}
}
