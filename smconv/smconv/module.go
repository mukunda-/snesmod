// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

// This file describes the conversion process between the common module format (modlib)
// and the structures for SNESMOD.

package smconv

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"go.mukunda.com/modlib"
	"go.mukunda.com/modlib/common"
)

var ErrInvalidSmOptions = errors.New("invalid snesmod options")

// See soundbank.txt for binary format documentation.

// Break up the structure to work with encoding/binary.
type SmModule struct {
	Id          string
	BankHeader  SmModuleBankHeader
	SourceList  []uint16
	Header      SmModuleHeader
	Patterns    []*SmPattern
	Instruments []*SmInstrument
	Samples     []*SmSample

	// Warnings gathered during conversion.
	Warnings []string

	// Metadata (used for SPC)
	Title       string
	Author      string
	SongMessage string
}

func (smm *SmModule) warn(msg string) {
	smm.Warnings = append(smm.Warnings, msg)
}

// SmModule is a SNESMOD module stored in a cartridge ROM area.
type SmModuleBankHeader struct {
	ModuleSize      uint16 // Measured in words (bytes/2)
	SourceListCount uint16
}

type SmModuleHeader struct {
	InitialVolume         uint8
	InitialTempo          uint8
	InitialSpeed          uint8
	InitialChannelVolume  [8]uint8
	InitialChannelPanning [8]uint8

	EchoVolumeL  int8
	EchoVolumeR  int8
	EchoDelay    uint8
	EchoFeedback int8
	EchoFir      [8]int8
	EchoEnable   uint8

	Sequence [200]uint8
}

type SmModuleHeaderPointers struct {
	PatternsL    [64]byte
	PatternsH    [64]byte
	InstrumentsL [64]byte
	InstrumentsH [64]byte
	SamplesL     [64]byte
	SamplesH     [64]byte
}

type SmPattern struct {
	Rows uint8
	Data []byte
}

type SmInstrument struct {
	Info     SmInstrumentInfo
	Envelope []SmEnvelopeNode
}

type SmEnvelopeNode struct {
	// Y value
	Level uint8

	// Ticks until next node
	Duration uint8

	// How much to add to the Level each tick.
	Delta int16
}

type SmInstrumentInfo struct {
	Fadeout           uint8
	SampleIndex       uint8
	GlobalVolume      uint8
	SetPanning        uint8
	EnvelopeLength    uint8
	EnvelopeSustain   uint8
	EnvelopeLoopStart uint8
	EnvelopeLoopEnd   uint8
}

type SmSample struct {
	DefaultVolume  uint8  // 0-64
	GlobalVolume   uint8  // 0-64
	PitchBase      uint16 // ???
	DirectoryIndex uint8  // Index into source directory
	SetPanning     uint8  // 0-64, &128 = disabled
}

func (smm *SmModule) readSmoIntArgs(tokens []string, minargs int, maxargs int, minval int, maxval int) []int {
	cmd := tokens[0]
	args := []int{}

	if len(tokens) < 1+minargs {
		smm.warn("Not enough params for " + cmd + " command.")
		return nil
	}

	if len(tokens) > 1+maxargs {
		smm.warn("Too many params for " + cmd + " command.")
		return nil
	}

	for i := 1; i < len(tokens); i++ {
		arg, err := strconv.Atoi(tokens[i])
		if err != nil {
			smm.warn("Error parsing " + cmd + " command value.")
			return nil
		}
		if arg < minval {
			smm.warn(fmt.Sprintf("%s value out of range: %d", cmd, arg))
			arg = minval
		}
		if arg > maxval {
			smm.warn(fmt.Sprintf("%s value out of range: %d", cmd, arg))
			arg = int(maxval)
		}

		args = append(args, arg)
	}

	return args
}

// Returns text that is found between [[section]] and [[/section]]. If the closing tag
// is not present, then the remainder of the text is selected.
func getTextSection(text, section string) string {

	start := strings.ToLower("[[" + section + "]]")
	end := strings.ToLower("[[/" + section + "]]")

	startIdx := strings.Index(strings.ToLower(text), start)
	if startIdx == -1 {
		return ""
	}

	text = text[startIdx+len(start):]

	endIdx := strings.Index(strings.ToLower(text), end)
	if endIdx == -1 {
		return text
	}

	return text[:endIdx]
}

