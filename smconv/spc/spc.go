// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

/*
For working with SNES SPC files.
*/
package spc

import (
	"encoding/binary"
	"io"

	cat "go.mukunda.com/errorcat"
)

func bread(r io.Reader, data any) {
	err := binary.Read(r, binary.LittleEndian, data)
	cat.Catch(err)
}

func bwrite(w io.Writer, data any) {
	err := binary.Write(w, binary.LittleEndian, data)
	cat.Catch(err)
}

type SpcFile struct {
	Header       SpcHeader
	Memory       [65536]byte
	DspRegisters [128]byte
	Reserved     [64]byte
	IplRom       [64]byte

	// Extended Id666 data, copied directly.
	Extended []byte
}

// 256 byte header.
type SpcHeader struct {
	Signature    [33]byte
	FixedBytes   [2]byte
	HasTags      byte
	VersionMinor byte
	PC           uint16
	A            uint8
	X            uint8
	Y            uint8
	PSW          uint8
	SP           uint8
	_            [2]byte // Reserved

	// This data is only valid if in TEXT mode and HasTags = 27
	Tags Id6Tags
}

type FromEmulatorType = uint8

const (
	FromEmulatorUnknown  FromEmulatorType = 0
	FromEmulatorZNES     FromEmulatorType = 1
	FromEmulatorSNES9x   FromEmulatorType = 2
	FromEmulatorZST2SPC  FromEmulatorType = 3
	FromEmulatorOther    FromEmulatorType = 4
	FromEmulatorSNEShout FromEmulatorType = 5
	FromEmulatorZSNESW   FromEmulatorType = 6
	FromEmulatorSNES9xpp FromEmulatorType = 7
	FromEmulatorSNESGT   FromEmulatorType = 8
	FromEmulatorSnesmod  FromEmulatorType = 15
)

// There are two known types of ID666 structures in an SPC file, TEXT and BINARY.
// https://dgrfactory-jp.translate.goog/spcplay/id666.html?_x_tr_sl=ja&_x_tr_tl=en&_x_tr_hl=en&_x_tr_pto=sc
//
// It seems to me like some people have been putting whatever the hell they felt like in
// this space, e.g., I've seen other data in the fields that doesn't align with either
// format. I'm going to do my part for the future's sanity and pretend like the BINARY
// version doesn't exist. Let's stick with TEXT and let the other die.
//
// -Maybe- we could support reading the BINARY format.
type Id6Tags struct {
	SongTitle              [32]byte
	GameTitle              [32]byte
	DumpedBy               [16]byte
	Comments               [32]byte
	DateDumped             [11]byte
	SongDuration           [3]byte // Seconds before fading out
	SongFadeOut            [5]byte // Fade out time (ms)
	Composer               [32]byte
	InitialChannelDisabled byte
	FromEmulator           FromEmulatorType
	_                      [45]byte
}

func (spc *SpcFile) Read(r io.Reader) (rerr error) {
	defer cat.Guard(&rerr)

	bread(r, &spc.Header)
	bread(r, &spc.Memory)
	bread(r, &spc.DspRegisters)
	bread(r, &spc.Reserved)
	bread(r, &spc.IplRom)

	var err error
	if spc.Extended, err = io.ReadAll(r); err != nil {
		return err
	}

	return nil
}

func (spc *SpcFile) Write(w io.Writer) (rerr error) {
	defer cat.Guard(&rerr)

	bwrite(w, &spc.Header)
	bwrite(w, &spc.Memory)
	bwrite(w, &spc.DspRegisters)
	bwrite(w, &spc.Reserved)
	bwrite(w, &spc.IplRom)
	bwrite(w, spc.Extended)

	return nil
}

func NewSpcFile() *SpcFile {
	spcf := &SpcFile{}
	spcf.SetHeaderDefaults()
	return spcf
}

func (spc *SpcFile) SetHeaderDefaults() {
	copy(spc.Header.Signature[:], "SNES-SPC700 Sound File Data v0.30")
	spc.Header.FixedBytes = [2]byte{26, 26}
	spc.Header.HasTags = 26
	spc.Header.VersionMinor = 30
}
