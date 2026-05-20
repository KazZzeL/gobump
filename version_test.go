package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPreRelease(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version string
		want    bool
	}{
		{"v1.0.0", false},
		{"v2.0.0", false},
		{"v0.67.5", false},
		{"v1.0.0-alpha", true},
		{"v1.0.0-beta.1", true},
		{"v1.0.0-rc1", true},
		{"v1.5.1-0.20250403130103-3d3abc24416a", true},
		{"v0.51.0-alpha.0", true},
		{"v1.1.0-alpha1", true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, isPreRelease(tt.version))
		})
	}
}

func TestIsValidCandidate(t *testing.T) {
	t.Run("default major level", func(t *testing.T) {
		config = &AppConfig{MaxBump: "major"}

		require.True(t, isValidCandidate("v1.0.0", "v2.0.0"), "higher major")
		require.True(t, isValidCandidate("v1.0.0", "v1.1.0"), "higher minor")
		require.True(t, isValidCandidate("v1.0.0", "v1.0.1"), "higher patch")
		require.False(t, isValidCandidate("v2.0.0", "v1.0.0"), "lower version")
		require.False(t, isValidCandidate("v1.0.0", "v1.0.0"), "same version")
	})

	t.Run("minor level", func(t *testing.T) {
		config = &AppConfig{MaxBump: "minor"}

		require.True(t, isValidCandidate("v1.0.0", "v1.1.0"), "same major, higher minor")
		require.True(t, isValidCandidate("v1.0.0", "v1.0.1"), "same major, higher patch")
		require.False(t, isValidCandidate("v1.0.0", "v2.0.0"), "different major")
		require.False(t, isValidCandidate("v2.0.0", "v1.0.0"), "lower version, different major")
		require.False(t, isValidCandidate("v1.1.0", "v1.0.0"), "same major, lower minor")
		require.False(t, isValidCandidate("v1.0.0", "v1.0.0"), "same version")
	})

	t.Run("patch level", func(t *testing.T) {
		config = &AppConfig{MaxBump: "patch"}

		require.True(t, isValidCandidate("v1.0.0", "v1.0.1"), "same major.minor, higher patch")
		require.False(t, isValidCandidate("v1.0.0", "v1.1.0"), "same major, different minor")
		require.False(t, isValidCandidate("v1.0.0", "v2.0.0"), "different major")
		require.False(t, isValidCandidate("v1.0.1", "v1.0.0"), "same major.minor, lower patch")
		require.False(t, isValidCandidate("v1.0.0", "v1.0.0"), "same version")
	})

	t.Run("pre-release filtering", func(t *testing.T) {
		config = &AppConfig{MaxBump: "major"}

		require.False(t, isValidCandidate("v1.0.0", "v2.0.0-beta"), "current stable, candidate pre-release")
		require.True(t, isValidCandidate("v1.0.0-beta", "v2.0.0"), "current pre-release, candidate stable")
		require.True(t, isValidCandidate("v1.0.0-beta", "v2.0.0-alpha"), "both pre-release")
	})

	t.Run("pre-release and minor level", func(t *testing.T) {
		config = &AppConfig{MaxBump: "minor"}

		require.False(t, isValidCandidate("v1.0.0", "v1.1.0-alpha"), "current stable, candidate pre-release")
		require.True(t, isValidCandidate("v1.0.0-alpha", "v1.1.0"), "current pre-release, candidate stable, same major")
	})
}
