// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

// This file describes the functions to export data to a soundbank file.

package smconv

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var ErrModuleSizeExceeded = errors.New("module size exceeded limit")

func (bank *SoundBank) Export(filename string, hirom bool) (rerr error) {
	defer pguard(&rerr)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	bwrite(file, uint16(len(bank.Sources)))
	bwrite(file, uint16(len(bank.Modules)))

	// reserve space for tables
	for i := 0; i < 128+len(bank.Sources); i++ {
		bwrite(file, []byte{0xAA, 0xAA, 0xAA})
	}

	modulePointers := []uint32{}
	sourcePointers := []uint32{}

	for _, module := range bank.Modules {
		modulePointers = append(modulePointers, uint32(ptell(file)))
		pcatch(module.Export(file, true))
	}

	for _, source := range bank.Sources {
		sourcePointers = append(sourcePointers, uint32(ptell(file)))
		pcatch(source.Export(file, false))
	}

	// export module pointers
	pseek(file, 4, io.SeekStart)

	for i := 0; i < 128; i++ {
		addr := uint16(0)
		addrBank := uint8(0)
		if i < len(bank.Modules) {
			if hirom { // 64k banks
				addr = uint16(modulePointers[i] & 65535)
				addrBank = uint8(modulePointers[i] >> 16)
			} else { // 32k banks
				addr = uint16(0x8000 + (modulePointers[i] & 32767))
				addrBank = uint8(modulePointers[i] >> 15)
			}
		}

		bwrite(file, addr)
		bwrite(file, addrBank)
	}

	// export source pointers
	for i := 0; i < len(bank.Sources); i++ {
		addr := uint16(0)
		addrBank := uint8(0)
		if hirom { // 64k banks
			addr = uint16(sourcePointers[i] & 65535)
			addrBank = uint8(sourcePointers[i] >> 16)
		} else { // 32k banks
			addr = uint16(0x8000 + (sourcePointers[i] & 32767))
			addrBank = uint8(sourcePointers[i] >> 15)
		}

		bwrite(file, addr)
		bwrite(file, addrBank)
	}

	return nil
}

func (mod *SmModule) Export(w io.WriteSeeker, writeHeader bool) (rerr error) {
	defer pguard(&rerr)

	headerStart := ptell(w)

	mod.BankHeader.ModuleSize = 0xAAAA
	mod.BankHeader.SourceListCount = uint16(len(mod.SourceList))

	if writeHeader {
		// Reserve for module size
		bwrite(w, mod.BankHeader)
		bwrite(w, mod.SourceList)
	}

	moduleStart := ptell(w)

	pointers := SmModuleHeaderPointers{}

	bwrite(w, mod.Header)

	patternPointers := []uint16{}
	instrumentPointers := []uint16{}
	samplePointers := []uint16{}

	startOfPointers := ptell(w)

	// This is just to reserve space for now. We'll fill it in after.
	bwrite(w, pointers)

	for i := 0; i < len(mod.Patterns); i++ {
		ptr := ptell(w)
		ptr -= moduleStart

		if ptr > kSpcRamSize {
			return ErrModuleSizeExceeded
		}

		patternPointers = append(patternPointers, uint16(ptr+kModuleBase))
		pcatch(mod.Patterns[i].Export(w))
	}

	for i := 0; i < len(mod.Instruments); i++ {
		ptr := ptell(w)
		ptr -= moduleStart

		if ptr > kSpcRamSize {
			return ErrModuleSizeExceeded
		}

		instrumentPointers = append(instrumentPointers, uint16(ptr+kModuleBase))
		pcatch(mod.Instruments[i].Export(w))
	}

	for i := 0; i < len(mod.Samples); i++ {
		ptr := ptell(w)
		ptr -= moduleStart

		if ptr > kSpcRamSize {
			return ErrModuleSizeExceeded
		}

		samplePointers = append(samplePointers, uint16(ptr+kModuleBase))
		pcatch(mod.Samples[i].Export(w))
	}

	moduleEnd := ptell(w)

	// Align end to 2 bytes.
	if moduleEnd&1 != 0 {
		bwrite(w, uint8(0))
		moduleEnd++
	}

	if writeHeader {
		pseek(w, headerStart, io.SeekStart)

		// +1 for rouding up last word (not needed with the align above.
		modSize := uint16((moduleEnd - moduleStart + 1) >> 1)
		bwrite(w, modSize)
	}

	pseek(w, startOfPointers, io.SeekStart)

	for i := 0; i < 64; i++ {
		if i < len(mod.Patterns) {
			pointers.PatternsL[i] = byte(patternPointers[i] & 0xFF)
			pointers.PatternsH[i] = byte(patternPointers[i] >> 8)
		} else {
			pointers.PatternsL[i] = 0xFF
			pointers.PatternsH[i] = 0xFF
		}

		if i < len(mod.Instruments) {
			pointers.InstrumentsL[i] = byte(instrumentPointers[i] & 0xFF)
			pointers.InstrumentsH[i] = byte(instrumentPointers[i] >> 8)
		} else {
			pointers.InstrumentsL[i] = 0xFF
			pointers.InstrumentsH[i] = 0xFF
		}

		if i < len(mod.Samples) {
			pointers.SamplesL[i] = byte(samplePointers[i] & 0xFF)
			pointers.SamplesH[i] = byte(samplePointers[i] >> 8)
		} else {
			pointers.SamplesL[i] = 0xFF
			pointers.SamplesH[i] = 0xFF
		}
	}

	bwrite(w, pointers)

	pseek(w, moduleEnd, io.SeekStart)

	return nil
}

