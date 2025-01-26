// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package smconv

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math"

	"go.mukunda.com/modlib/common"
	"go.mukunda.com/snesbrr/v2/brr"
)

const (
	kMaxUnrollThreshold = 2000
	kLoopLossTolerance  = 30000
)

// const (
// 	kSampHeadEnd	=1
// 	SAMPHEAD_LOOP	=2
// 	SAMPHEAD_FILTER =12
// 	SAMPHEAD_RANGE	=240
// )

// A source is a term for an audio sample in the soundbank.
type Source struct {
	// Loop start point in bytes (should be a multiple of 9, pointing to the BRR header).
	Loop int

	Hash string

	// BRR-compressed data
	Data []byte

	TuningFactor float64
	ID           string
}

var ErrUnsupportedSampleProperties = errors.New("unsupported sample properties")

func createSource(modsamp common.Sample) (*Source, error) {
	source := &Source{
		TuningFactor: 1.0,
	}

	var sampleData []int16

	if modsamp.Data.Bits == 16 {
		sampleData = modsamp.Data.Data[0].([]int16)
	} else if modsamp.Data.Bits == 8 {
		i8data := modsamp.Data.Data[0].([]int8)
		sampleData = make([]int16, len(i8data))
		for i := 0; i < len(i8data); i++ {
			// upsample to 16bit
			sampleData[i] = int16((int(i8data[i]) * 32767) / 128)
		}
	} else {
		return source, ErrUnsupportedSampleProperties
	}

	length := len(sampleData)

	if length == 0 {
		return source, nil
	}

	loopStart := modsamp.LoopStart
	loopLength := modsamp.LoopEnd - modsamp.LoopStart
	if !modsamp.Loop {
		loopLength = 0
	}

	if loopLength > 0 {
		// Discard data after loop end.
		length = min(length, loopStart+loopLength)
	}

	if modsamp.PingPong {
		// Unroll BIDI loop.
		for i := loopStart + loopLength - 1; i >= loopStart; i-- {
			sampleData = append(sampleData, sampleData[i])
		}
		loopLength *= 2
	}

	tuningFactor := 1.0

	if loopLength != 0 {
		if loopLength&0xF != 0 {

			unrollsNeeded := 0
			for ll := loopLength; ll&15 != 0; ll += loopLength {
				unrollsNeeded++
			}

			if loopLength*(1+unrollsNeeded) < kMaxUnrollThreshold {
				// Unroll the loop to align.
				// BrrCodec will handle this.
			} else {
				tuningFactor, sampleData, length, loopStart = resampleLoop(sampleData, loopStart, length, 16-(loopLength&15))
			}

		}
	}

	// Padding is handled by BRR Codec

	codec := brr.NewCodec()
	codec.PcmData = sampleData
	if loopLength > 0 {
		codec.SetLoop(loopStart)
	}
	codec.Encode()

	source.Loop = loopStart / 16 * 9
	source.Data = codec.BrrData
	source.TuningFactor = tuningFactor

	hash := sha256.Sum256(source.Data)
	source.Hash = fmt.Sprintf("%x", hash[:])

	return source, nil
}

// Add `amount` samples to the loop region and return the new size and loop start.
func resampleLoop(data []int16, loopStart int, length int, amount int) (tuning float64, resampledData []int16, newLength int, newLoopStart int) {

	oldLength := length
	oldLoopLength := length - loopStart
	newLoopLength := oldLoopLength + amount
	resampleFactor := float64(newLoopLength) / float64(oldLoopLength)
	iResampleFactor := 1 / resampleFactor
	newLength = int(math.Round(float64(length) * resampleFactor))
	newLoopStart = newLength - newLoopLength
	resampledData = make([]int16, newLength)

	// resample with linear interpolation (in the future, use cubic or something?)
	for x := 0; x < newLength; x++ {
		// For each sample in the new data:

		index := float64(x) * iResampleFactor
		index1 := int(math.Floor(index))
		index2 := index1 + 1
		if index1 >= oldLength {
			// I don't think this will ever happen.
			index1 -= oldLoopLength
		}
		if index2 >= oldLength {
			index2 -= oldLoopLength
		}

		s1 := data[index1]
		s2 := data[index2]
		delta := s2 - s1

		resamp := float64(s1) + float64(delta)*(index-float64(index1))
		if resamp < -32768 {
			resamp = -32768
		} else if resamp > 32767 {
			resamp = 32767
		}

		resampledData[x] = int16(resamp)
	}

	return iResampleFactor, resampledData, newLength, newLoopStart
}
