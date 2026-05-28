package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	hivecfg "github.com/perplexityai/bumblebee/internal/hive"
	"github.com/perplexityai/bumblebee/internal/model"
)

func runHive(args []string) int {
	if len(args) < 1 {
		hiveUsage(os.Stderr)
		return 2
	}
	switch args[0] {
	case "join":
		return runHiveJoin(args[1:])
	case "catalog":
		if len(args) >= 2 && args[1] == "sync" {
			return runHiveCatalogSync(args[2:])
		}
	case "run":
		return runHiveRun(args[1:])
	case "-h", "--help", "help":
		hiveUsage(os.Stdout)
		return 0
	}
	fmt.Fprintf(os.Stderr, "unknown hive command %q\n", strings.Join(args, " "))
	hiveUsage(os.Stderr)
	return 2
}

func hiveUsage(w io.Writer) {
	fmt.Fprintln(w, `bumblebee hive <command>

commands:
  hive join          enroll this endpoint with Hive and write local config
  hive catalog sync  fetch and validate the current Hive exposure catalog
  hive run           sync catalog, run scan with it, and upload to Hive`)
}

func runHiveJoin(args []string) int {
	fs := flag.NewFlagSet("hive join", flag.ExitOnError)
	var roots stringList
	baseURL := fs.String("base-url", "", "Hive base URL")
	configDir := fs.String("config-dir", hivecfg.DefaultConfigDir(), "Hive config directory")
	cacheDir := fs.String("cache-dir", hivecfg.DefaultCacheDir(), "Hive catalog cache directory")
	accessIDEnv := fs.String("access-client-id-env", "BUMBLEBEE_ACCESS_CLIENT_ID", "env var holding Cloudflare Access client id")
	accessSecretEnv := fs.String("access-client-secret-env", "BUMBLEBEE_ACCESS_CLIENT_SECRET", "env var holding Cloudflare Access client secret")
	enrollTokenEnv := fs.String("enrollment-token-env", "BUMBLEBEE_ENROLLMENT_TOKEN", "env var holding Hive enrollment token")
	environment := fs.String("environment", hivecfg.DefaultEnvironment, "Hive device environment for new enrollment: production or test")
	newDevice := fs.Bool("new-device", false, "enroll a new Hive device even when local Hive config already exists")
	profile := fs.String("scan-profile", model.ProfileBaseline, "default scan profile for hive run")
	fs.Var(&roots, "root", "default scan root for hive run (repeatable)")
	_ = fs.Parse(args)
	seen := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		seen[f.Name] = true
	})

	existingCfg, existingSec, loadErr := hivecfg.LoadConfig(*configDir)
	existingOK := loadErr == nil
	if !existingOK && !*newDevice && hiveConfigFileExists(*configDir) {
		fmt.Fprintf(os.Stderr, "existing hive config is not reusable; use --new-device to replace it: %v\n", loadErr)
		return 2
	}

	base := strings.TrimRight(strings.TrimSpace(*baseURL), "/")
	if base == "" && existingOK {
		base = strings.TrimRight(existingCfg.BaseURL, "/")
	}
	if existingOK && seen["base-url"] && base != strings.TrimRight(existingCfg.BaseURL, "/") && !*newDevice {
		fmt.Fprintln(os.Stderr, "existing hive config is for a different base URL; use --new-device to enroll with a different Hive")
		return 2
	}
	if base == "" {
		fmt.Fprintln(os.Stderr, "--base-url is required when no reusable Hive config exists")
		return 2
	}

	accessID, accessIDSet := envValue(*accessIDEnv)
	if !accessIDSet && existingOK {
		accessID = existingSec.AccessClientID
	}
	if accessID == "" {
		fmt.Fprintf(os.Stderr, "env var %q is empty\n", *accessIDEnv)
		return 2
	}
	accessSecret, accessSecretSet := envValue(*accessSecretEnv)
	if !accessSecretSet && existingOK {
		accessSecret = existingSec.AccessClientSecret
	}
	if accessSecret == "" {
		fmt.Fprintf(os.Stderr, "env var %q is empty\n", *accessSecretEnv)
		return 2
	}

	targetEnvironment := hivecfg.DefaultEnvironment
	if existingOK {
		var err error
		targetEnvironment, err = hivecfg.NormalizeEnvironment(existingCfg.Environment)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return 2
		}
	}
	if seen["environment"] || !existingOK {
		var err error
		targetEnvironment, err = hivecfg.NormalizeEnvironment(*environment)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return 2
		}
	}
	if existingOK && !*newDevice {
		existingEnvironment, _ := hivecfg.NormalizeEnvironment(existingCfg.Environment)
		if seen["environment"] && targetEnvironment != existingEnvironment {
			fmt.Fprintln(os.Stderr, "existing hive config has a different environment; use --new-device to enroll a device in another environment")
			return 2
		}
		cfg := existingCfg
		cfg.BaseURL = base
		cfg.Environment = existingEnvironment
		if seen["scan-profile"] {
			cfg.ScanProfile = *profile
		}
		if len(roots) > 0 {
			cfg.ScanRoots = append([]string(nil), roots...)
		}
		sec := existingSec
		sec.AccessClientID = accessID
		sec.AccessClientSecret = accessSecret
		if err := hivecfg.SaveConfig(*configDir, cfg, sec); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return 1
		}
		if err := os.MkdirAll(*cacheDir, 0o700); err != nil {
			fmt.Fprintf(os.Stderr, "create hive cache: %v\n", err)
			return 1
		}
		writeHiveResult(map[string]any{"ok": true, "joined": true, "reused": true, "environment": cfg.Environment})
		return 0
	}

	enrollToken, ok := envValue(*enrollTokenEnv)
	if !ok {
		fmt.Fprintf(os.Stderr, "env var %q is empty\n", *enrollTokenEnv)
		return 2
	}
	client := hivecfg.Client{
		BaseURL:            base,
		AccessClientID:     accessID,
		AccessClientSecret: accessSecret,
	}
	enrollment, err := client.Enroll(enrollToken, hivecfg.EnrollOptions{Environment: targetEnvironment})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	targetEnvironment, err = hivecfg.NormalizeEnvironment(enrollment.Environment)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	scanProfile := *profile
	if existingOK && !seen["scan-profile"] {
		scanProfile = existingCfg.ScanProfile
	}
	scanRoots := append([]string(nil), roots...)
	if existingOK && len(roots) == 0 {
		scanRoots = append([]string(nil), existingCfg.ScanRoots...)
	}
	cfg := hivecfg.Config{
		BaseURL:     base,
		IngestPath:  enrollment.IngestPath,
		DeviceID:    enrollment.DeviceID,
		Environment: targetEnvironment,
		ScanProfile: scanProfile,
		ScanRoots:   scanRoots,
	}
	sec := hivecfg.Secrets{
		AccessClientID:     accessID,
		AccessClientSecret: accessSecret,
		HMACKey:            enrollment.HMACKey,
	}
	if err := hivecfg.SaveConfig(*configDir, cfg, sec); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	if err := os.MkdirAll(*cacheDir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "create hive cache: %v\n", err)
		return 1
	}
	writeHiveResult(map[string]any{"ok": true, "joined": true, "reused": false, "environment": cfg.Environment})
	return 0
}

