package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

// moduleVersionOrigin is the VCS origin recorded by the module proxy for a release.
type moduleVersionOrigin struct {
	VCS  string `json:"VCS"`
	URL  string `json:"URL"`
	Hash string `json:"Hash"`
	Ref  string `json:"Ref"`
}

// moduleVersionInfo is the JSON body of a module proxy @v/VERSION.info response.
type moduleVersionInfo struct {
	Version string              `json:"Version"`
	Time    string              `json:"Time"`
	Origin  moduleVersionOrigin `json:"Origin"`
}

type GoProxy struct {
	baseURL string
	client  *http.Client
}

var ErrWrongStatusCode = errors.New("failed to check status code")

// ModuleProxyBaseURL resolves the module proxy base URL: a non-empty
// explicit base URL, else the first usable entry in $GOPROXY, else the public proxy.
func ModuleProxyBaseURL(explicit string) string {
	explicit = strings.TrimSpace(explicit)
	if explicit != "" {
		return strings.TrimSuffix(explicit, "/")
	}
	gp := strings.TrimSpace(os.Getenv("GOPROXY"))
	if gp == "" {
		return "https://proxy.golang.org"
	}
	for _, p := range strings.Split(gp, ",") {
		p = strings.TrimSpace(p)
		if p == "" || p == "direct" || p == "off" {
			continue
		}
		if strings.HasPrefix(p, "file://") {
			continue
		}
		return strings.TrimSuffix(p, "/")
	}
	return "https://proxy.golang.org"
}

func NewGoProxy(configured string) *GoProxy {
	base := ModuleProxyBaseURL(configured)
	return &GoProxy{
		baseURL: base,
		client:  newHTTPClient(),
	}
}

func isPreRelease(version string) bool {
	return strings.Contains(version, "-")
}

// FetchVersions fetches the list of versions for a given module from the Go proxy.
// It returns a slice of module.Version structs sorted in descending order.
// Pre-release versions will return pre-release versions
func (p *GoProxy) FetchVersions(modName string, version string) ([]module.Version, error) {
	versions := []module.Version{}

	modName, err := module.EscapePath(modName)
	if err != nil {
		return nil, fmt.Errorf("failed to escape module path: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		fmt.Sprintf("%s/%s/@v/list", p.baseURL, modName), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	setDefaultHTTPHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch versions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrWrongStatusCode, resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// skip pre-release versions
		if !isPreRelease(version) && isPreRelease(line) {
			continue
		}

		// skip lower versions
		if semver.Compare(version, line) >= 0 {
			continue
		}

		v := module.Version{
			Path:    modName,
			Version: strings.TrimSpace(line),
		}

		versions = append(versions, v)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	module.Sort(versions)
	slices.Reverse(versions)

	return versions, nil
}

// FetchVersionInfo returns proxy metadata for a single module version.
func (p *GoProxy) FetchVersionInfo(modPath, version string) (moduleVersionInfo, error) {
	var info moduleVersionInfo
	escaped, err := module.EscapePath(modPath)
	if err != nil {
		return info, fmt.Errorf("failed to escape module path: %w", err)
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		fmt.Sprintf("%s/%s/@v/%s.info", p.baseURL, escaped, version), nil)
	if err != nil {
		return info, fmt.Errorf("failed to build request: %w", err)
	}
	setDefaultHTTPHeaders(req)
	resp, err := p.client.Do(req)
	if err != nil {
		return info, fmt.Errorf("failed to fetch version info: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return info, fmt.Errorf("%w: %s", ErrWrongStatusCode, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return info, fmt.Errorf("failed to decode version info: %w", err)
	}
	return info, nil
}

// discardBody drains and closes a response body when the caller does not need it.
func discardBody(resp *http.Response) {
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

// fetchModFile fetches the go.mod content for a specific module version from the Go proxy.
func (p *GoProxy) FetchModFile(modName, version string) (*modfile.File, error) {
	modName, err := module.EscapePath(modName)
	if err != nil {
		return nil, fmt.Errorf("failed to escape module path: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		fmt.Sprintf("%s/%s/@v/%s.mod", p.baseURL, modName, version), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	setDefaultHTTPHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mod file: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrWrongStatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read mod file: %w", err)
	}

	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mod file: %w", err)
	}

	return modFile, nil
}