var reStripSnesmod = regexp.MustCompile(`(?i)\[\[snesmod\]\][\S\s]*?(\[\[/snesmod\]\]|$)`)

// Returns text with the [[SNESMOD]] section removed.
func stripSnesmodTag(text string) string {
	return reStripSnesmod.ReplaceAllString(text, "")
}

// Parse the Message in the mod and load any parameters found into the SmModule.
func (smm *SmModule) parseSmOptions(mod *modlib.Module) {
	smm.SongMessage = stripSnesmodTag(mod.Message)

	text := getTextSection(mod.Message, "snesmod")

	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		tokens := strings.Fields(line)

		switch strings.ToLower(tokens[0]) {
		case "edl":
			if args := smm.readSmoIntArgs(tokens, 1, 1, 0, 15); args != nil {
				smm.Header.EchoDelay = uint8(args[0])
			}
		case "efb":
			if args := smm.readSmoIntArgs(tokens, 1, 1, -128, 127); args != nil {
				smm.Header.EchoFeedback = int8(args[0])
			}
		case "evol":
			if args := smm.readSmoIntArgs(tokens, 1, 2, -128, 127); args != nil {
				smm.Header.EchoVolumeL = int8(args[0])
				if len(args) == 2 {
					smm.Header.EchoVolumeR = int8(args[1])
				} else {
					smm.Header.EchoVolumeR = int8(args[0])
				}
			}
		case "efir":
			if args := smm.readSmoIntArgs(tokens, 1, 8, -128, 127); args != nil {
				for i := 0; i < 8; i++ {
					smm.Header.EchoFir[i] = int8(args[i])
				}
			}
		case "eon":
			if args := smm.readSmoIntArgs(tokens, 1, 8, 1, 8); args != nil {
				enabled := 0
				for i := 0; i < len(args); i++ {
					enabled |= (1 << (args[i] - 1))
				}
				smm.Header.EchoEnable = uint8(enabled)
			}
		}
	}
}

func convertModule(mod *modlib.Module, filename string, sourceList []SourceIndex, sampleDirectory []uint8, sources []*Source) *SmModule {
	var smm = new(SmModule)

	// Metadata for SPC
	//smm.Author= mod.Author
	smm.Title = mod.Title

	smm.Id = pathToId("MOD_", filename)
	smm.Header.InitialVolume = uint8(mod.GlobalVolume)
	smm.Header.InitialTempo = uint8(mod.InitialTempo)
	smm.Header.InitialSpeed = uint8(mod.InitialSpeed)

	smm.SourceList = sourceList
	smm.BankHeader.SourceListCount = uint16(len(sourceList))

	if mod.Channels > 8 {
		smm.warn("module has too many channels")
	}

	for i := 0; i < 8; i++ {
		if i < len(mod.ChannelSettings) {
			smm.Header.InitialChannelVolume[i] = uint8(mod.ChannelSettings[i].InitialVolume)
			smm.Header.InitialChannelPanning[i] = uint8(mod.ChannelSettings[i].InitialPan)
		}
	}

	smm.Header.EchoFir[0] = 127
	smm.parseSmOptions(mod)

	for i := 0; i < 200; i++ {
		if i < len(mod.Order) {
			smm.Header.Sequence[i] = uint8(mod.Order[i])
		} else {
			smm.Header.Sequence[i] = 255
		}
	}

	for _, pattern := range mod.Patterns {
		// Convert patterns
		smp := convertPattern(&pattern)
		smm.Patterns = append(smm.Patterns, smp)
	}

	for _, instr := range mod.Instruments {
		// Convert instruments
		smi := convertInstrument(&instr)
		smm.Instruments = append(smm.Instruments, smi)
	}

	for i, sample := range mod.Samples {
		// Convert samples
		sms := convertSample(&sample, sampleDirectory[i], sources[sourceList[sampleDirectory[i]]].TuningFactor)
		smm.Samples = append(smm.Samples, sms)
	}

	return smm
}

