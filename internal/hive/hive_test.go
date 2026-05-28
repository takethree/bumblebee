package hive

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeEnvironment(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{name: "default", in: "", want: "production"},
		{name: "production", in: " Production ", want: "production"},
		{name: "test", in: "TEST", want: "test"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeEnvironment(tc.in)
			if err != nil {
				t.Fatalf("NormalizeEnvironment returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeEnvironment(%q)=%q, want %q", tc.in, got, tc.want)
			}
		})
	}
	if _, err := NormalizeEnvironment("dev"); err == nil {
		t.Fatal("NormalizeEnvironment accepted invalid environment")
	}
}

func TestEnrollSendsEnvironment(t *testing.T) {
	var gotBody map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/enroll" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		if r.Header.Get("CF-Access-Client-Id") != "access-id" || r.Header.Get("CF-Access-Client-Secret") != "access-secret" {
			t.Fatal("missing Access service headers")
		}
		if r.Header.Get("X-Hive-Enroll-Token") != "enroll-token" {
			t.Fatal("missing enrollment token")
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"device_id":"device-1","hmac_key":"hmac-key","ingest_path":"/v1/ingest","environment":"test"}`))
	}))
	defer server.Close()

	client := Client{
		BaseURL:            server.URL,
		AccessClientID:     "access-id",
		AccessClientSecret: "access-secret",
	}
	enrollment, err := client.Enroll("enroll-token", EnrollOptions{Environment: "test"})
	if err != nil {
		t.Fatalf("Enroll returned error: %v", err)
	}
	if gotBody["environment"] != "test" {
		t.Fatalf("environment body=%q", gotBody["environment"])
	}
	if enrollment.Environment != "test" {
		t.Fatalf("response environment=%q", enrollment.Environment)
	}
}

func TestSyncCatalogValidatesAndPromotesCurrent(t *testing.T) {
	cache := t.TempDir()
	content := `{"schema_version":"0.1.0","entries":[{"id":"adv-1","ecosystem":"npm","package":"left-pad","versions":["1.3.0"]}]}`
	sum := sha256.Sum256([]byte(content))
	sha := hex.EncodeToString(sum[:])
	bundle := CatalogBundle{
		Manifest: CatalogManifest{
			ReleaseID:    "catalog-test",
			Schema:       "0.1.0",
			FileCount:    1,
			EntryCount:   1,
			BundleSHA256: sha,
			PublishedAt:  "2026-05-27T00:00:00Z",
			Files: []CatalogManifestFile{{
				Path:       "left-pad.json",
				SHA256:     sha,
				EntryCount: 1,
			}},
		},
		Files: []CatalogFile{{
			Path:    "left-pad.json",
			SHA256:  sha,
			Content: content,
		}},
	}

	manifest, err := SyncCatalog(cache, bundle)
	if err != nil {
		t.Fatalf("SyncCatalog returned error: %v", err)
	}
	if manifest.ReleaseID != "catalog-test" {
		t.Fatalf("release_id=%q", manifest.ReleaseID)
	}
	if _, err := os.Stat(filepath.Join(CurrentCatalogDir(cache), "left-pad.json")); err != nil {
		t.Fatalf("current catalog file missing: %v", err)
	}
	cached, err := LoadCachedCatalog(cache)
	if err != nil {
		t.Fatalf("LoadCachedCatalog returned error: %v", err)
	}
	if cached.EntryCount != 1 {
		t.Fatalf("entry_count=%d", cached.EntryCount)
	}
}

func TestSyncCatalogRejectsHashMismatch(t *testing.T) {
	cache := t.TempDir()
	bundle := CatalogBundle{
		Manifest: CatalogManifest{
			ReleaseID:    "catalog-test",
			Schema:       "0.1.0",
			FileCount:    1,
			EntryCount:   1,
			BundleSHA256: "bad",
			Files: []CatalogManifestFile{{
				Path:       "left-pad.json",
				SHA256:     "bad",
				EntryCount: 1,
			}},
		},
		Files: []CatalogFile{{
			Path:    "left-pad.json",
			SHA256:  "bad",
			Content: `{"schema_version":"0.1.0","entries":[]}`,
		}},
	}

	if _, err := SyncCatalog(cache, bundle); err == nil {
		t.Fatal("SyncCatalog succeeded with mismatched hash")
	}
	if _, err := os.Stat(CurrentCatalogDir(cache)); err == nil {
		t.Fatal("current catalog directory was promoted after failed sync")
	}
}
