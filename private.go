package main

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"golang.org/x/mod/module"
)

type GoPrivate struct {
	goprivate string
}

var ErrNoVerionsInCmd = errors.New("no versions in cmd output")

func NewGoPrivate(goprivate string) *GoPrivate {
	if goprivate == "" {
		goprivate = os.Getenv("GOPRIVATE")
	}

	return &GoPrivate{
		goprivate: goprivate,
	}
}

func (p *GoPrivate) IsPrivate(modPath string) bool {
	return module.MatchPrefixPatterns(p.goprivate, modPath)
}

// FetchVersions fetches the version list for a module via go list -m -versions.
// Used for GOPRIVATE modules that should be accessed via the GOPROXY.
func (p *GoPrivate) FetchVersions(modPath, version string) ([]module.Version, error) {
	envs := []string{"GOPRIVATE=" + p.goprivate}
	args := []string{"list", "-m", "-versions", modPath}

	data, err := cmdOutput(config.GoBinary, envs, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}

	parts := strings.Fields(data)
	if len(parts) < 2 {
		return nil, ErrNoVerionsInCmd
	}

	versions := make([]module.Version, 0, len(parts)-1)
	for _, candidate := range parts[1:] {
		if !isValidCandidate(version, candidate) {
			continue
		}

		versions = append(versions, module.Version{
			Path:    modPath,
			Version: candidate,
		})
	}

	module.Sort(versions)
	slices.Reverse(versions)

	return versions, nil
}
