package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Mantelabs/translaas-sdk-go/models"
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
	var transportErr *models.TransportError
	if !errors.As(err, &transportErr) {
		t.Fatalf("expected TransportError, got %T (%v)", err, err)
	}
	var apiErr *models.APIError
	if errors.As(err, &apiErr) {
		t.Fatalf("transport failure must not be APIError, got status %d", apiErr.StatusCode)
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Fatalf("message = %q", err.Error())
	}
}

func TestGetEntry_TransportError_Table(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
	}{
		{name: "connection refused", err: errors.New("connection refused")},
		{name: "dns", err: errors.New("no such host")},
		{name: "tls", err: errors.New("tls: failed to verify certificate")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cause := tc.err
			cli, err := New(Options{
				APIKey:  "test-api-key",
				BaseURL: "https://api.test.com",
			}, WithHTTPClient(&http.Client{
				Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
					return nil, cause
				}),
			}))
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			_, err = cli.GetEntry(context.Background(), "ui", "entry", "en")
			var transportErr *models.TransportError
			if !errors.As(err, &transportErr) {
				t.Fatalf("expected TransportError, got %T (%v)", err, err)
			}
			var apiErr *models.APIError
			if errors.As(err, &apiErr) {
				t.Fatalf("got APIError status %d", apiErr.StatusCode)
			}
		})
	}
}

func TestNew_InsecureSkipVerify_ConfiguresTLS(t *testing.T) {
	t.Parallel()

	cli, err := New(Options{
		APIKey:             "key",
		BaseURL:            "https://api.test.com",
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	impl := cli.(*client)
	transport, ok := impl.httpClient.Transport.(*http.Transport)
	if !ok || transport == nil || transport.TLSClientConfig == nil {
		t.Fatalf("expected cloned http.Transport with TLS config, got %#v", impl.httpClient.Transport)
	}
	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("expected InsecureSkipVerify on built-in transport")
	}
}

func TestNew_InsecureSkipVerify_WithHTTPClientUnchanged(t *testing.T) {
	t.Parallel()

	custom := &http.Client{}
	cli, err := New(Options{
		APIKey:             "key",
		BaseURL:            "https://api.test.com",
		InsecureSkipVerify: true,
	}, WithHTTPClient(custom))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	impl := cli.(*client)
	if impl.httpClient != custom {
		t.Fatal("WithHTTPClient must win over InsecureSkipVerify")
	}
	if custom.Transport != nil {
		t.Fatal("must not attach a skip-verify transport to the injected client")
	}
}

func TestNew_DefaultLanguageTrimmed(t *testing.T) {
	t.Parallel()

	cli, err := New(Options{
		APIKey:          "key",
		BaseURL:         "https://api.test.com",
		DefaultLanguage: "  de  ",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	impl := cli.(*client)
	if impl.DefaultLanguage() != "de" {
		t.Fatalf("DefaultLanguage = %q, want de", impl.DefaultLanguage())
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
