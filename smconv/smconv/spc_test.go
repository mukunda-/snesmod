// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package smconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A test to help ensure that the SPC patch location is correct. If this test fails then
// the patch location needs to be updated. See `verifySpcPatchSignature` for more info.
func TestConfirmPatchSignature(t *testing.T) {
	// When exporting an SPC, there is a small patch made to start playback during the boot
	// process. The signature is checked to ensure that the code will be correctly
	// patched. The program will panic if there's a signature mismatch.
	assert.True(t, verifySpcPatchSignature())
}
