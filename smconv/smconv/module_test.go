package smconv

import (
	"encoding/binary"
	"io"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mukunda.com/modlib"
)

func read16(file *os.File) uint16 {
	var result uint16
	binary.Read(file, binary.LittleEndian, &result)
	return result
}

func read8(file *os.File) uint8 {
	var result uint8
	binary.Read(file, binary.LittleEndian, &result)
	return result
}

func TestSoundbankExport(t *testing.T) {

	bank := SoundBank{}

	mod, _ := modlib.LoadModule("../test/pollen8.it")
	bank.AddModule(mod, "test/pollen8.it")
	assert.Equal(t, "MOD_POLLEN8", bank.Modules[0].Id)

	bank.Export(".testdata-pollen", false)

	file, err := os.Open(".testdata-pollen.smbank")
	assert.NoError(t, err)

	assert.EqualValues(t, 17, read16(file), "17 sources")
	assert.EqualValues(t, 1, read16(file), "1 module")

	modOffset := int(read16(file)) + (int(read8(file)) << 16)
	assert.EqualValues(t, 0x8000+0x184+17*3, modOffset, "soundbank header size")
	file.Seek(int64(modOffset), 0)

	// Todo: more testing...
}

func TestSoundbankExport2(t *testing.T) {
	bank := SoundBank{}

	mod, _ := modlib.LoadModule("test/reflection.it")
	bank.AddModule(mod, "test/reflection.it")
	assert.NoError(t, bank.Export(".testdata-reflection", false))

	f, _ := os.Open(".testdata-reflection.smbank")

	assert.EqualValues(t, 1, read16(f), "1 source")
	assert.EqualValues(t, 1, read16(f), "1 module")

	type sbOffset struct {
		Offset uint16
		Bank   uint8
	}

	var moduleOffset sbOffset

	binary.Read(f, binary.LittleEndian, &moduleOffset)

	// LoRom address.
	// TODO: the module start should be aligned by 2 bytes
	assert.EqualValues(t, 0x8000+0x184+3, int(moduleOffset.Offset))
	assert.EqualValues(t, 0, int(moduleOffset.Bank))

	startOfModuleHeader := int64(moduleOffset.Offset-0x8000) + int64(moduleOffset.Bank)*0x8000
	f.Seek(startOfModuleHeader, io.SeekStart)

	assert.NotZero(t, read16(f))        // Size of module in words
	assert.EqualValues(t, 1, read16(f)) // num sources used

	assert.EqualValues(t, 0, read16(f)) // source list (one entry)
	// TODO: Aren't we supposed to have an additional ZERO word for termination? according to soundbank.txt. Verify in SNES driver.
	startOfModule := startOfModuleHeader + 6
	var header SmModuleHeader
	binary.Read(f, binary.LittleEndian, &header)

	assert.EqualValues(t, 45, header.InitialVolume)
	assert.EqualValues(t, 135, header.InitialTempo)
	assert.EqualValues(t, 6, header.InitialSpeed)

	assert.EqualValues(t, 64, header.InitialChannelVolume[0])
	assert.EqualValues(t, 64, header.InitialChannelVolume[1])

	assert.EqualValues(t, 32, header.InitialChannelPanning[0])
	assert.EqualValues(t, 32, header.InitialChannelPanning[1])

	// TODO: set echo in reflection.it
	assert.EqualValues(t, 0, header.EchoVolumeL)
	assert.EqualValues(t, 0, header.EchoVolumeR)
	assert.EqualValues(t, 0, header.EchoDelay)
	assert.EqualValues(t, 0, header.EchoFeedback)
	assert.EqualValues(t, 0x7F, header.EchoFir[0])
	assert.EqualValues(t, 0, header.EchoFir[1])
	assert.EqualValues(t, 0, header.EchoEnable)

	// Sequence | 0 |+++|---| 0 |---|
	assert.Equal(t, []byte{0, 254, 255, 0, 255}, header.Sequence[0:5])

	pointers := SmModuleHeaderPointers{}
	binary.Read(f, binary.LittleEndian, &pointers)

	// Pointers include the offset of the module data in memory
	// kModuleBase
	for i := 0; i < 64; i++ {
		assert.True(t, pointers.InstrumentsH[i] >= kModuleBase>>8)
		assert.True(t, pointers.PatternsH[i] >= kModuleBase>>8)
		assert.True(t, pointers.SamplesH[i] >= kModuleBase>>8)
	}

	ptrToFileOffset := func(L uint8, H uint8) int {
		return int(startOfModule) + int(int(L)+(int(H)<<8)-kModuleBase)
	}

	patternOffset := ptrToFileOffset(pointers.PatternsL[0], pointers.PatternsH[0])
	f.Seek(int64(patternOffset), io.SeekStart)

	assert.EqualValues(t, 63, read8(f)) // num rows - 1

	{
		notes := []uint8{}
		instrs := []uint8{}
		volumes := []uint8{}
		effects := []uint8{}
		params := []uint8{}

		for row := 0; row < 64; row++ {
			assert.EqualValues(t, 0xff, read8(f)) // spc hints (unused)
			channelmask := read8(f)

			for channel := 0; channel < 8; channel++ {
				if channelmask&(1<<channel) == 0 {
					continue // no data for channel
				}

				updatemask := read8(f)

				// If a higher order bit 4-7 is set, then the corresponding lower order bit must
				// also be set. In other words, if a new note byte follows, then the note flag
				// must always be set. Bit4 = new note byte, Bit0 = note is present.
				//
				// If bit4 is unset and bit0 is set, then the last note value is used (no new
				// byte).
				assert.EqualValues(t, updatemask>>4, updatemask&(updatemask>>4))

				if updatemask&0x10 != 0 {
					notes = append(notes, read8(f)) // note
				} else if updatemask&0x01 != 0 {
					assert.NotEmpty(t, notes)
					notes = append(notes, notes[len(notes)-1])
				}

				if updatemask&0x20 != 0 {
					instrs = append(instrs, read8(f)) // instrument
				} else if updatemask&0x02 != 0 {
					assert.NotEmpty(t, instrs)
					instrs = append(instrs, instrs[len(instrs)-1])
				}

				if updatemask&0x40 != 0 {
					volumes = append(volumes, read8(f)) // volume
				} else if updatemask&0x04 != 0 {
					assert.NotEmpty(t, volumes)
					volumes = append(volumes, volumes[len(volumes)-1])
				}

				if updatemask&0x80 != 0 {
					effects = append(effects, read8(f)) // effect
					params = append(params, read8(f))   // effect parameter
				} else if updatemask&0x08 != 0 {
					assert.NotEmpty(t, effects)
					assert.NotEmpty(t, params)
					effects = append(effects, effects[len(effects)-1])
					params = append(params, params[len(params)-1])
				}

			}
		}

		assert.Equal(t, 16, len(notes))
		assert.Equal(t, 16, len(instrs))
		assert.Equal(t, 6, len(volumes))
		assert.Equal(t, 63, len(effects))
		assert.Equal(t, 63, len(params))
	}

	{
		// Verify instrument
		instrumentOffset := ptrToFileOffset(pointers.InstrumentsL[0], pointers.InstrumentsH[0])
		f.Seek(int64(instrumentOffset), io.SeekStart)

		var instr SmInstrument
		binary.Read(f, binary.LittleEndian, &instr.Info)

		// The data may not always contain the full instrument info. If the LENGTH of the
		// envelope is 0, then the sustain,loop,and data don't follow.

		// We should change that to be simpler and include those 3 bytes - not worth the
		// complexity to not.

		// Fadeout is stored as the IT value / 4
		assert.EqualValues(t, 1, instr.Info.Fadeout)
		assert.EqualValues(t, 0, instr.Info.SampleIndex)

		// Volume and panning match IT.
		assert.EqualValues(t, 126, instr.Info.GlobalVolume)
		assert.EqualValues(t, 128|33, instr.Info.SetPanning)

		// The loop parameters are in multiples of 4 (bytes). They are direct byte offsets
		// in the node data.
		assert.EqualValues(t, 3*4, instr.Info.EnvelopeLength)
		assert.EqualValues(t, 0x80, instr.Info.EnvelopeSustain) // 128=disabled
		assert.EqualValues(t, 0, instr.Info.EnvelopeLoopStart)
		assert.EqualValues(t, 2*4, instr.Info.EnvelopeLoopEnd)

		instr.Envelope = make([]SmEnvelopeNode, instr.Info.EnvelopeLength/4)
		binary.Read(f, binary.LittleEndian, instr.Envelope)

		// Delta is computed as delta / duration * 256, rounded.
		assert.Equal(t, []SmEnvelopeNode{
			{Level: 32, Duration: 9, Delta: ((51-32)*256 + (9 / 2)) / 9},
			{Level: 51, Duration: 2, Delta: ((4-51)*256 + (2 / 2)) / 2},
			{Level: 4, Duration: 0, Delta: 0},
		}, instr.Envelope)
	}

	{
		// Verify sample
		sampleOffset := ptrToFileOffset(pointers.SamplesL[0], pointers.SamplesH[0])
		f.Seek(int64(sampleOffset), io.SeekStart)

		var sample SmSample
		binary.Read(f, binary.LittleEndian, &sample)

		assert.EqualValues(t, 64, sample.DefaultVolume)
		assert.EqualValues(t, 64, sample.GlobalVolume)

		// PitchBase is missing documentation, but it should be calculated as:
		// f = C5speed * TuningFactor
		// round(log2(f/8363) * 768)
		// From what I remember, it's an index into a frequency lookup table.
		// The lookup table is used to map linear values into frequencies.

		assert.EqualValues(t, math.Round(math.Log2(8363.0/8363.0)*768.0), sample.PitchBase)
		assert.EqualValues(t, 0, sample.DirectoryIndex)
		assert.EqualValues(t, 128|32, sample.SetPanning)
	}

	{
		// Source data
		f.Seek(0x184, io.SeekStart)
		var sourceOffset sbOffset
		binary.Read(f, binary.LittleEndian, &sourceOffset)

		// For LoRom, it should be 0x80 or higher offsets.
		assert.True(t, sourceOffset.Offset >= 0x80)
		assert.EqualValues(t, 0, int(sourceOffset.Bank))

		startOfSourceHeader := int64(sourceOffset.Offset-0x8000) + int64(sourceOffset.Bank)*0x8000
		f.Seek(startOfSourceHeader, io.SeekStart)

		realSampleLength := 64
		brrLength := realSampleLength / 16 * 9
		assert.EqualValues(t, brrLength, read16(f)) // length of data, 9 bytes per 16 samples
		assert.EqualValues(t, 0, read16(f))         // loop offset

		data := make([]byte, brrLength)
		err := binary.Read(f, binary.LittleEndian, data)
		assert.NoError(t, err)

		// BRR data should contain the END bit in the last block only
		for i := 0; i < len(data); i += 9 {
			if i == len(data)-9 {
				assert.EqualValues(t, 1, data[i]&1)
			} else {
				assert.EqualValues(t, 0, data[i]&1)
			}
		}

		// BRR data should contain the LOOP bit in the last block only, if the sample loops
		// (this sample does loop)
		for i := 0; i < len(data); i += 9 {
			if i == len(data)-9 {
				assert.EqualValues(t, 2, data[i]&2)
			} else {
				assert.EqualValues(t, 0, data[i]&2)
			}
		}

		// The first BRR block must use filter 0
		assert.EqualValues(t, 0, data[0]&(3<<2))

		// The loop start must also use filter 0 (but it's
		// also the first block in this case).

		// There should be data present.
		dataSum := 0
		for i := 0; i < len(data); i++ {
			if i%9 == 0 {
				continue

			}
			dataSum += int(data[i])
		}
		assert.NotZero(t, dataSum)

	}
}
