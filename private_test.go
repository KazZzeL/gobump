package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGoPrivate(t *testing.T) {
	t.Run("empty goprivate falls back to env", func(t *testing.T) {
		goprivate := "private.corp.com"

		t.Setenv("GOPRIVATE", goprivate)
		p := NewGoPrivate("")

		require.Equalf(t, goprivate, p.goprivate, "expected %s as goprivate, got %q", goprivate, p.goprivate)
	})

	t.Run("explicit goprivate overrides env", func(t *testing.T) {
		goprivate := "private.corp.com"

		t.Setenv("GOPRIVATE", "should.not.use")
		p := NewGoPrivate(goprivate)

		require.Equalf(t, goprivate, p.goprivate, "expected %s as goprivate, got %q", goprivate, p.goprivate)
	})
}
