package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCanonicalizer_Canonicalize(t *testing.T) {
	canonicalizer := NewCanonicalizer()

	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, "", canonicalizer.Canonicalize(""))
	})

	t.Run("Simple cell names", func(t *testing.T) {
		assert.Equal(t, "a1", canonicalizer.Canonicalize("A1"))
		assert.Equal(t, "a1", canonicalizer.Canonicalize("a1"))
	})

	t.Run("Cell names with underscores", func(t *testing.T) {
		assert.Equal(t, "a_1", canonicalizer.Canonicalize("A_1"))
		assert.Equal(t, "a_1_2s", canonicalizer.Canonicalize("A_1_2s"))
	})

	t.Run("Cell names with operators and keywords", func(t *testing.T) {
		assert.Equal(t, "or", canonicalizer.Canonicalize("OR"))
		assert.Equal(t, "or", canonicalizer.Canonicalize("or"))
		assert.Equal(t, "formula", canonicalizer.Canonicalize("formula"))
		assert.Equal(t, "cheep_and_dail", canonicalizer.Canonicalize("cheep_and_dail"))
	})

	t.Run("Floats", func(t *testing.T) {
		assert.Equal(t, "123.456", canonicalizer.Canonicalize("123.456"))
		assert.Equal(t, "123e5", canonicalizer.Canonicalize("123E5"))
		assert.Equal(t, "123.1e3", canonicalizer.Canonicalize("123.1E3"))

		// escape dot in non-numeric cell names
		assert.Equal(t, "123_r$46$r_1awesome3", canonicalizer.Canonicalize("123.1AWESOME3"))
		assert.Equal(t, "123_r$46$r_w", canonicalizer.Canonicalize("123.w"))
		assert.Equal(t, "123_r$46$r_w56", canonicalizer.Canonicalize("123.w56"))

		assert.Equal(t, "123w_r$46$r_456", canonicalizer.Canonicalize("123w.456"))
		assert.Equal(t, "w123_r$46$r_456", canonicalizer.Canonicalize("W123.456"))
	})

	t.Run("Keep dot in float in formulas", func(t *testing.T) {
		assert.Equal(t, "123.456+123", canonicalizer.Canonicalize("123.456+123"))
		assert.Equal(t, "789+123.456", canonicalizer.Canonicalize("789+123.456"))
		assert.Equal(t, "789+123.456+234", canonicalizer.Canonicalize("789+123.456+234"))

		assert.Equal(t, "9090 + 123.456", canonicalizer.Canonicalize("9090 + 123.456"))
	})
}
