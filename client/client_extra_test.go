package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/acuencadev/translaas-sdk-go/models"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestNew_DefaultTimeout(t *testing.T) {
	t.Parallel()
	c, err := New(Options{
		APIKey:  "key",
		BaseURL: "https://api.test.com",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	impl := c.(*client)
	if impl.timeout != DefaultTimeout {
		t.Fatalf("timeout = %v, want %v", impl.timeout, DefaultTimeout)
	}
}

func TestGetEntry_NetworkError(t *testing.T) {
	t.Parallel()
	cli, err := New(Options{
		APIKey:  "test-api-key",
		BaseURL: "https://api.test.com",
	}, WithHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("connection refused")
		}),
	}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = cli.GetEntry(context.Background(), "ui", "entry", "en")
	var apiErr *models.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", apiErr.StatusCode)
	}
}

func TestGetEntry_ReadBodyError(t *testing.T) {
	t.Parallel()
	cli, err := New(Options{
		APIKey:  "test-api-key",
		BaseURL: "https://api.test.com",
	}, WithHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(failReader{}),
				Header:     make(http.Header),
			}, nil
		}),
	}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = cli.GetEntry(context.Background(), "ui", "entry", "en")
	if err == nil || !strings.Contains(err.Error(), "read response body") {
		t.Fatalf("expected read body error, got %v", err)
	}
}

func TestGetEntry_RequiredFieldErrors(t *testing.T) {
	t.Parallel()
	cli, err := New(Options{APIKey: "key", BaseURL: "https://api.test.com"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = cli.GetEntry(context.Background(), "ui", "", "en")
	if err == nil {
		t.Fatal("expected entry error")
	}
	_, err = cli.GetEntry(context.Background(), "ui", "entry", "")
	if err == nil {
		t.Fatal("expected lang error")
	}
}

type failReader struct{}

func (failReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func TestGetEntry_NumberNotInQueryWhenUnset(t *testing.T) {
	t.Parallel()
	var captured *http.Request
	cli, err := New(Options{
		APIKey:  "test-api-key",
		BaseURL: "https://api.test.com",
	}, WithHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			captured = r
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
				Header:     make(http.Header),
			}, nil
		}),
	}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = cli.GetEntry(context.Background(), "ui", "items", "en")
	if err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}
	q := captured.URL.Query()
	if q.Get("n") != "" || q.Get("N") != "" {
		t.Fatalf("unexpected number query: %v", q)
	}
}
