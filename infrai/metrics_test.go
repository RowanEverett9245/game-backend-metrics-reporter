package infrai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReportSendsMetricEnvelopeAndBearerKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/metrics/report" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("authorization header = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"data":{"accepted":true},"error":null,"metadata":{}}`))
	}))
	defer server.Close()
	client := &Client{Key: "test-key", HTTPClient: server.Client()}
	oldBase := baseURL
	_ = oldBase
	// Report uses the production endpoint; this focused test verifies envelope parsing helpers through a local transport.
	transport := server.Client().Transport
	client.HTTPClient = &http.Client{Transport: rewriteTransport{base: server.URL, next: transport}}
	if err := client.Report(context.Background(), Metric{Type: "counter", Name: "game.matches", Value: 1}); err != nil {
		t.Fatal(err)
	}
}

func TestReportRejectsNon2xxSuccessfulEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"ok":true,"data":{"accepted":true},"error":null,"metadata":{}}`))
	}))
	defer server.Close()

	client := &Client{
		Key:        "test-key",
		HTTPClient: &http.Client{Transport: rewriteTransport{base: server.URL, next: server.Client().Transport}},
	}
	if err := client.Report(context.Background(), Metric{Type: "counter", Name: "game.matches", Value: 1}); err == nil {
		t.Fatal("expected non-2xx response to be rejected")
	}
}

type rewriteTransport struct {
	base string
	next http.RoundTripper
}

func (t rewriteTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	copy := r.Clone(r.Context())
	copy.URL.Scheme = "http"
	copy.URL.Host = t.base[len("http://"):]
	return t.next.RoundTrip(copy)
}
