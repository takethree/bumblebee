// Package nuget scans NuGet project metadata.
//
// The parser is intentionally file-based and does not execute nuget, dotnet,
// PowerShell, Visual Studio, or any package-manager command. v1 support reads
// project/deep metadata only: packages.config for legacy projects and
// packages.lock.json for PackageReference restore locks.
package nuget

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/perplexityai/bumblebee/internal/model"
)

const Ecosystem = model.EcosystemNuGet

type Scanner struct {
	MaxFileSize int64
	Emit        func(model.Record)
	Diag        func(level, path, msg string)
}

func IsPackagesConfig(base string) bool { return strings.EqualFold(base, "packages.config") }
func IsLockfile(base string) bool       { return strings.EqualFold(base, "packages.lock.json") }

type packagesConfig struct {
	Packages []packagesConfigPackage `xml:"package"`
}

type packagesConfigPackage struct {
	ID      string `xml:"id,attr"`
	Version string `xml:"version,attr"`
}

func (s *Scanner) ScanPackagesConfig(path string, base model.Record) error {
	data, err := s.readBounded(path)
	if err != nil {
		return err
	}
	var cfg packagesConfig
	if err := xml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	projectPath := filepath.Dir(path)
	for _, p := range cfg.Packages {
		s.emitPackage(path, projectPath, "nuget-packages-config", p.ID, p.Version, "", "", base)
	}
	return nil
}

type lockfile struct {
	Dependencies map[string]map[string]lockDependency `json:"dependencies"`
}

type lockDependency struct {
	Type      string `json:"type"`
	Requested string `json:"requested"`
	Resolved  string `json:"resolved"`
}

func (s *Scanner) ScanLockfile(path string, base model.Record) error {
	data, err := s.readBounded(path)
	if err != nil {
		return err
	}
	var lf lockfile
	if err := json.Unmarshal(data, &lf); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	projectPath := filepath.Dir(path)
	seen := make(map[string]struct{})
	for _, deps := range lf.Dependencies {
		for name, dep := range deps {
			if dep.Resolved == "" || strings.EqualFold(dep.Type, "project") {
				continue
			}
			scope := strings.ToLower(dep.Type)
			key := strings.ToLower(name) + "\x00" + dep.Resolved + "\x00" + scope
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			s.emitPackage(path, projectPath, "nuget-lockfile", name, dep.Resolved, dep.Requested, scope, base)
		}
	}
	return nil
}

func (s *Scanner) emitPackage(source, projectPath, sourceType, name, version, requested, scope string, base model.Record) {
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)
	if name == "" || version == "" {
		return
	}
	requested = strings.TrimSpace(requested)
	r := base
	r.Ecosystem = Ecosystem
	r.PackageName = name
	r.NormalizedName = strings.ToLower(name)
	r.Version = version
	r.RequestedSpec = requested
	r.ProjectPath = projectPath
	r.PackageManager = "nuget"
	r.SourceType = sourceType
	r.SourceFile = source
	r.Confidence = "high"
	switch scope {
	case "direct":
		direct := true
		r.DirectDependency = &direct
	case "transitive":
		direct := false
		r.DirectDependency = &direct
		r.InstallScope = "transitive"
	case "":
	default:
		r.InstallScope = scope
	}
	s.Emit(r)
}

func (s *Scanner) readBounded(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("not a regular file")
	}
	if s.MaxFileSize > 0 && info.Size() > s.MaxFileSize {
		if s.Diag != nil {
			s.Diag("warn", path, fmt.Sprintf("skipping: size %d exceeds max %d", info.Size(), s.MaxFileSize))
		}
		return nil, fmt.Errorf("file %s exceeds max size %d", path, s.MaxFileSize)
	}
	return io.ReadAll(f)
}
