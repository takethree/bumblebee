// Package powershell scans PowerShell module manifest metadata.
//
// The parser is intentionally file-based and does not execute PowerShell,
// PowerShellGet, PSResourceGet, or any Gallery/API lookup. It reads only the
// top-level manifest fields needed for package inventory.
package powershell

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/perplexityai/bumblebee/internal/model"
)

const Ecosystem = model.EcosystemPowerShellModule

type Scanner struct {
	MaxFileSize int64
	Emit        func(model.Record)
	Diag        func(level, path, msg string)
}

func IsManifest(base string) bool {
	return strings.EqualFold(filepath.Ext(base), ".psd1")
}

func (s *Scanner) ScanManifest(path string, base model.Record) error {
	data, err := s.readBounded(path)
	if err != nil {
		return err
	}
	fields := parseManifestFields(string(data))
	moduleName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	version := strings.TrimSpace(fields["moduleversion"])
	if moduleName == "" || version == "" {
		return nil
	}
	r := base
	r.Ecosystem = Ecosystem
	r.PackageName = moduleName
	r.NormalizedName = strings.ToLower(moduleName)
	r.Version = version
	r.ProjectPath = filepath.Dir(path)
	r.PackageManager = "powershell"
	r.SourceType = "powershell-module-manifest"
	r.SourceFile = path
	r.Confidence = "high"
	s.Emit(r)
	return nil
}

func parseManifestFields(input string) map[string]string {
	fields := map[string]string{}
	i := strings.Index(input, "@{")
	if i < 0 {
		return fields
	}
	p := parser{s: input, pos: i + 2, depth: 1, fields: fields}
	p.parse()
	return fields
}

type parser struct {
	s      string
	pos    int
	depth  int
	fields map[string]string
}

func (p *parser) parse() {
	for p.pos < len(p.s) && p.depth > 0 {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.s) || p.depth <= 0 {
			return
		}
		if p.peek() == '}' {
			p.depth--
			p.pos++
			continue
		}
		if p.depth != 1 {
			p.skipValue()
			continue
		}
		key, ok := p.readKey()
		if !ok {
			p.pos++
			continue
		}
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.s) || p.peek() != '=' {
			continue
		}
		p.pos++
		p.skipWhitespaceAndComments()
		value, ok := p.readScalarValue()
		if ok {
			p.fields[strings.ToLower(key)] = value
		} else {
			p.skipValue()
		}
	}
}

func (p *parser) readKey() (string, bool) {
	if p.pos >= len(p.s) {
		return "", false
	}
	switch p.peek() {
	case '\'', '"':
		return p.readQuotedString()
	default:
		start := p.pos
		for p.pos < len(p.s) {
			r := rune(p.s[p.pos])
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
				p.pos++
				continue
			}
			break
		}
		if p.pos == start {
			return "", false
		}
		return p.s[start:p.pos], true
	}
}

func (p *parser) readScalarValue() (string, bool) {
	if p.pos >= len(p.s) {
		return "", false
	}
	switch p.peek() {
	case '\'', '"':
		return p.readQuotedString()
	case '@', '{', '(':
		return "", false
	default:
		start := p.pos
		for p.pos < len(p.s) {
			switch p.peek() {
			case '\r', '\n', ';', ',', '}':
				return strings.TrimSpace(p.s[start:p.pos]), p.pos > start
			case '#':
				return strings.TrimSpace(p.s[start:p.pos]), p.pos > start
			default:
				p.pos++
			}
		}
		return strings.TrimSpace(p.s[start:p.pos]), p.pos > start
	}
}

func (p *parser) readQuotedString() (string, bool) {
	if p.pos >= len(p.s) {
		return "", false
	}
	quote := p.peek()
	if quote != '\'' && quote != '"' {
		return "", false
	}
	p.pos++
	var b strings.Builder
	for p.pos < len(p.s) {
		ch := p.peek()
		p.pos++
		switch ch {
		case quote:
			if quote == '\'' && p.pos < len(p.s) && p.peek() == '\'' {
				b.WriteByte('\'')
				p.pos++
				continue
			}
			return b.String(), true
		case '`':
			if quote == '"' && p.pos < len(p.s) {
				b.WriteByte(p.peek())
				p.pos++
				continue
			}
			b.WriteByte(ch)
		default:
			b.WriteByte(ch)
		}
	}
	return "", false
}

func (p *parser) skipWhitespaceAndComments() {
	for p.pos < len(p.s) {
		ch := p.peek()
		switch ch {
		case ' ', '\t', '\r', '\n', ';', ',':
			p.pos++
		case '#':
			for p.pos < len(p.s) && p.peek() != '\n' {
				p.pos++
			}
		default:
			return
		}
	}
}

func (p *parser) skipValue() {
	for p.pos < len(p.s) {
		ch := p.peek()
		switch ch {
		case '\'', '"':
			_, _ = p.readQuotedString()
		case '@':
			if p.pos+1 < len(p.s) && p.s[p.pos+1] == '{' {
				p.depth++
				p.pos += 2
			} else {
				p.pos++
			}
		case '{':
			p.depth++
			p.pos++
		case '}':
			p.depth--
			p.pos++
			return
		case '\n', ';', ',':
			if p.depth <= 1 {
				p.pos++
				return
			}
			p.pos++
		default:
			p.pos++
		}
		if p.depth <= 0 {
			return
		}
	}
}

func (p *parser) peek() byte {
	return p.s[p.pos]
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
