// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package smconv

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

var ErrModuleSizeExceeded = errors.New("module size exceeded limit")

func binwrite(w io.Writer, data any) error {
	return binary.Write(w, binary.LittleEndian, data)
}

func wtell(w io.WriteSeeker) (int64, error) {
	return w.Seek(0, io.SeekCurrent)
}

/*
func safecall(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case error:
				err = r.(error)
			default:
				err = errors.New("unknown error")
			}
		}
	}()

	return f()
}*/

/*
func (bank *SoundBank) export(filenameBase string) {

}*/

func (bank *SoundBank) Export(filenameBase string, hirom bool) (result error) {

	bankFilename := filenameBase + ".smbank"
	file, err := os.Create(bankFilename)
	if err != nil {
		return err
	}

	if err := binwrite(file, uint16(len(bank.Sources))); err != nil {
		return err
	}

	if err := binwrite(file, uint16(len(bank.Modules))); err != nil {
		return err
	}

	// reserve space for tables
	for i := 0; i < 128+len(bank.Sources); i++ {
		file.Write([]byte{0xAA, 0xAA, 0xAA})
	}

	modulePointers := []uint32{}
	sourcePointers := []uint32{}

	for _, module := range bank.Modules {
		if position, err := wtell(file); err != nil {
			return err
		} else {
			modulePointers = append(modulePointers, uint32(position))
		}

		if err := module.Export(file, true); err != nil {
			return err
		}
	}

	for _, source := range bank.Sources {
		if position, err := wtell(file); err != nil {
			return err
		} else {
			sourcePointers = append(sourcePointers, uint32(position))
		}
		if err := source.Export(file, false); err != nil {
			return err
		}
	}

	if _, err := file.Seek(4, io.SeekStart); err != nil {
		return err
	}

	// export module pointers

	if _, err := file.Seek(4, io.SeekStart); err != nil {
		return err
	}

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

		if err := binwrite(file, addr); err != nil {
			return err
		}

		if err := binwrite(file, addrBank); err != nil {
			return err
		}
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
		if err := binwrite(file, addr); err != nil {
			return err
		}

		if err := binwrite(file, addrBank); err != nil {
			return err
		}
	}

	return nil
}

func (mod *SmModule) Export(w io.WriteSeeker, writeHeader bool) error {

	headerStart, err := w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	mod.BankHeader.ModuleSize = 0xAAAA
	mod.BankHeader.SourceListCount = uint16(len(mod.SourceList))

	if writeHeader {
		// Reserve for module size
		if err := binwrite(w, mod.BankHeader); err != nil {
			return err
		}

		if err := binwrite(w, mod.SourceList); err != nil {
			return err
		}
	}

	moduleStart, err := w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	pointers := SmModuleHeaderPointers{}

	if err := binwrite(w, mod.Header); err != nil {
		return err
	}

	patternPointers := []uint16{}
	instrumentPointers := []uint16{}
	samplePointers := []uint16{}

	startOfPointers, err := w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	// This is just to reserve space for now. We'll fill it in after.
	if err := binwrite(w, pointers); err != nil {
		return err
	}

	for i := 0; i < len(mod.Patterns); i++ {
		ptr, err := w.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		ptr -= moduleStart

		if ptr > kSpcRamSize {
			return ErrModuleSizeExceeded
		}

		patternPointers = append(patternPointers, uint16(ptr+kModuleBase))
		if err := mod.Patterns[i].Export(w); err != nil {
			return err
		}
	}

	for i := 0; i < len(mod.Instruments); i++ {
		ptr, err := w.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		ptr -= moduleStart

		if ptr > kSpcRamSize {
			return ErrModuleSizeExceeded
		}

		instrumentPointers = append(instrumentPointers, uint16(ptr+kModuleBase))
		if err := mod.Instruments[i].Export(w); err != nil {
			return err
		}
	}

	for i := 0; i < len(mod.Samples); i++ {
		ptr, err := w.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		ptr -= moduleStart

		if ptr > kSpcRamSize {
			return ErrModuleSizeExceeded
		}

		samplePointers = append(samplePointers, uint16(ptr+kModuleBase))
		if err := mod.Samples[i].Export(w); err != nil {
			return err
		}
	}

	moduleEnd, err := w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	// Align end to 2 bytes.
	if moduleEnd&1 != 0 {
		_, err := w.Write([]byte{0x00})
		if err != nil {
			return err
		}
		moduleEnd++
	}

	if writeHeader {
		_, err := w.Seek(headerStart, io.SeekStart)
		if err != nil {
			return err
		}

		// +1 for rouding up last word (not needed with the align above.
		modSize := uint16((moduleEnd - moduleStart + 1) >> 1)
		if err := binwrite(w, modSize); err != nil {
			return err
		}
	}

	_, err = w.Seek(startOfPointers, io.SeekStart)
	if err != nil {
		return err
	}

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

	if err := binwrite(w, pointers); err != nil {
		return err
	}

	_, err = w.Seek(moduleEnd, io.SeekStart)
	if err != nil {
		return err
	}

	return nil
}