func runHiveCatalogSync(args []string) int {
	fs := flag.NewFlagSet("hive catalog sync", flag.ExitOnError)
	configDir := fs.String("config-dir", hivecfg.DefaultConfigDir(), "Hive config directory")
	cacheDir := fs.String("cache-dir", hivecfg.DefaultCacheDir(), "Hive catalog cache directory")
	_ = fs.Parse(args)

	cfg, sec, err := hivecfg.LoadConfig(*configDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	manifest, err := syncConfiguredCatalog(cfg, sec, *cacheDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	cfg = applyCatalogManifest(cfg, manifest)
	if err := hivecfg.SaveConfig(*configDir, cfg, sec); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	writeHiveResult(map[string]any{
		"ok":          true,
		"release_id":  manifest.ReleaseID,
		"entry_count": manifest.EntryCount,
	})
	return 0
}

func runHiveRun(args []string) int {
	fs := flag.NewFlagSet("hive run", flag.ExitOnError)
	var roots stringList
	var ecosystems stringList
	configDir := fs.String("config-dir", hivecfg.DefaultConfigDir(), "Hive config directory")
	cacheDir := fs.String("cache-dir", hivecfg.DefaultCacheDir(), "Hive catalog cache directory")
	profile := fs.String("profile", "", "scan profile override")
	maxDuration := fs.Duration("max-duration", 0, "max wall-clock duration for the scan")
	findingsOnly := fs.Bool("findings-only", false, "suppress package records while emitting findings and summary")
	fs.Var(&roots, "root", "scan root override (repeatable)")
	fs.Var(&ecosystems, "ecosystem", "limit scanning to ecosystem values (repeatable or comma-separated)")
	_ = fs.Parse(args)

	cfg, sec, err := hivecfg.LoadConfig(*configDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	manifest, syncErr := syncConfiguredCatalog(cfg, sec, *cacheDir)
	if syncErr == nil {
		cfg = applyCatalogManifest(cfg, manifest)
		_ = hivecfg.SaveConfig(*configDir, cfg, sec)
	} else {
		fmt.Fprintln(os.Stderr, "hive catalog sync failed; trying last-known-good cache")
		manifest, err = hivecfg.LoadCachedCatalog(*cacheDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "hive catalog sync failed and no valid cache is available: %v\n", err)
			return 1
		}
	}

	scanProfile := cfg.ScanProfile
	if *profile != "" {
		scanProfile = *profile
	}
	if scanProfile == "" {
		scanProfile = model.ProfileBaseline
	}
	scanRoots := append([]string(nil), cfg.ScanRoots...)
	if len(roots) > 0 {
		scanRoots = append([]string(nil), roots...)
	}

	restore := setHiveEnv(sec, cfg.DeviceID)
	defer restore()
	scanArgs := []string{
		"--profile", scanProfile,
		"--exposure-catalog", hivecfg.CurrentCatalogDir(*cacheDir),
		"--output", "http",
		"--http-url", hiveIngestURL(cfg),
		"--http-auth", "hmac-sha256",
		"--http-hmac-key-env", "BUMBLEBEE_HIVE_HMAC_KEY",
		"--http-gzip",
		"--http-header-env", "CF-Access-Client-Id=BUMBLEBEE_HIVE_ACCESS_CLIENT_ID",
		"--http-header-env", "CF-Access-Client-Secret=BUMBLEBEE_HIVE_ACCESS_CLIENT_SECRET",
		"--http-header-env", "X-Inventory-Device-Id=BUMBLEBEE_HIVE_DEVICE_ID",
		"--device-id-env", "BUMBLEBEE_HIVE_DEVICE_ID",
	}
	if *findingsOnly {
		scanArgs = append(scanArgs, "--findings-only")
	}
	if *maxDuration > 0 {
		scanArgs = append(scanArgs, "--max-duration", maxDuration.String())
	}
	for _, ecosystem := range ecosystems {
		scanArgs = append(scanArgs, "--ecosystem", ecosystem)
	}
	for _, root := range scanRoots {
		scanArgs = append(scanArgs, "--root", root)
	}
	return runScanWithCatalogMetadata(scanArgs, &model.CatalogMetadata{
		ReleaseID:    manifest.ReleaseID,
		BundleSHA256: manifest.BundleSHA256,
		Source:       manifest.Source,
		PublishedAt:  manifest.PublishedAt,
		SyncedAt:     time.Now().UTC().Format(time.RFC3339Nano),
		EntryCount:   manifest.EntryCount,
	})
}

func syncConfiguredCatalog(cfg hivecfg.Config, sec hivecfg.Secrets, cacheDir string) (hivecfg.CatalogManifest, error) {
	client := hivecfg.Client{
		BaseURL:            cfg.BaseURL,
		AccessClientID:     sec.AccessClientID,
		AccessClientSecret: sec.AccessClientSecret,
		DeviceID:           cfg.DeviceID,
	}
	bundle, err := client.FetchCurrentCatalog()
	if err != nil {
		return hivecfg.CatalogManifest{}, err
	}
	return hivecfg.SyncCatalog(cacheDir, bundle)
}

func applyCatalogManifest(cfg hivecfg.Config, manifest hivecfg.CatalogManifest) hivecfg.Config {
	cfg.CatalogReleaseID = manifest.ReleaseID
	cfg.CatalogBundleSHA = manifest.BundleSHA256
	cfg.CatalogSource = manifest.Source
	cfg.CatalogPublished = manifest.PublishedAt
	cfg.CatalogEntryCount = manifest.EntryCount
	cfg.CatalogSyncedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return cfg
}

func hiveIngestURL(cfg hivecfg.Config) string {
	path := cfg.IngestPath
	if path == "" {
		path = "/v1/ingest"
	}
	return strings.TrimRight(cfg.BaseURL, "/") + path
}

func envValue(name string) (string, bool) {
	v := strings.TrimSpace(os.Getenv(name))
	return v, v != ""
}

func hiveConfigFileExists(dir string) bool {
	if dir == "" {
		dir = hivecfg.DefaultConfigDir()
	}
	_, err := os.Stat(filepath.Join(dir, hivecfg.ConfigFileName))
	return err == nil
}

func setHiveEnv(sec hivecfg.Secrets, deviceID string) func() {
	values := map[string]string{
		"BUMBLEBEE_HIVE_HMAC_KEY":             sec.HMACKey,
		"BUMBLEBEE_HIVE_ACCESS_CLIENT_ID":     sec.AccessClientID,
		"BUMBLEBEE_HIVE_ACCESS_CLIENT_SECRET": sec.AccessClientSecret,
		"BUMBLEBEE_HIVE_DEVICE_ID":            deviceID,
	}
	old := map[string]*string{}
	for k, v := range values {
		if existing, ok := os.LookupEnv(k); ok {
			copy := existing
			old[k] = &copy
		} else {
			old[k] = nil
		}
		_ = os.Setenv(k, v)
	}
	return func() {
		for k, v := range old {
			if v == nil {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, *v)
			}
		}
	}
}

func writeHiveResult(v any) {
	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(v)
}
