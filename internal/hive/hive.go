package hive

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/perplexityai/bumblebee/internal/exposure"
)

const (
	ConfigFileName     = "config.json"
	SecretsFileName    = "secrets.json"
	DefaultEnvironment = "production"
	currentDirName     = "current"
	manifestName       = "manifest.meta"
)

type Config struct {
	BaseURL           string   `json:"base_url"`
	IngestPath        string   `json:"ingest_path"`
	DeviceID          string   `json:"device_id"`
	Environment       string   `json:"environment,omitempty"`
	ScanProfile       string   `json:"scan_profile,omitempty"`
	ScanRoots         []string `json:"scan_roots,omitempty"`
	CatalogReleaseID  string   `json:"catalog_release_id,omitempty"`
	CatalogBundleSHA  string   `json:"catalog_bundle_sha256,omitempty"`
	CatalogSource     string   `json:"catalog_source,omitempty"`
	CatalogPublished  string   `json:"catalog_published_at,omitempty"`
	CatalogSyncedAt   string   `json:"catalog_synced_at,omitempty"`
	CatalogEntryCount int      `json:"catalog_entry_count,omitempty"`
}

type Secrets struct {
	AccessClientID     string `json:"access_client_id"`
	AccessClientSecret string `json:"access_client_secret"`
	HMACKey            string `json:"hmac_key"`
}

type EnrollResponse struct {
	DeviceID    string `json:"device_id"`
	HMACKey     string `json:"hmac_key"`
	IngestPath  string `json:"ingest_path"`
	Environment string `json:"environment"`
}

type EnrollOptions struct {
	DeviceID    string
	Environment string
}

type CatalogBundle struct {
	Manifest CatalogManifest `json:"manifest"`
	Files    []CatalogFile   `json:"files"`
}

type CatalogManifest struct {
	ReleaseID    string                `json:"release_id"`
	Source       string                `json:"source,omitempty"`
	Schema       string                `json:"schema_version"`
	FileCount    int                   `json:"file_count"`
	EntryCount   int                   `json:"entry_count"`
	BundleSHA256 string                `json:"bundle_sha256"`
	PublishedAt  string                `json:"published_at"`
	Files        []CatalogManifestFile `json:"files"`
}

type CatalogManifestFile struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	EntryCount int    `json:"entry_count"`
}

type CatalogFile struct {
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
	Content string `json:"content"`
}

type Client struct {
	BaseURL            string
	AccessClientID     string
	AccessClientSecret string
	DeviceID           string
	HTTPClient         *http.Client
}

func DefaultConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		if v := os.Getenv("APPDATA"); v != "" {
			return filepath.Join(v, "Bumblebee")
		}
	case "darwin":
		if h := userHome(); h != "" {
			return filepath.Join(h, "Library", "Application Support", "Bumblebee")
		}
	default:
		if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
			return filepath.Join(v, "bumblebee")
		}
	}
	if h := userHome(); h != "" {
		return filepath.Join(h, ".config", "bumblebee")
	}
	return "."
}

func DefaultCacheDir() string {
	switch runtime.GOOS {
	case "windows":
		if v := os.Getenv("LOCALAPPDATA"); v != "" {
			return filepath.Join(v, "Bumblebee", "catalog-cache")
		}
	case "darwin":
		if h := userHome(); h != "" {
			return filepath.Join(h, "Library", "Caches", "Bumblebee", "catalog")
		}
	default:
		if v := os.Getenv("XDG_CACHE_HOME"); v != "" {
			return filepath.Join(v, "bumblebee", "catalog")
		}
	}
	if h := userHome(); h != "" {
		return filepath.Join(h, ".cache", "bumblebee", "catalog")
	}
	return filepath.Join(".", ".bumblebee-cache")
}

func userHome() string {
	h, _ := os.UserHomeDir()
	return h
}

func LoadConfig(dir string) (Config, Secrets, error) {
	if dir == "" {
		dir = DefaultConfigDir()
	}
	var cfg Config
	var sec Secrets
	if err := readJSON(filepath.Join(dir, ConfigFileName), &cfg); err != nil {
		return cfg, sec, err
	}
	if err := readJSON(filepath.Join(dir, SecretsFileName), &sec); err != nil {
		return cfg, sec, err
	}
	return cfg, sec, validateConfig(cfg, sec)
}

