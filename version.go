package main

import (
	"strings"

	"golang.org/x/mod/semver"
)

func isPreRelease(version string) bool {
	return strings.Contains(version, "-")
}

func isValidCandidate(oldV, newV string) bool {
	// skip pre-release versions
	if !isPreRelease(oldV) && isPreRelease(newV) {
		return false
	}

	// skip versions higher than MaxBump
	switch config.MaxBump {
	case "minor":
		if semver.Major(oldV) != semver.Major(newV) {
			return false
		}
	case "patch":
		if semver.MajorMinor(oldV) != semver.MajorMinor(newV) {
			return false
		}
	}

	// skip lower versions
	return semver.Compare(oldV, newV) < 0
}