func (smp *SmPattern) Export(w io.WriteSeeker) (rerr error) {
	defer pguard(&rerr)

	bwrite(w, smp.Rows)
	bwrite(w, smp.Data)

	return nil
}

func (smi *SmInstrument) Export(w io.WriteSeeker) (rerr error) {
	defer pguard(&rerr)

	info1 := struct {
		Fadeout        uint8
		SampleIndex    uint8
		GlobalVolume   uint8
		SetPanning     uint8
		EnvelopeLength uint8
	}{}

	info1.Fadeout = smi.Info.Fadeout
	info1.SampleIndex = smi.Info.SampleIndex
	info1.GlobalVolume = smi.Info.GlobalVolume
	info1.SetPanning = smi.Info.SetPanning
	info1.EnvelopeLength = smi.Info.EnvelopeLength

	bwrite(w, info1)

	if smi.Info.EnvelopeLength > 0 {
		info2 := struct {
			EnvelopeSustain   uint8
			EnvelopeLoopStart uint8
			EnvelopeLoopEnd   uint8
		}{
			EnvelopeSustain:   smi.Info.EnvelopeSustain,
			EnvelopeLoopStart: smi.Info.EnvelopeLoopStart,
			EnvelopeLoopEnd:   smi.Info.EnvelopeLoopEnd,
		}

		bwrite(w, info2)
		bwrite(w, smi.Envelope)
	}

	return nil
}

func (sms *SmSample) Export(w io.WriteSeeker) (rerr error) {
	defer pguard(&rerr)

	bwrite(w, sms)

	return nil
}

// Export to a file. `dataOnly` writes the BRR data only. This is used when building an
// SPC file. When the source is loaded into SPC memory, there is no header or alignment.
func (source *Source) Export(w io.WriteSeeker, dataOnly bool) (rerr error) {
	defer pguard(&rerr)

	if !dataOnly {
		bwrite(w, uint16(len(source.Data)))
		bwrite(w, uint16(source.Loop))
	}

	bwrite(w, source.Data)

	if !dataOnly {
		if len(source.Data)&1 != 0 {
			bwrite(w, uint8(0))
		}
	}

	return nil
}

// Write ca65 assembly source file that includes the soundbank binary.
func (bank *SoundBank) ExportAssembly(filename string, binfile string) (rerr error) {
	pguard(&rerr)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	bwrite(file, []byte(`; SNESMOD Soundbank Data
; Generated by SMCONV

	.global __SOUNDBANK__
	.segment "SOUNDBANK" ; This needs dedicated bank(s)
__SOUNDBANK__:
`))

	binfile = strings.ReplaceAll(binfile, "\\", "/")
	lastSlash := strings.LastIndex(binfile, "/")
	if lastSlash != -1 {
		binfile = binfile[lastSlash+1:]
	}

	bwrite(file, []byte(fmt.Sprintf("\t.incbin \"%s\"\n", binfile)))

	return nil
}

// Write ca65 assembly include file that contains the soundbank definitions.
func (bank *SoundBank) ExportAssemblyInclude(filename string) (rerr error) {
	pguard(&rerr)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	bwrite(f, []byte(`; SNESMOD Soundbank Definitions
; GENERATED BY SMCONV

.ifndef __SOUNDBANK_DEFINITIONS__
.define __SOUNDBANK_DEFINITIONS__

.import __SOUNDBANK__

`))

	for index, mod := range bank.Modules {
		if mod.Id != "" {
			bwrite(f, []byte(fmt.Sprintf("%-32s = %d\n", mod.Id, index)))
		}
	}

	bwrite(f, []byte("\n"))

	for index, source := range bank.Sources {
		if source.Id != "" {
			bwrite(f, []byte(fmt.Sprintf("%-32s = %d\n", source.Id, index)))
		}
	}
	bwrite(f, []byte("\n.endif ; __SOUNDBANK_DEFINITIONS__\n"))

	return nil
}