func SaveConfig(dir string, cfg Config, sec Secrets) error {
	if dir == "" {
		dir = DefaultConfigDir()
	}
	if err := validateConfig(cfg, sec); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create hive config dir: %w", err)
	}
	if err := writeJSONRestricted(filepath.Join(dir, ConfigFileName), cfg); err != nil {
		return err
	}
	if err := writeJSONRestricted(filepath.Join(dir, SecretsFileName), sec); err != nil {
		return err
	}
	return nil
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read hive config: %w", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parse hive config: %w", err)
	}
	return nil
}

func writeJSONRestricted(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("write hive config: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return fmt.Errorf("write hive config: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("write hive config: %w", err)
	}
	_ = os.Chmod(path, 0o600)
	return nil
}

func validateConfig(cfg Config, sec Secrets) error {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return errors.New("hive config missing base_url")
	}
	if _, err := url.ParseRequestURI(cfg.BaseURL); err != nil {
		return fmt.Errorf("hive config invalid base_url: %w", err)
	}
	if strings.TrimSpace(cfg.DeviceID) == "" {
		return errors.New("hive config missing device_id")
	}
	if strings.TrimSpace(sec.AccessClientID) == "" || strings.TrimSpace(sec.AccessClientSecret) == "" {
		return errors.New("hive config missing Access service credentials")
	}
	if strings.TrimSpace(sec.HMACKey) == "" {
		return errors.New("hive config missing hmac_key")
	}
	if _, err := NormalizeEnvironment(cfg.Environment); err != nil {
		return err
	}
	return nil
}

func NormalizeEnvironment(value string) (string, error) {
	raw := strings.ToLower(strings.TrimSpace(value))
	if raw == "" {
		return DefaultEnvironment, nil
	}
	switch raw {
	case "production", "test":
		return raw, nil
	default:
		return "", fmt.Errorf("invalid hive environment %q", value)
	}
}

func (c Client) Enroll(enrollmentToken string, options ...EnrollOptions) (EnrollResponse, error) {
	if strings.TrimSpace(enrollmentToken) == "" {
		return EnrollResponse{}, errors.New("enrollment token is required")
	}
	var opts EnrollOptions
	if len(options) > 0 {
		opts = options[0]
	}
	body := map[string]string{}
	if deviceID := strings.TrimSpace(opts.DeviceID); deviceID != "" {
		body["device_id"] = deviceID
	}
	environment, err := NormalizeEnvironment(opts.Environment)
	if err != nil {
		return EnrollResponse{}, err
	}
	body["environment"] = environment
	bodyData, err := json.Marshal(body)
	if err != nil {
		return EnrollResponse{}, err
	}
	req, err := c.newRequest(http.MethodPost, "/v1/enroll", bytes.NewReader(bodyData))
	if err != nil {
		return EnrollResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hive-Enroll-Token", enrollmentToken)
	var out EnrollResponse
	if err := c.doJSON(req, &out); err != nil {
		return EnrollResponse{}, err
	}
	if out.IngestPath == "" {
		out.IngestPath = "/v1/ingest"
	}
	if out.Environment == "" {
		out.Environment = environment
	}
	if _, err := NormalizeEnvironment(out.Environment); err != nil {
		return EnrollResponse{}, err
	}
	if out.DeviceID == "" || out.HMACKey == "" {
		return EnrollResponse{}, errors.New("hive enrollment response missing device_id or hmac_key")
	}
	return out, nil
}

func (c Client) FetchCurrentCatalog() (CatalogBundle, error) {
	req, err := c.newRequest(http.MethodGet, "/v1/catalog/current", nil)
	if err != nil {
		return CatalogBundle{}, err
	}
	req.Header.Set("X-Inventory-Device-Id", c.DeviceID)
	var out CatalogBundle
	if err := c.doJSON(req, &out); err != nil {
		return CatalogBundle{}, err
	}
	if err := validateBundle(out); err != nil {
		return CatalogBundle{}, err
	}
	return out, nil
}

func (c Client) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid hive base_url: %w", err)
	}
	ref, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, base.ResolveReference(ref).String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("CF-Access-Client-Id", c.AccessClientID)
	req.Header.Set("CF-Access-Client-Secret", c.AccessClientSecret)
	return req, nil
}

func (c Client) doJSON(req *http.Request, out any) error {
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("hive request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var body struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(io.LimitReader(resp.Body, 1024)).Decode(&body)
		if body.Error == "" {
			body.Error = resp.Status
		}
		return fmt.Errorf("hive returned %d: %s", resp.StatusCode, body.Error)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode hive response: %w", err)
	}
	return nil
}

