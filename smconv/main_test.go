package main

import "testing"

func TestForSmoke(t *testing.T) {
	smconvCli([]string{"smconv", "--help"})
}

func TestToSpc(t *testing.T) {
	smconvCli([]string{"-o", ".testdata-pollen2.spc", "test/pollen8.it"})
	// TODO:  verify things
}

func TestToSoundbank(t *testing.T) {
	smconvCli([]string{"-s", "-o", ".testdata-pollen", "test/pollen8.it"})
	// TODO:  verify things
}
