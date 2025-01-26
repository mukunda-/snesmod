package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func TestForSmoke(t *testing.T) {
	smconvCli([]string{"smconv", "--help"})
}

func TestToSpc(t *testing.T) {
	smconvCli([]string{"-o", ".testdata-pollen.spc", "test/pollen8.it"})

	assert.True(t, fileExists(".testdata-pollen.spc"))

	// TODO:  verify things?
}

func TestToSoundbank(t *testing.T) {
	smconvCli([]string{"-s", "-o", ".testdata-pollen", "test/pollen8.it"})

	assert.True(t, fileExists(".testdata-pollen.smbank"))
	assert.True(t, fileExists(".testdata-pollen.asm"))
	assert.True(t, fileExists(".testdata-pollen.inc"))
	// TODO:  verify things
}
