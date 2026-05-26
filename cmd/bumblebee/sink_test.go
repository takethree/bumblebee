package main

import (
	"strings"
	"testing"
)

func TestBuildHTTPHeadersFromEnv(t *testing.T) {
	t.Setenv("BUMBLEBEE_TEST_HEADER", "secret-value")

	headers, err := buildHTTPHeaders([]string{"X-Test-Header=BUMBLEBEE_TEST_HEADER"})
	if err != nil {
		t.Fatal(err)
	}
	if got := headers["X-Test-Header"]; got != "secret-value" {
		t.Fatalf("header value = %q, want secret-value", got)
	}
}

func TestBuildHTTPHeadersRejectsBadShape(t *testing.T) {
	_, err := buildHTTPHeaders([]string{"X-Test-Header"})
	if err == nil || !strings.Contains(err.Error(), "Header-Name=ENV_VAR") {
		t.Fatalf("expected shape error, got %v", err)
	}
}

func TestBuildHTTPHeadersRejectsMissingEnv(t *testing.T) {
	_, err := buildHTTPHeaders([]string{"X-Test-Header=BUMBLEBEE_TEST_MISSING"})
	if err == nil || !strings.Contains(err.Error(), "BUMBLEBEE_TEST_MISSING") {
		t.Fatalf("expected missing env error, got %v", err)
	}
}
