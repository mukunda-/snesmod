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

//go:embed spc-driver.bin
var spcDriverBinary []byte

var ErrModuleTooBig = errors.New("total module data is too big to fit in SPC memory")

func (bank *SoundBank) WriteSpcFile(filename string) error {
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

	// See "patch for it->spc conversion" in driver source. This allows the module
	// to start playing immediately.
	spcf.Memory[memSpcProgram+0x3C] = 0 // PATCH
	spcf.Memory[memSpcProgram+0x3D] = 0

	moduleBuffer := &SeekingByteBuffer{}
	err = bank.Modules[0].Export(moduleBuffer, false)
	if err != nil {
		return err
	}

	//sourceTable := make([]uint16, len(bank.Modules[0].SourceList)*2)

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
