package main

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
)

var (
	ErrGoVersionTooHigh   = errors.New("version requires higher Go version")
	ErrVersionIsRetracted = errors.New("version is retracted")
)

// checkModFile fetches the go.mod for the given version via proxy, checks its Go version and
// if the version has been retracted.
func CheckModFile(
	modFile *modfile.File,
	okGoVer,
	canVersion string,
	knownRetractions []*modfile.Retract,
) ([]*modfile.Retract, error) {
	if modFile.Go != nil && semver.Compare(toSemver(modFile.Go.Version), toSemver(okGoVer)) > 0 {
		return knownRetractions, fmt.Errorf("%w: %s -> %s", ErrGoVersionTooHigh, canVersion, modFile.Go.Version)
	}

	if len(modFile.Retract) > 0 {
		knownRetractions = append(knownRetractions, modFile.Retract...)
	}

	for _, ret := range knownRetractions {
		if semver.Compare(canVersion, ret.Low) >= 0 && semver.Compare(canVersion, ret.High) <= 0 {
			return knownRetractions, fmt.Errorf("%w: %s", ErrVersionIsRetracted, canVersion)
		}
	}

	return knownRetractions, nil
}

func toSemver(v string) string {
	if strings.Count(v, ".") == 1 {
		v += ".0"
	}
	return "v" + v
}
