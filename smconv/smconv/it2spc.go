// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package smconv

import (
	"regexp"
	"strings"

	"go.mukunda.com/modlib/common"
)

const (
	// Free SPC memory to be used by module data.
	kSpcRamSize = 58000

	// Base of module in SPC memory.
	kModuleBase = 0x1A00
)

type SourceIndex = uint16

type SoundBank struct {
	HiRom   bool
	Sources []*Source
	Modules []*SmModule
}

// TODO ADD MODULE CODE
/*
Bank::Bank( const ITLoader::Bank &bank, bool HIROM ) {

	HiROM = HIROM;
	for i := 0, n = bank.modules.size(); i < n; i++ {
		AddModule( *(bank.modules[i]) );
	}

	for i := 0, n = bank.sounds.size(); i < n; i++ {
		AddSource( *(bank.sounds[i]) );
	}
}*/

func (bank *SoundBank) AddModule(mod *common.Module, filename string) error {

	// usedSources indexes into the bank.Sources.
	// The sampleSourceMap indexes into usedSources, mapping samples -> used sources.

	// In other words, usedSources dictates what to load, and sampleSourceMap references
	// loaded sources.

	usedSources := []SourceIndex{}
	sampleSourceMap := []uint8{}

	for i := 0; i < len(mod.Samples); i++ {
		s, err := createSource(mod.Samples[i])
		if err != nil {
			return err
		}
		index := bank.AddSource(s)
		exists := false

		for j := 0; j < len(usedSources); j++ {
			if usedSources[j] == index {
				sampleSourceMap = append(sampleSourceMap, uint8(j))
				exists = true
				break
			}
		}

		if !exists {
			sampleSourceMap = append(sampleSourceMap, uint8(len(usedSources)))
			usedSources = append(usedSources, index)
		}
	}

	smMod := convertModule(mod, filename, usedSources, sampleSourceMap, bank.Sources)
	bank.Modules = append(bank.Modules, smMod)
	return nil
}

// Adds a source and returns the index of it. If a duplicate source exists, then the
// existing index is returned instead and nothing is added.
func (bank *SoundBank) AddSource(s *Source) SourceIndex {
	for i, es := range bank.Sources {
		if es.Hash == s.Hash {
			return SourceIndex(i)
		}
	}

	bank.Sources = append(bank.Sources, s)
	return SourceIndex(len(bank.Sources) - 1)
}

// Convert a file path into an ID for inserting into definition files.
// This strips the directory and extension, converts to uppercase, and replaces non
// alphanumeric with "_".
func pathToId(prefix string, path string) string {
	path = strings.ToUpper(path)
	path = strings.ReplaceAll(path, "\\", "/")

	slash := strings.LastIndex(path, "/")
	if slash != -1 {
		path = path[slash+1:]
	}

	dot := strings.Index(path, ".")
	if dot != -1 {
		path = path[:dot]
	}

	path = prefix + path

	stripper := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	path = stripper.ReplaceAllString(path, "_")

	return path
}