func SyncCatalog(cacheDir string, bundle CatalogBundle) (CatalogManifest, error) {
	if cacheDir == "" {
		cacheDir = DefaultCacheDir()
	}
	if err := validateBundle(bundle); err != nil {
		return CatalogManifest{}, err
	}
	tmp, err := os.MkdirTemp(cacheDir, "sync-*")
	if err != nil {
		if mkErr := os.MkdirAll(cacheDir, 0o700); mkErr != nil {
			return CatalogManifest{}, fmt.Errorf("create catalog cache: %w", mkErr)
		}
		tmp, err = os.MkdirTemp(cacheDir, "sync-*")
		if err != nil {
			return CatalogManifest{}, fmt.Errorf("create catalog cache: %w", err)
		}
	}
	defer os.RemoveAll(tmp)

	for _, file := range bundle.Files {
		if err := os.WriteFile(filepath.Join(tmp, file.Path), []byte(file.Content), 0o600); err != nil {
			return CatalogManifest{}, fmt.Errorf("write catalog cache: %w", err)
		}
	}
	if _, err := exposure.Load(tmp, 64*1024*1024); err != nil {
		return CatalogManifest{}, fmt.Errorf("validate synced catalog: %w", err)
	}
	manifestData, err := json.MarshalIndent(bundle.Manifest, "", "  ")
	if err != nil {
		return CatalogManifest{}, err
	}
	if err := os.WriteFile(filepath.Join(tmp, manifestName), append(manifestData, '\n'), 0o600); err != nil {
		return CatalogManifest{}, fmt.Errorf("write catalog manifest: %w", err)
	}

	current := CurrentCatalogDir(cacheDir)
	old := filepath.Join(cacheDir, "previous")
	_ = os.RemoveAll(old)
	if _, err := os.Stat(current); err == nil {
		if err := os.Rename(current, old); err != nil {
			return CatalogManifest{}, fmt.Errorf("replace catalog cache: %w", err)
		}
	}
	if err := os.Rename(tmp, current); err != nil {
		_ = os.Rename(old, current)
		return CatalogManifest{}, fmt.Errorf("promote catalog cache: %w", err)
	}
	_ = os.RemoveAll(old)
	return bundle.Manifest, nil
}

func LoadCachedCatalog(cacheDir string) (CatalogManifest, error) {
	if cacheDir == "" {
		cacheDir = DefaultCacheDir()
	}
	current := CurrentCatalogDir(cacheDir)
	if _, err := exposure.Load(current, 64*1024*1024); err != nil {
		return CatalogManifest{}, fmt.Errorf("validate cached catalog: %w", err)
	}
	var manifest CatalogManifest
	if err := readJSON(filepath.Join(current, manifestName), &manifest); err != nil {
		return CatalogManifest{}, err
	}
	return manifest, nil
}

func CurrentCatalogDir(cacheDir string) string {
	if cacheDir == "" {
		cacheDir = DefaultCacheDir()
	}
	return filepath.Join(cacheDir, currentDirName)
}

func validateBundle(bundle CatalogBundle) error {
	if bundle.Manifest.ReleaseID == "" || bundle.Manifest.BundleSHA256 == "" {
		return errors.New("catalog manifest missing release_id or bundle_sha256")
	}
	if bundle.Manifest.Schema != "0.1.0" {
		return fmt.Errorf("unsupported catalog schema_version %q", bundle.Manifest.Schema)
	}
	if len(bundle.Files) == 0 {
		return errors.New("catalog bundle contains no files")
	}
	byPath := map[string]CatalogManifestFile{}
	for _, file := range bundle.Manifest.Files {
		byPath[file.Path] = file
	}
	for _, file := range bundle.Files {
		manifestFile, ok := byPath[file.Path]
		if !ok {
			return fmt.Errorf("catalog file %q missing from manifest", file.Path)
		}
		if !safeFileName(file.Path) {
			return fmt.Errorf("unsafe catalog file path %q", file.Path)
		}
		sum := sha256.Sum256([]byte(file.Content))
		got := hex.EncodeToString(sum[:])
		if got != file.SHA256 || got != manifestFile.SHA256 {
			return fmt.Errorf("catalog file %q sha256 mismatch", file.Path)
		}
	}
	return nil
}

func safeFileName(name string) bool {
	if name == "" || strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") || !strings.HasSuffix(name, ".json") {
		return false
	}
	return true
}
