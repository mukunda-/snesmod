// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package smconv

import (
	"errors"
	"os"
	"time"

	"go.mukunda.com/snesmod/smconv/spc"

	_ "embed"
)

//go:embed sm_spc.bin
var spcDriverBinary []byte

const kSpcPatchStart = 0x3C

var ErrModuleTooBig = errors.New("total module data is too big to fit in SPC memory")

// Returns true if the SPC driver patch signature could be verified.
func verifySpcPatchSignature() bool {
	// When updating the SPC driver, the patch location must be verified.
	// We'll do a signature check against the patch region. This is executed in tests and
	// at runtime.

	// 0457   043C             spc_patch_start:
	// 0458   043C 2F 08       	bra	spc_patch_end
	// 0460   043E 3F CF 06    	call	Module_Stop
	// 0461   0441 E8 00       	mov	a, #0
	// 0462   0443 3F D9 06    	call	Module_Start
	// 0463   0446             spc_patch_end:

	// 0x55 = wildcard
	signature := []byte{
		0x2f, 0x08,
		0x3f, 0x55, 0x55,
		0xe8, 0x00,
		0x3f, 0x55, 0x55,
	}

	// To update the signature, refer to build/sm_spc.lst after assembling the driver.

	for i := 0; i < len(signature); i++ {
		if signature[i] == 0x55 {
			continue //wildcard (linked address)
		}
		if signature[i] != spcDriverBinary[kSpcPatchStart+i] {
			return false
		}
	}

	return true
}

// Export the soundbank to an SPC file. The soundbank must have only one module in it.
func (bank *SoundBank) WriteSpcFile(filename string) error {

	if !verifySpcPatchSignature() {
		panic("SPC driver signature mismatch. Please update to use the SPC export function.")
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	mod := bank.Modules[0]
	spcf := spc.NewSpcFile()
	spcf.Header.PC = 0x400
	spcf.Header.SP = 0xEF
	copy(spcf.Header.Tags.SongTitle[:], mod.Title)
	copy(spcf.Header.Tags.GameTitle[:], "SNESMOD")
	copy(spcf.Header.Tags.DumpedBy[:], "SNESMOD")

	copy(spcf.Header.Tags.Comments[:], mod.SongMessage)

	datestring := time.Now().Format("01/02/2006")

	copy(spcf.Header.Tags.DateDumped[:], datestring)
	copy(spcf.Header.Tags.SongDuration[:], "180")
	copy(spcf.Header.Tags.SongFadeOut[:], "5000")
	copy(spcf.Header.Tags.Composer[:], "") /// Todo: load artist information from mod
	spcf.Header.Tags.FromEmulator = spc.FromEmulatorSnesmod

	// Memory offsets for SPC driver
	const memSpcProgram = 0x400
	const memSampleTable = 0x200
	const memModuleStart = 0x1a00

	if len(spcDriverBinary) > memModuleStart-memSpcProgram {
		return errors.New("spc driver is too large")
	}

	copy(spcf.Memory[memSpcProgram:], spcDriverBinary)

	// See "spc_patch_start" in driver source. This allows the module to start playing
	// immediately. This address is verified in spc_test.go.
	spcf.Memory[memSpcProgram+kSpcPatchStart] = 0
	spcf.Memory[memSpcProgram+kSpcPatchStart+1] = 0

	moduleBuffer := &SeekingByteBuffer{}
	err = bank.Modules[0].Export(moduleBuffer, false)
	if err != nil {
		return err
	}

	for i := 0; i < len(bank.Modules[0].SourceList); i++ {
		source := bank.Sources[bank.Modules[0].SourceList[i]]

		// Copy the sample START and LOOP points to memory.
		sampleStart := uint16(moduleBuffer.Tell()) + memModuleStart
		sampleLoop := sampleStart + uint16(source.Loop)

		spcf.Memory[memSampleTable+i*4] = byte(sampleStart & 0xFF)
		spcf.Memory[memSampleTable+i*4+1] = byte((sampleStart >> 8))
		spcf.Memory[memSampleTable+i*4+2] = byte(sampleLoop & 0xFF)
		spcf.Memory[memSampleTable+i*4+3] = byte((sampleLoop >> 8))

		// Load the BRR data into memory.
		moduleBuffer.Write(source.Data)
	}

	if memModuleStart+len(moduleBuffer.Bytes()) > 0xFF80 {
		return ErrModuleTooBig
	}

	copy(spcf.Memory[memModuleStart:], moduleBuffer.Bytes())

	err = spcf.Write(file)
	if err != nil {
		return err
	}

	return nil
}
