package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/mod/modfile"
)

func parseModFile(t *testing.T, content string) *modfile.File {
	t.Helper()
	mf, err := modfile.Parse("go.mod", []byte(content), nil)
	if err != nil {
		t.Fatalf("failed to parse modfile: %v", err)
	}
	return mf
}

func TestCheckModFile(t *testing.T) {
	t.Parallel()

	t.Run("same Go version ok", func(t *testing.T) {
		t.Parallel()

		modFile := parseModFile(t, `go 1.22`)

		_, err := CheckModFile(modFile, "1.22", "v1.0.0", nil)

		require.NoErrorf(t, err, "expected no error, got: %v", err)
	})

	t.Run("candidate lower Go version ok", func(t *testing.T) {
		t.Parallel()

		modFile := parseModFile(t, `go 1.21`)

		_, err := CheckModFile(modFile, "1.22", "v1.0.0", nil)

		require.NoErrorf(t, err, "expected no error, got: %v", err)
	})

	t.Run("candidate requires higher Go version preserves retractions", func(t *testing.T) {
		t.Parallel()

		knownRet := []*modfile.Retract{
			{VersionInterval: modfile.VersionInterval{Low: "v1.0.0", High: "v1.0.0"}},
		}
		modFile := parseModFile(t, `go 1.23`)

		rets, err := CheckModFile(modFile, "1.22", "v1.0.0", knownRet)

		require.ErrorIs(t, err, ErrGoVersionTooHigh, "expected error for higher Go version")
		require.Lenf(t, rets, 1, "expected retractions preserved, got %d", len(rets))
	})

	t.Run("candidate requires higher Go with 3-part version", func(t *testing.T) {
		t.Parallel()

		modFile := parseModFile(t, `go 1.23.0`)

		_, err := CheckModFile(modFile, "1.22.0", "v1.0.0", nil)

		require.ErrorIs(t, err, ErrGoVersionTooHigh, "expected error for higher Go version")
	})

	t.Run("known retraction skips version but preserves retractions", func(t *testing.T) {
		t.Parallel()

		knownRet := []*modfile.Retract{
			{VersionInterval: modfile.VersionInterval{Low: "v0.5.0", High: "v0.5.0"}},
		}
		modFile := parseModFile(t, `go 1.22`)

		rets, err := CheckModFile(modFile, "1.22", "v0.5.0", knownRet)

		require.ErrorIs(t, err, ErrVersionIsRetracted, "expected error for retracted version")
		require.Lenf(t, rets, 1, "expected retractions preserved, got %d", len(rets))
	})

	t.Run("known retract range skips version", func(t *testing.T) {
		t.Parallel()

		knownRet := []*modfile.Retract{
			{VersionInterval: modfile.VersionInterval{Low: "v1.0.0", High: "v1.5.0"}},
		}
		modFile := parseModFile(t, `go 1.22`)

		_, err := CheckModFile(modFile, "1.22", "v1.2.3", knownRet)

		require.ErrorIs(t, err, ErrVersionIsRetracted, "expected error for retracted version")
	})

	t.Run("non-retracted version passes", func(t *testing.T) {
		t.Parallel()

		knownRet := []*modfile.Retract{
			{VersionInterval: modfile.VersionInterval{Low: "v2.0.0", High: "v2.0.0"}},
		}
		modFile := parseModFile(t, `go 1.22`)

		rets, err := CheckModFile(modFile, "1.22", "v1.0.0", knownRet)

		require.NoErrorf(t, err, "expected no error, got: %v", err)
		require.NotEmpty(t, rets, "expected non-empty retractions")
	})

	t.Run("accumulates retractions from modFile", func(t *testing.T) {
		t.Parallel()

		knownRet := []*modfile.Retract{
			{VersionInterval: modfile.VersionInterval{Low: "v0.5.0", High: "v0.5.0"}},
		}
		modFile := parseModFile(t, `go 1.22
			retract v1.0.0
		`)

		rets, err := CheckModFile(modFile, "1.22", "v2.0.0", knownRet)

		require.NoErrorf(t, err, "expected no error, got: %v", err)
		require.Lenf(t, rets, 2, "expected 2 retractions, got %d", len(rets))
	})

	t.Run("retraction from modFile triggers skip but preserves retractions", func(t *testing.T) {
		t.Parallel()

		modFile := parseModFile(t, `go 1.22
			retract v1.0.0
		`)

		rets, err := CheckModFile(modFile, "1.22", "v1.0.0", nil)

		require.ErrorIs(t, err, ErrVersionIsRetracted, "expected error for version retracted by modFile")
		require.Lenf(t, rets, 1, "expected 1 retraction from modFile, got %d", len(rets))
	})

	t.Run("Go version comparison uses toSemver normalization", func(t *testing.T) {
		t.Parallel()

		modFile := parseModFile(t, `go 1.23`)

		_, err := CheckModFile(modFile, "1.22.0", "v1.0.0", nil)

		require.ErrorIs(t, err, ErrGoVersionTooHigh, "expected error for higher Go version")
	})

	t.Run("nil Go in modfile does not panic", func(t *testing.T) {
		t.Parallel()

		mf := &modfile.File{}

		_, err := CheckModFile(mf, "1.22", "v1.0.0", nil)

		require.NoErrorf(t, err, "expected no error, got: %v", err)
	})
}