func (smp *SmPattern) Export(w io.WriteSeeker) error {
	if err := binwrite(w, smp.Rows); err != nil {
		return err
	}

	if err := binwrite(w, smp.Data); err != nil {
		return err
	}

	return nil
}

func (smi *SmInstrument) Export(w io.WriteSeeker) error {
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

	if err := binwrite(w, info1); err != nil {
		return err
	}

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

		if err := binwrite(w, info2); err != nil {
			return err
		}

		if err := binwrite(w, smi.Envelope); err != nil {
			return err
		}
	}

	return nil
}

func (sms *SmSample) Export(w io.WriteSeeker) error {

	if err := binwrite(w, sms); err != nil {
		return err
	}

	return nil
}

// Export to a file. `spcDirect` indicates that the file is directly in SPC memory, so
// certain headers are not saved (since they would be stripped during the SPC loading
// process). spcDirect is DATA only.
func (source *Source) Export(w io.WriteSeeker, spcDirect bool) error {
	if !spcDirect {
		if err := binwrite(w, uint16(len(source.Data))); err != nil {
			return err
		}

		if err := binwrite(w, uint16(source.Loop)); err != nil {
			return err
		}
	}

	if err := binwrite(w, source.Data); err != nil {
		return err
	}

	if !spcDirect {
		if len(source.Data)&1 != 0 {
			if err := binwrite(w, uint8(0)); err != nil {
				return err
			}
		}
	}

	return nil
}

/*



file.Close();

std::string asm_out = output;
asm_out += ".asm";
ExportASM( bin_output.c_str(), asm_out.c_str() );

std::string inc_out = output;
inc_out += ".inc";
ExportINC( inc_out.c_str() );
}

void Bank::ExportINC( const char *output ) const {
FILE *f = fopen( output, "w" );

fprintf( f,
	"; snesmod soundbank definitions\n\n"
	".ifndef __SOUNDBANK_DEFINITIONS__\n"
	".define __SOUNDBANK_DEFINITIONS__\n\n"
	".import __SOUNDBANK__\n\n");

for i := 0; i < Modules.size(); i++ {
	if( !Modules[i]->id.empty() ) {
		fprintf( f, "%-32s = %i\n", Modules[i]->id.c_str(), i );
	}
}
fprintf( f, "\n" );

for i := 0; i < Sources.size(); i++ {
	if( !Sources[i]->id.empty() ) {
		fprintf( f, "%-32s = %i\n", Sources[i]->id.c_str(), i );
	}
}
fprintf( f, "\n.endif ; __SOUNDBANK_DEFINITIONS__\n" );
fclose(f);
}

void Bank::ExportASM( const char *inputfile, const char *outputfile ) const {
FILE *f = fopen( outputfile, "w" );

int size = IO::FileSize( inputfile );

fprintf( f,
	";************************************************\n"
	"; snesmod soundbank data                        *\n"
	"; total size: %10i bytes                  *\n"
	";************************************************\n"
	"\n"
	"\t.global __SOUNDBANK__\n"
	"\t.segment \"SOUNDBANK\" ; need dedicated bank(s)\n\n"
	"__SOUNDBANK__:\n",
	size
);

std::string foo = inputfile;

for i := 0; i < foo.size(); i++ {
	if( foo[i] == '\\' ) foo[i] = '/';
}
int ffo = foo.find_last_of( '/' );
if( ffo != std::string::npos )
	foo = foo.substr( ffo + 1 );

fprintf( f, "\t.incbin \"%s\"\n", foo.c_str() );

}

void Source::Export( IO::File &file, bool spc_direct ) const {

if( !spc_direct ) {
	file.Write16( Length );
	file.Write16( Loop );
}

for( int i = 0; i < Length; i++ {
	file.Write8( Data[i] );
}

if( !spc_direct ) {
	if( Length & 1 )
		file.Write8( 0 ); // align by 2
}
}









*/