const (
	VcmdSetVolume      = 1
	VcmdFineVolUp      = 2
	VcmdFineVolDown    = 3
	VcmdVolSlideUp     = 4
	VcmdVolSlideDown   = 5
	VcmdPitchSlideDown = 6
	VcmdPitchSlideUp   = 7
	VcmdSetPan         = 8
	VcmdPortaToNote    = 9
	VcmdVibratoDepth   = 10
)

func vCmdToSmByte(command uint8, param uint8) uint8 {
	// Assuming all ranges are valid.

	// Same as IT
	switch command {
	case VcmdSetVolume:
		return param
	case VcmdFineVolUp:
		return 65 + param
	case VcmdFineVolDown:
		return 75 + param
	case VcmdVolSlideUp:
		return 85 + param
	case VcmdVolSlideDown:
		return 95 + param
	case VcmdPitchSlideDown:
		return 105 + param
	case VcmdPitchSlideUp:
		return 115 + param
	case VcmdSetPan:
		return 128 + param
	case VcmdPortaToNote:
		return 193 + param
	case VcmdVibratoDepth:
		return 203 + param
	}

	// TODO: does the driver handle 255 as noop? or will it crash?
	// This should never be reached in a valid modlib.Module
	return 255
}

func noteToSmNote(note uint8) uint8 {
	// Same as IT
	// common format is 1-120, 253, 254, 255
	// translate to IT format 0-119, 253, 254, 255

	if note > 120 {
		return note
	}

	// Remove note offset, 0 = C-0
	return note - 1
}

func convertPattern(pattern *common.Pattern) *SmPattern {
	var smp = new(SmPattern)

	smp.Rows = uint8(len(pattern.Rows) - 1)
	data := []byte{}

	prevNote := [8]int16{-1, -1, -1, -1, -1, -1, -1, -1}
	prevIns := [8]int16{-1, -1, -1, -1, -1, -1, -1, -1}
	prevVcmd := [8]int16{-1, -1, -1, -1, -1, -1, -1, -1}
	prevEffect := [8]int16{-1, -1, -1, -1, -1, -1, -1, -1}
	prevParam := [8]int16{-1, -1, -1, -1, -1, -1, -1, -1}

	for i := 0; i < len(pattern.Rows); i++ {
		row := pattern.Rows[i]

		noteHints := 0xFF // Not used currently. I believe this is for volume ramping? Can't remember.
		updateBits := 0

		for _, entry := range row.Entries {
			if entry.Channel >= 8 {
				continue
			}
			updateBits |= (1 << entry.Channel)
		}

		data = append(data, uint8(noteHints), uint8(updateBits))

		for _, entry := range row.Entries {
			mask := uint8(0)
			channelData := []byte{}

			if entry.Note != 0 {
				mask |= 1
				note := noteToSmNote(entry.Note)
				if int16(note) != prevNote[entry.Channel] {
					mask |= 16
					channelData = append(channelData, note)
					prevNote[entry.Channel] = int16(note)
				}
			}

			if entry.Instrument != 0 {
				mask |= 2
				ins := uint8(entry.Instrument)
				if int16(ins) != prevIns[entry.Channel] {
					mask |= 32
					channelData = append(channelData, ins)
					prevIns[entry.Channel] = int16(ins)
				}
			}

			if entry.VolumeCommand != 0 {
				mask |= 4
				vCmd := vCmdToSmByte(entry.VolumeCommand, entry.VolumeParam)
				if int16(vCmd) != prevVcmd[entry.Channel] {
					mask |= 64
					channelData = append(channelData, vCmd)
					prevVcmd[entry.Channel] = int16(vCmd)
				}
			}

			if entry.Effect != 0 {
				mask |= 8
				eff := uint8(entry.Effect)
				param := uint8(entry.EffectParam)
				if int16(eff) != prevEffect[entry.Channel] || int16(param) != prevParam[entry.Channel] {
					mask |= 128
					channelData = append(channelData, eff, param)
					prevEffect[entry.Channel] = int16(eff)
					prevParam[entry.Channel] = int16(param)
				}
			}

			data = append(data, mask)
			data = append(data, channelData...)
		}

	}

	smp.Data = data

	return smp
}