/*
static int SearchForSNESMODtag( const char *message ) {

	const char *test = "[[SNESMOD]]";
	int matches = 0;

	for i := 0; message[i]; i++ {
		if( message[i] == test[matches] )
			matches++;
		else
			matches = 0;
		if( matches == 11 )
			return i - 10;
	}
	return -1;
}

static int SeekNextLine( const char *text, int offset ) {
	while( text[offset] != 0 &&
		text[offset] != '\n' &&
		text[offset] != '\r' )

		offset++;

	while( text[offset] == '\n' || text[offset] == '\r' )
		offset++;

	return offset;
}

static bool is_whitespace( char c ) {
	return c == ' ' || c == '\t';
}

static bool not_term( char c ) {
	return c != 0 && c != '\r' && c != '\n';
}

static int minmax( int value, int min, int max ) {
	value = value < min ? min : value;
	return value > max ? max : value;
}

void Module::ParseSMOption( const char *text ) {
	std::vector<std::string> args;
	if( !text ) return;
	if( !text[0] ) return;

	int offs = 0;

	// skip whitespace
	while( is_whitespace(text[offs]) ) {
		offs++;
	}

	while( not_term(text[offs]) ) { //!= 0 && text[offs] != '\r' && text[offs] != '\n' ) {
		int len;
		for( len = 0; not_term(text[offs+len]); len++ {
			if( is_whitespace( text[offs+len] ) ) {
				break;
			}
		}
		if( len == 0 ) {
			offs++;
		} else {
			args.push_back( "" );
			args[args.size()-1].assign( text + offs, len );
			offs += len;
		}

		// skip whitespace
		while( is_whitespace(text[offs]) ) {
			offs++;
		}
	}

	if( args.empty() ) {
		// no args?
		return;
	}

	// args is filled with arguments
	// determine command type

	for( u32 i = 0; i < args[0].size(); i++ {
		if( args[0].at(i) >= 'A' && args[0].at(i) <= 'Z' ) {
			args[0].at(i) += 'a' - 'A';
		}
	}

#define TESTCMD(cmd) if( args[0] == cmd )

	TESTCMD( "edl" ) {
		if( args.size() < 2 ) {
			return;
		}
		EchoDelay = minmax( atoi( args[1].c_str() ), 0, 15 );
	} else TESTCMD( "efb" ) {
		if( args.size() < 2 ) {
			return;
		}
		EchoFeedback = minmax( atoi( args[1].c_str() ), -128, 127 );
	} else TESTCMD( "evol" ) {
		if( args.size() < 2 ) {
			return;
		} else if( args.size() < 3 ) {
			EchoVolumeL = minmax( atoi( args[1].c_str() ), -128, 127 );
			EchoVolumeR = EchoVolumeL;
		} else {
			EchoVolumeL = minmax( atoi( args[1].c_str() ), -128, 127 );
			EchoVolumeR = minmax( atoi( args[2].c_str() ), -128, 127 );
		}
	} else TESTCMD( "efir" ) {
		for( u32 i = 0; i < 8; i++ {
			if( args.size() <= (1+i) ) {
				return;
			}
			EchoFIR[i] = minmax( atoi( args[1+i].c_str() ), -128, 127 );
		}
	} else TESTCMD( "eon" ) {
		for( u32 i = 1; i < args.size(); i++ {
			int c = atoi( args[i].c_str() );
			if( c >= 1 && c <= 8 ) {
				EchoEnable |= (1<<(c-1));
			}
		}
	}
}

void Module::ParseSMOptions( const ITLoader::Module &mod ) {
	EchoVolumeL = 0;
	EchoVolumeR = 0;
	EchoDelay = 0;
	EchoFeedback = 0;
	EchoFIR[0] = 127;
	for i := 1; i < 8; i++ )
		EchoFIR[i] = 0;

	EchoEnable = 0;

	if( mod.Message ) {
		int offset = SearchForSNESMODtag( mod.Message );
		if( offset != -1 ) {

			offset = SeekNextLine( mod.Message, offset );
			while( mod.Message[offset] != 0 ) {
				ParseSMOption( mod.Message + offset );
				offset = SeekNextLine( mod.Message, offset );
			}


		}
	}
}

Module::~Module() {
	for( u32 i = 0; i < Patterns.size(); i++ )
		delete Patterns[i];

	for( u32 i = 0; i < Instruments.size(); i++ )
		delete Instruments[i];

	for( u32 i = 0; i < Samples.size(); i++ )
		delete Samples[i];
}
*/
/***********************************************************************************************
*
* Pattern
*
***********************************************************************************************/
/*
Pattern::Pattern( ITLoader::Pattern &source ) {
	Rows = (u8)(source.Rows - 1);

	int row;
	u8 *read = source.Data;

	if( source.DataLength != 0 ) {

		std::vector<u8> row_buffer;

		u8 spc_hints;
		u8 data_bits;
		u8 p_maskvar[8];

		spc_hints = 0;
		data_bits = 0;

		for( row = 0; row < source.Rows; ) {
			u8 chvar = *read++;

			if( chvar == 0 ) {
				Data.push_back( 0xFF ); //spc_hints
				Data.push_back( data_bits );

				for i := 0, n = row_buffer.size(); i < n; i++ )
					Data.push_back( row_buffer[i] );
				row_buffer.clear();

				data_bits = 0;
				spc_hints = 0;
				row++;
				continue;
			}

			int channel = (chvar - 1) & 63;
			data_bits |= 1<<channel;

			u8 maskvar;
			if( chvar & 128 ) {
				maskvar = *read++;
				maskvar = ((maskvar>>4) | (maskvar<<4)) & 0xFF;
				maskvar |= maskvar>>4;
				p_maskvar[channel] = maskvar;
			} else {
				maskvar = p_maskvar[channel];
			}

			row_buffer.push_back( maskvar );

			if( maskvar & 16 ) {		// note
				row_buffer.push_back( *read++ );
			}
			if( maskvar & 32 ) {		// instr
				row_buffer.push_back( *read++ );
			}
			if( maskvar & 64 ) {		// vcmd
				row_buffer.push_back( *read++ );
			}

			u8 cmd,param;
			if( maskvar & 128 ) {		// cmd+param

				row_buffer.push_back( cmd = *read++ );
				row_buffer.push_back( param = *read++ );
			}

			if( (maskvar & 1) ) {
				spc_hints |= 1<<channel;
				if( maskvar & 128 ) {
					if( (cmd == 7) || (cmd == 19 && (param&0xF0) == 0xD0) ) {
						// glissando or note delay:
						// cancel hint
						spc_hints &= ~(1<<channel);
					}
				}
			}
		}
	} else {
		// empty pattern
		for i := 0; i < source.Rows; i++ {
			Data.push_back( 0 );
			Data.push_back( 0 );
		}
	}
}

Pattern::~Pattern() {

}
*/
/***********************************************************************************************
*
* Instrument
*
***********************************************************************************************/
/*
Instrument::Instrument( const ITLoader::Instrument &source ) {
	int a =  source.Fadeout / 4;
	Fadeout = a > 255 ? 255 : a;
	SampleIndex = source.Notemap[60].Sample - 1;
	GlobalVolume = source.GlobalVolume;
	SetPan = source.DefaultPan;

	// load envelope
	const ITLoader::Envelope &e = *source.VolumeEnvelope;

	EnvelopeLength = e.Enabled ? e.Length*4 : 0;

	if( EnvelopeLength ) {
		EnvelopeSustain = e.Sustain ? e.SustainStart*4 : 0xFF;
		EnvelopeLoopStart = e.Loop ? e.LoopStart*4 : 0xFF;
		EnvelopeLoopEnd = e.Loop ? e.LoopEnd*4 : 0xFF;

		EnvelopeData = new EnvelopeNode[ EnvelopeLength/4 ];
		for i := 0; i < EnvelopeLength/4; i++ {
			EnvelopeNode &ed = EnvelopeData[i];
			ed.y = e.Nodes[i].y;
			if( i != (EnvelopeLength/4)-1 ) {
				int duration = e.Nodes[i+1].x - e.Nodes[i].x;
				if( duration == 0 ) duration = 1;
				ed.duration = duration;
				ed.delta = ((e.Nodes[i+1].y - e.Nodes[i].y) * 256 + duration/2) / duration;
			} else {
				ed.delta = 0;
				ed.duration = 0;
			}
		}
	} else {
		EnvelopeData = 0;
	}

}

Instrument::~Instrument() {
	if( EnvelopeData )
		delete EnvelopeData;
}
*/
/***********************************************************************************************
*
* Sample
*
***********************************************************************************************/
/*
Sample::Sample( const ITLoader::Sample &source, int directory, double tuning ) {
	DefaultVolume = source.DefaultVolume;
	GlobalVolume = source.GlobalVolume;
	SetPan = source.DefaultPanning ^ 128;

	double a = ((double)source.Data.C5Speed * tuning);
	PitchBase = (int)floor(log(a / 8363.0) / log(2.0) * 768.0 + 0.5);

	DirectoryIndex = directory;
}
*/
/***********************************************************************************************
*
* Source
*
***********************************************************************************************/
/*
bool Source::Compare( const Source &test ) const {
	if( Length != test.Length )
		return false;
	if( Loop != test.Loop )
		return false;
	for i := 0; i < Length; i++ {
		if( Data[i] != test.Data[i] ) {
			return false;
		}
	}
	return true;
}
*/
/***********************************************************************************************
*
* Export
*
***********************************************************************************************/
/*
void Bank::MakeSPC( const char *spcfile ) const {
	std::string nstr;

	IO::File file( spcfile, IO::MODE_WRITE );

//		for( u32 i = 0; i < Modules.size(); i++ {
//			nstr = output;
//			char tnumber[64];
//			sprintf( tnumber, "%i", i );
//			nstr += tnumber;
//			nstr += ".spc";
		// open file
	//	IO::File file( nstr.c_str(), IO::MODE_WRITE );

		file.WriteAscii( "SNES-SPC700 Sound File Data v0.30" );
		file.Write8( 26 ); // 26,26
		file.Write8( 26 );
		file.Write8( 26 );	// header contins id666 tag
		file.Write8( 30 ); // version minor

		// SPC700 registers
		file.Write16( 0x400 );	// PC
		file.Write8( 0 );		// A
		file.Write8( 0 );		// X
		file.Write8( 0 );		// Y
		file.Write8( 0 );		// PSW
		file.Write8( 0xEF );	// SP
		file.Write16( 0 );		// reserved

		// ID666 tag
		file.WriteAsciiF( "<INSERT SONG TITLE>", 32 );
		file.WriteAsciiF( "<INSERT GAME TITLE>", 32 );
		file.WriteAsciiF( "NAME OF DUMPER", 16 );
		file.WriteAsciiF( "comments...", 32 );
		file.WriteAsciiF( "", 11 );
		file.WriteAsciiF( "180", 3 );
		file.WriteAsciiF( "5000", 5 );
		file.WriteAsciiF( "<INSERT SONG ARTIST>", 32 );
		file.Write8( 0);
		file.Write8( '0' );
		file.ZeroFill( 45 ); // reserved

		//-------------------------------------------------------
		// SPC memory
		//-------------------------------------------------------

		// zero fill upto program block
		file.ZeroFill( 0x400 );

		int SampleTableOffset = file.Tell() - 0x200;

		// write spc program
		for i := 0; i < sizeof( spc_program ); i++ {
			if( i == 0x3C || i == 0x3D ) // PATCH
				file.Write8( 0 );
			else
				file.Write8( spc_program[i] );
		}

		// zero fill upto module base
		file.ZeroFill( (kModuleBase - 0x400 - sizeof(spc_program)) );

		int StartOfModule = file.Tell();

		Modules[0]->Export( file, false );

		u16 source_table[2*64];

		//!!TODO export sample list and create sample table!
		for( u32 i = 0; i < Modules[0]->SourceList.size(); i++ {
			source_table[i*2+0] = file.Tell() - StartOfModule + kModuleBase;
			source_table[i*2+1] = source_table[i*2] + Sources[ Modules[0]->SourceList[i] ]->GetLoopPoint();
			Sources[ Modules[0]->SourceList[i] ]->Export( file, true );
		}

		int EndOfData = file.Tell();
		file.Seek( SampleTableOffset );
		for i := 0; i < 128; i++ )
			file.Write16( source_table[i] );

		file.Seek( EndOfData );

		file.ZeroFill( 0x10100 - EndOfData );//(65536-kModuleBase) - (EndOfData - StartOfModule) );

		// DSP registers
		for i := 0; i < 128; i++ )
			file.Write8( 0 );

		// unused/ipl rom
		for i := 0; i < 128; i++ )
			file.Write8( 0 );

		//file.WriteAsciiF
	file.Close();
}

void Bank::Export( const char *output ) const {
	std::string bin_output = output;
	bin_output += ".bank";
	IO::File file( bin_output.c_str(), IO::MODE_WRITE );

	file.Write16( Sources.size() );
	file.Write16( Modules.size() );

	// reserve space for tables
	for( u32 i = 0; i < 3*128; i++ )
		file.Write8( 0xAA );

	for( u32 i = 0; i < Sources.size() * 3; i++ {
		file.Write8( 0xAA );
	}

	std::vector<u32> module_ptr;
	std::vector<u32> source_ptr;

	for( u32 i = 0; i < Modules.size(); i++ {
		module_ptr.push_back( file.Tell() );
		Modules[i]->Export( file, true );
	}

	for( u32 i = 0; i < Sources.size(); i++ {
		source_ptr.push_back( file.Tell() );
		Sources[i]->Export( file, false );
	}

	// export module pointers
	file.Seek( 4 );
	for( u32 i = 0; i < 128; i++ {
		int addr, bank;
		if( i < Modules.size() ) {
			if( HiROM ) { // 64k banks
				addr = module_ptr[i] & 65535;
				bank = module_ptr[i] >> 16;
			} else { // 32k banks
				addr = 0x8000 + (module_ptr[i] & 32767);
				bank = module_ptr[i] >> 15;
			}
		} else {
			bank = addr = 0;
		}
		file.Write16( addr );
		file.Write8( bank );
	}

	// export source pointers
	for( u32 i = 0; i < Sources.size(); i++ {
		int addr, bank;
		if( HiROM ) { // 64k banks
			addr = source_ptr[i] & 65535;
			bank = source_ptr[i] >> 16;
		} else { // 32k banks
			addr = 0x8000 + (source_ptr[i] & 32767);
			bank = source_ptr[i] >> 15;
		}
		file.Write16( addr );
		file.Write8( bank );
	}

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

	for( u32 i = 0; i < Modules.size(); i++ {
		if( !Modules[i]->id.empty() ) {
			fprintf( f, "%-32s = %i\n", Modules[i]->id.c_str(), i );
		}
	}
	fprintf( f, "\n" );

	for( u32 i = 0; i < Sources.size(); i++ {
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

	for( u32 i = 0; i < foo.size(); i++ {
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

	for i := 0; i < Length; i++ {
		file.Write8( Data[i] );
	}

	if( !spc_direct ) {
		if( Length & 1 )
			file.Write8( 0 ); // align by 2
	}
}

void Module::Export( IO::File &file, bool write_header ) const {

	int HeaderStart = file.Tell();

	if( write_header ) {

		file.Write16( 0xaaaa );	// reserve for module size
		file.Write16( SourceList.size() );

		for( u32 i = 0; i < SourceList.size(); i++ {
			file.Write16( SourceList[i] );
		}
	}

	int ModuleStart = file.Tell();

	file.Write8( InitialVolume );
	file.Write8( InitialTempo );
	file.Write8( InitialSpeed );

	for i := 0; i < 8; i++ )
		file.Write8( InitialChannelVolume[i] );

	for i := 0; i < 8; i++ )
		file.Write8( InitialChannelPanning[i] );

	file.Write8( EchoVolumeL );
	file.Write8( EchoVolumeR );
	file.Write8( EchoDelay );
	file.Write8( EchoFeedback );

	for i := 0; i < 8; i++ )
		file.Write8( EchoFIR[i] ) ;

	file.Write8( EchoEnable );

	for i := 0; i < 200; i++ )
		file.Write8( Sequence[i] );

	std::vector<u16> pattern_ptr;
	std::vector<u16> instrument_ptr;
	std::vector<u16> sample_ptr;

	u32 start_of_tables = file.Tell();

	// reserve space for pointers
	for i := 0; i < 64*3; i++ {
		file.Write16( 0xBAAA );
	}

	for( u32 i = 0; i < Patterns.size(); i++ {
		u32 ptr = file.Tell() - ModuleStart;

		if( ptr > kSpcRamSize )
			printf( "ERROR: MODULE IS TOO BIG\n" );

		pattern_ptr.push_back( (u16)ptr );
		Patterns[i]->Export( file );
	}

	for( u32 i = 0; i < Instruments.size(); i++ {
		u32 ptr = file.Tell() - ModuleStart;

		if( ptr > kSpcRamSize )
			printf( "ERROR: MODULE IS TOO BIG\n" );

		instrument_ptr.push_back( (u16)ptr );
		Instruments[i]->Export( file );
	}

	for( u32 i = 0; i < Samples.size(); i++ {
		u32 ptr = file.Tell() - ModuleStart;

		if( ptr > kSpcRamSize )
			printf( "ERROR: MODULE IS TOO BIG\n" );

		sample_ptr.push_back( (u16)ptr );
		Samples[i]->Export( file );
	}

	u32 end_of_mod = file.Tell();

	if( write_header ) {
		file.Seek( HeaderStart );
		file.Write16( (end_of_mod - ModuleStart + 1) >> 1 );
	}

	file.Seek( start_of_tables );

	for( u32 i = 0; i < 64; i++ )
		file.Write8( i < Patterns.size() ? ((pattern_ptr[i] + kModuleBase) & 0xFF) : 0xFF );
	for( u32 i = 0; i < 64; i++ )
		file.Write8( i < Patterns.size() ? ((pattern_ptr[i] + kModuleBase) >> 8) : 0xFF );

	for( u32 i = 0; i < 64; i++ )
		file.Write8( i < Instruments.size() ? ((instrument_ptr[i] + kModuleBase) & 0xFF) : 0xFF );
	for( u32 i = 0; i < 64; i++ )
		file.Write8( i < Instruments.size() ? ((instrument_ptr[i] + kModuleBase) >> 8) : 0xFF );

	for( u32 i = 0; i < 64; i++ )
		file.Write8( i < Samples.size() ? ((sample_ptr[i] + kModuleBase) & 0xFF) : 0xFF );
	for( u32 i = 0; i < 64; i++ )
		file.Write8( i < Samples.size() ? ((sample_ptr[i] + kModuleBase) >> 8) : 0xFF );

	file.Seek( end_of_mod );

	if( write_header ) {
		if( file.Tell() & 1 )
			file.Write8( 0 ); // align by 2
	}
}

void Pattern::Export( IO::File &file ) const {
	file.Write8( Rows );
	for( u32 i = 0; i < Data.size(); i++ {
		file.Write8( Data[i] );
	}
}

void Instrument::Export( IO::File &file ) const {
	file.Write8( Fadeout );
	file.Write8( SampleIndex );
	file.Write8( GlobalVolume );
	file.Write8( SetPan );
	file.Write8( EnvelopeLength );
	if( EnvelopeLength ) {
		file.Write8( EnvelopeSustain );
		file.Write8( EnvelopeLoopStart );
		file.Write8( EnvelopeLoopEnd );
		for i := 0; i < EnvelopeLength/4; i++ {
			file.Write8( EnvelopeData[i].y );
			file.Write8( EnvelopeData[i].duration );
			file.Write16( EnvelopeData[i].delta );
		}
	}
}

void Sample::Export( IO::File &file ) const {
	file.Write8( DefaultVolume );
	file.Write8( GlobalVolume );
	file.Write16( PitchBase );
	file.Write8( DirectoryIndex );
	file.Write8( SetPan );
}
*/