func convertInstrument(instr *common.Instrument) *SmInstrument {
	var smi = new(SmInstrument)

	// Fadeout is /4 according to the driver documentation.
	// Note this removes a lot of granularity.  The IT format
	// will have these values as low as "8" by default, which
	// would end up as "2" here.
	smi.Info.Fadeout = uint8(min(instr.Fadeout/4, 255))

	smi.Info.SampleIndex = uint8(instr.Notemap[60].Sample - 1)
	smi.Info.GlobalVolume = uint8(instr.GlobalVolume)
	smi.Info.SetPanning = uint8(instr.DefaultPan)
	if !instr.DefaultPanEnabled {
		smi.Info.SetPanning |= 128
	}

	// Convert the envelope if it exists. Only volume is supported.
	var env *common.Envelope
	for _, ienv := range instr.Envelopes {
		if ienv.Type == common.EnvelopeTypeVolume {
			env = &ienv
			break
		}
	}

	if env != nil {

		// Initialize to 0x80 (disabled)
		smi.Info.EnvelopeSustain = 0xff
		smi.Info.EnvelopeLoopStart = 0xff
		smi.Info.EnvelopeLoopEnd = 0xff

		if env.Sustain {
			// Only a single sustain point is supported by the driver.
			smi.Info.EnvelopeSustain = uint8(env.SustainStart * 4)
		}

		envLength := len(env.Nodes)

		if env.Loop {
			smi.Info.EnvelopeLoopStart = uint8(env.LoopStart * 4)
			smi.Info.EnvelopeLoopEnd = uint8(env.LoopEnd * 4)

			// Cut to the end of the envelope - it's impossible to move past the loop.
			envLength = int(env.LoopEnd + 1)
		}

		smi.Info.EnvelopeLength = uint8(envLength * 4)
		smi.Envelope = []SmEnvelopeNode{}
		for i := 0; i < envLength; i++ {
			node := SmEnvelopeNode{
				Level: uint8(env.Nodes[i].Y),
			}

			if i != envLength-1 {
				node.Duration = uint8(env.Nodes[i+1].X - env.Nodes[i].X)
				if node.Duration == 0 {
					node.Duration = 1
				}
				node.Delta = int16((int(env.Nodes[i+1].Y-env.Nodes[i].Y)*256 + int(node.Duration/2)) / int(node.Duration))
			} else {
				node.Duration = 0
				node.Delta = 0
			}

			smi.Envelope = append(smi.Envelope, node)
		}

	}

	return smi
}

func convertSample(sample *common.Sample, directoryIndex uint8, tuning float64) *SmSample {
	var sms = new(SmSample)

	sms.DefaultVolume = uint8(sample.DefaultVolume)
	sms.GlobalVolume = uint8(sample.GlobalVolume)
	sms.SetPanning = uint8(sample.DefaultPanning ^ 128)

	a := float64(sample.C5) * tuning
	sms.PitchBase = uint16(math.Round(math.Log2(a/8363.0) * 768.0))

	sms.DirectoryIndex = directoryIndex

	return sms
}

/*

	if( VERBOSE ) {
		int pattsize = 0;
		for i := 0; i < mod.PatternCount; i++ {
			pattsize += Patterns[i]->GetExportSize();
		}
		int sampsize = 0;
		for( u32 i = 0; i < source_list.size(); i++ {
			sampsize += sources[source_list[i]]->GetDataLength();
		}
		int othersize = 0;
		for (u32 i = 0; i < Instruments.size(); i++ {
			othersize += Instruments[i]->GetExportSize();
		}
		othersize += GetExportSize_Header();

		int echosize = EchoDelay * 2048;
		int totalsize = pattsize + sampsize + othersize + echosize;
		printf(
			"Conversion report:\n"
			"Length: %i\n"
			"Patterns: %i\n"
			"Instruments: %i\n"
			"Samples: %i\n"
			"Pattern data size: %i bytes\n"
			"Sample data size: %i bytes\n"
			"Instruments/Other data size: %i bytes\n"
			"Echo region size: %i bytes\n"
			"Total size: %i bytes\n"
			,mod.Length,
			mod.PatternCount,
			mod.InstrumentCount,
			mod.SampleCount,
			pattsize,
			sampsize,
			othersize,
			echosize,
			totalsize
			);
		if( totalsize > kSpcRamSize ) {
			printf( "MODULE IS TOO BIG. maximum is %i bytes\n", kSpcRamSize );
		}
	}
}*/
