// This is BrrCodec.cpp by DMV47 converted to Golang.
// BrrCodec.cpp is licensed under Common Development and Distribution License, Version 1.0 only.
package brr

import (
	"io"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

var gauss_table = []int16{
	0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000,
	0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000,
	0x001, 0x001, 0x001, 0x001, 0x001, 0x001, 0x001, 0x001,
	0x001, 0x001, 0x001, 0x002, 0x002, 0x002, 0x002, 0x002,
	0x002, 0x002, 0x003, 0x003, 0x003, 0x003, 0x003, 0x004,
	0x004, 0x004, 0x004, 0x004, 0x005, 0x005, 0x005, 0x005,
	0x006, 0x006, 0x006, 0x006, 0x007, 0x007, 0x007, 0x008,
	0x008, 0x008, 0x009, 0x009, 0x009, 0x00A, 0x00A, 0x00A,
	0x00B, 0x00B, 0x00B, 0x00C, 0x00C, 0x00D, 0x00D, 0x00E,
	0x00E, 0x00F, 0x00F, 0x00F, 0x010, 0x010, 0x011, 0x011,
	0x012, 0x013, 0x013, 0x014, 0x014, 0x015, 0x015, 0x016,
	0x017, 0x017, 0x018, 0x018, 0x019, 0x01A, 0x01B, 0x01B,
	0x01C, 0x01D, 0x01D, 0x01E, 0x01F, 0x020, 0x020, 0x021,
	0x022, 0x023, 0x024, 0x024, 0x025, 0x026, 0x027, 0x028,
	0x029, 0x02A, 0x02B, 0x02C, 0x02D, 0x02E, 0x02F, 0x030,
	0x031, 0x032, 0x033, 0x034, 0x035, 0x036, 0x037, 0x038,
	0x03A, 0x03B, 0x03C, 0x03D, 0x03E, 0x040, 0x041, 0x042,
	0x043, 0x045, 0x046, 0x047, 0x049, 0x04A, 0x04C, 0x04D,
	0x04E, 0x050, 0x051, 0x053, 0x054, 0x056, 0x057, 0x059,
	0x05A, 0x05C, 0x05E, 0x05F, 0x061, 0x063, 0x064, 0x066,
	0x068, 0x06A, 0x06B, 0x06D, 0x06F, 0x071, 0x073, 0x075,
	0x076, 0x078, 0x07A, 0x07C, 0x07E, 0x080, 0x082, 0x084,
	0x086, 0x089, 0x08B, 0x08D, 0x08F, 0x091, 0x093, 0x096,
	0x098, 0x09A, 0x09C, 0x09F, 0x0A1, 0x0A3, 0x0A6, 0x0A8,
	0x0AB, 0x0AD, 0x0AF, 0x0B2, 0x0B4, 0x0B7, 0x0BA, 0x0BC,
	0x0BF, 0x0C1, 0x0C4, 0x0C7, 0x0C9, 0x0CC, 0x0CF, 0x0D2,
	0x0D4, 0x0D7, 0x0DA, 0x0DD, 0x0E0, 0x0E3, 0x0E6, 0x0E9,
	0x0EC, 0x0EF, 0x0F2, 0x0F5, 0x0F8, 0x0FB, 0x0FE, 0x101,
	0x104, 0x107, 0x10B, 0x10E, 0x111, 0x114, 0x118, 0x11B,
	0x11E, 0x122, 0x125, 0x129, 0x12C, 0x130, 0x133, 0x137,
	0x13A, 0x13E, 0x141, 0x145, 0x148, 0x14C, 0x150, 0x153,
	0x157, 0x15B, 0x15F, 0x162, 0x166, 0x16A, 0x16E, 0x172,
	0x176, 0x17A, 0x17D, 0x181, 0x185, 0x189, 0x18D, 0x191,
	0x195, 0x19A, 0x19E, 0x1A2, 0x1A6, 0x1AA, 0x1AE, 0x1B2,
	0x1B7, 0x1BB, 0x1BF, 0x1C3, 0x1C8, 0x1CC, 0x1D0, 0x1D5,
	0x1D9, 0x1DD, 0x1E2, 0x1E6, 0x1EB, 0x1EF, 0x1F3, 0x1F8,
	0x1FC, 0x201, 0x205, 0x20A, 0x20F, 0x213, 0x218, 0x21C,
	0x221, 0x226, 0x22A, 0x22F, 0x233, 0x238, 0x23D, 0x241,
	0x246, 0x24B, 0x250, 0x254, 0x259, 0x25E, 0x263, 0x267,
	0x26C, 0x271, 0x276, 0x27B, 0x280, 0x284, 0x289, 0x28E,
	0x293, 0x298, 0x29D, 0x2A2, 0x2A6, 0x2AB, 0x2B0, 0x2B5,
	0x2BA, 0x2BF, 0x2C4, 0x2C9, 0x2CE, 0x2D3, 0x2D8, 0x2DC,
	0x2E1, 0x2E6, 0x2EB, 0x2F0, 0x2F5, 0x2FA, 0x2FF, 0x304,
	0x309, 0x30E, 0x313, 0x318, 0x31D, 0x322, 0x326, 0x32B,
	0x330, 0x335, 0x33A, 0x33F, 0x344, 0x349, 0x34E, 0x353,
	0x357, 0x35C, 0x361, 0x366, 0x36B, 0x370, 0x374, 0x379,
	0x37E, 0x383, 0x388, 0x38C, 0x391, 0x396, 0x39B, 0x39F,
	0x3A4, 0x3A9, 0x3AD, 0x3B2, 0x3B7, 0x3BB, 0x3C0, 0x3C5,
	0x3C9, 0x3CE, 0x3D2, 0x3D7, 0x3DC, 0x3E0, 0x3E5, 0x3E9,
	0x3ED, 0x3F2, 0x3F6, 0x3FB, 0x3FF, 0x403, 0x408, 0x40C,
	0x410, 0x415, 0x419, 0x41D, 0x421, 0x425, 0x42A, 0x42E,
	0x432, 0x436, 0x43A, 0x43E, 0x442, 0x446, 0x44A, 0x44E,
	0x452, 0x455, 0x459, 0x45D, 0x461, 0x465, 0x468, 0x46C,
	0x470, 0x473, 0x477, 0x47A, 0x47E, 0x481, 0x485, 0x488,
	0x48C, 0x48F, 0x492, 0x496, 0x499, 0x49C, 0x49F, 0x4A2,
	0x4A6, 0x4A9, 0x4AC, 0x4AF, 0x4B2, 0x4B5, 0x4B7, 0x4BA,
	0x4BD, 0x4C0, 0x4C3, 0x4C5, 0x4C8, 0x4CB, 0x4CD, 0x4D0,
	0x4D2, 0x4D5, 0x4D7, 0x4D9, 0x4DC, 0x4DE, 0x4E0, 0x4E3,
	0x4E5, 0x4E7, 0x4E9, 0x4EB, 0x4ED, 0x4EF, 0x4F1, 0x4F3,
	0x4F5, 0x4F6, 0x4F8, 0x4FA, 0x4FB, 0x4FD, 0x4FF, 0x500,
	0x502, 0x503, 0x504, 0x506, 0x507, 0x508, 0x50A, 0x50B,
	0x50C, 0x50D, 0x50E, 0x50F, 0x510, 0x511, 0x511, 0x512,
	0x513, 0x514, 0x514, 0x515, 0x516, 0x516, 0x517, 0x517,
	0x517, 0x518, 0x518, 0x518, 0x518, 0x518, 0x519, 0x519,
}

type BrrCodec struct {
	loop_start         uint32
	loop_enabled       bool
	gauss_enabled      bool
	user_pitch_enabled bool
	pitch_step_base    uint16
	input_sample_rate  uint32
	output_sample_rate uint32
	wav_data           []int16
	brr_data           []uint8
	cb_func            func(*BrrCodec)
	cb_data            interface{}
	cur_progress       uint
	last_progress      uint
	cb_progress        uint
	total_blocks       uint32
	total_error        float64
	avg_error          float64
	min_error          float64
	max_error          float64
}

func NewBrrCodec() *BrrCodec {
	codec := BrrCodec{}
	codec.reset()
	return &codec
}

func (bc *BrrCodec) reset_progress() {
	bc.cur_progress = 0
	bc.last_progress = 0
	bc.cb_progress = 0
	if bc.cb_func != nil {
		bc.cb_func(bc)
	}
}

func (bc *BrrCodec) set_progress(n uint) {
	if bc.cb_func != nil {
		bc.cur_progress = n
		if bc.cur_progress != bc.last_progress {
			bc.cb_func(bc)
			bc.last_progress = bc.cur_progress
		}
	}
}

func (bc *BrrCodec) reset() {
	*bc = BrrCodec{} // Rather than initializing everything, use zero initialization and only set what's needed.
	bc.pitch_step_base = 0x1000

	bc.input_sample_rate = 32000
	bc.output_sample_rate = 32000
}

func clamp[T int8 | int16 | int32 | int64](value T, bits int) T {
	var low T = -1 << (bits - 1)
	var high T = (1 << (bits - 1)) - 1

	if value > high {
		return high
	} else if value < low {
		return low
	}

	return value
}

func (bc *BrrCodec) decode() {

	if !bc.gauss_enabled {
		// 7.8125 = 32000 / 0x1000
		bc.output_sample_rate = uint32(float64(bc.pitch_step_base)*7.8125 + 0.5)
	}

	bc.wav_data = []int16{}

	if len(bc.brr_data) == 0 {
		return
	}

	// If the length of the data isn't a multiple of 9 (9 bytes per block), pad it with
	// zero bytes.
	for len(bc.brr_data)%9 != 0 {
		bc.brr_data = append(bc.brr_data, 0)
	}
	// Make sure that the last block has the "END" flag set.
	bc.brr_data[len(bc.brr_data)-9] |= 1

	data := 0
	sample := [8]int16{} // 4 samples stored twice
	last_sample := [2]int16{}
	var header uint8
	var samp_i uint
	var brr_counter uint = 1 // --1 == 0
	var pitch int32 = 0x3000 // decode 4 samples

	for {
		for pitch >= 0 {
			pitch -= 0x1000

			brr_counter--
			if brr_counter == 0 {
				// End of block

				if header&1 != 0 {
					// End of sample (END set)
					return
				}

				header = bc.brr_data[data]
				data++
				brr_counter = 16

				if (header & 3) == 1 {
					// TOOD: doesn't make sense why this is checked here. Wouldn't this skip the last sample?
					return
				}
			}

			var brange uint8 = header >> 4
			var filter uint8 = (header >> 2) & 3

			var samp uint8 = bc.brr_data[data]
			var s int32

			// the high nybble is decoded before the low nybble
			if (brr_counter & 1) == 1 {
				data++
				s = int32((samp&0x0F)^8) - 8
			} else {
				s = int32((samp>>4)^8) - 8
			}

			if brange > 12 {
				s &= ^0x07FF // -2048 or 0
			} else {
				s = (s << brange) >> 1 // -16384 to 16383
			}

			// BRR accumulates to 17 bits, saturates to 16 bits, and then wraps to 15 bits

			switch filter {
			// last[1] * 15 / 16
			case 1:
				s += int32(last_sample[1])       // add 16/16
				s += -int32(last_sample[1]) >> 4 // add (-1)/16
				// don't clamp - result does not overflow 16 bits
			case 2:
				// last[1] * 61 / 32 - last[0] * 15 / 16
				s += int32(last_sample[1]) << 1                                   // add 64/32
				s += -(int32(last_sample[1]) + (int32(last_sample[1]) << 1)) >> 5 // add (-3)/32
				s += -int32(last_sample[0])                                       // add (-16)/16
				s += int32(last_sample[0]) >> 4                                   // add 1/16
				s = clamp(s, 16)
			case 3:
				// last[1] * 115 / 64 - last[0] * 13 / 16
				s += int32(last_sample[1]) << 1                                                                  // add 128/64
				s += -(int32(last_sample[1]) + (int32(last_sample[1]) << 2) + (int32(last_sample[1]) << 3)) >> 6 // add (-13)/64
				s += -int32(last_sample[0])                                                                      // add (-16)/16
				s += (int32(last_sample[0]) + (int32(last_sample[0]) << 1)) >> 4                                 // add 3/16
				s = clamp(s, 16)
			}

			s = int32(int16(s<<1) >> 1) // wrap to 15 bits, sign-extend to 16 bits
			last_sample[0] = last_sample[1]
			last_sample[1] = int16(s) // 15-bit

			samp_i = (samp_i - 1) & 3   // do this before storing the sample
			sample[samp_i] = int16(s)   // 15-bit
			sample[samp_i+4] = int16(s) // store twice
		} // pitch loop

		var samp []int16 = sample[samp_i : samp_i+4]
		var s int32

		if bc.gauss_enabled {
			var p int32 = pitch >> 4
			var np int32 = -p
			var G4 int16 = gauss_table[-1+np]
			var G3 int16 = gauss_table[255+np]
			var G2 int16 = gauss_table[512+p]
			var G1 int16 = gauss_table[256+p]

			// p is always < 0 and >= -256
			// the first 3 steps wrap using a 15-bit accumulator
			// the last step accumulates to 16-bits then saturates to 15-bits

			s = int32(G4*samp[3]) >> 11
			s += int32(G3*samp[2]) >> 11
			s += int32(G2*samp[1]) >> 11
			s = int32(int16(s<<1) >> 1)
			s += int32(G1*samp[0]) >> 11
			s = clamp(s, 15)

			s = (s * 0x07FF) >> 11 // envx
			s = (s * 0x7F) >> 7    // volume

			pitch += int32(bc.pitch_step_base)
		} else {
			s = int32(samp[3])
			pitch += 0x1000
		}

		s <<= 1
		bc.wav_data = append(bc.wav_data, int16(s))
	}
}

func testGauss(G4, G3, G2 int16, ls []int16) uint8 {
	var s int32
	s = (int32(G4) * int32(ls[0])) >> 11
	s += (int32(G3) * int32(ls[1])) >> 11
	s += (int32(G2) * int32(ls[2])) >> 11
	if s > 0x3FFF || s < -0x4000 {
		return 1
	}
	return 0
}

/*
There are 3 pitch values that can cause sign inversion in the gaussian
filtering by overflowing the 15-bit accumulator if the input samples are
too close to the min/max value.

The sum of the first 3 gauss_table values for each of these 3 pitch values
is 2049 while all other pitch values are 2048 or less.
*/

func testOverflow(ls []int16) uint8 {
	var r uint8

	// p = -256; gauss_table[255, 511, 256]
	r = testGauss(370, 1305, 374, ls)

	// p = -255; gauss_table[254, 510, 257]
	r |= testGauss(366, 1305, 378, ls)

	// p = -247; gauss_table[246, 502, 265]
	r |= testGauss(336, 1303, 410, ls)

	return r
}

func (bc *BrrCodec) encode() {
	bc.reset_progress()
	bc.brr_data = []uint8{}

	if bc.loop_start >= uint32(len(bc.wav_data)) {
		bc.loop_start = 0
		bc.loop_enabled = false
	}

	if bc.loop_enabled {
		var start_align uint32 = uint32(16-(bc.loop_start&15)) & 15
		var loop_size uint32 = uint32(len(bc.wav_data) - int(bc.loop_start))
		var end_align uint32 = loop_size

		for (end_align & 15) != 0 {
			end_align <<= 1
		}

		// remove the existing loop block from the alignment
		end_align -= loop_size

		// also include the loop_start alignment
		end_align += start_align

		if end_align != 0 {
			// TODO: make sure that loop_start + loop_size is the end of wav_data
			//wav_data.resize(wav_data.size() + end_align, 0);

			var src = bc.loop_start
			var dst = bc.loop_start + loop_size
			var end = bc.loop_start + uint32(loop_size) + end_align

			for dst != end {
				bc.wav_data = append(bc.wav_data, bc.wav_data[src])
				//wav_data[dst] = wav_data[src];
				dst++
				src++
			}

			// 16-sample align loop_start
			bc.loop_start += start_align
		}
	} else {
		for (len(bc.wav_data) & 15) != 0 {
			bc.wav_data = append(bc.wav_data, 0)
		}
	}

	const base_adjust_rate float64 = 0.0004
	var adjust_rate float64 = base_adjust_rate
	var loop_block uint32 = bc.loop_start / 16
	var wimax uint32 = uint32(len(bc.wav_data) / 16)
	var wi uint32 = 0
	var best_samp [18]int16

	//best_samp[0] = 0; already inited
	//best_samp[1] = 0;

	bc.total_blocks = wimax
	bc.total_error = 0
	bc.avg_error = 0
	bc.min_error = 1e20
	bc.max_error = 0

	for wi != wimax {
		var p = bc.wav_data[wi*16:]
		var best_err float64 = 1e20
		var blk_samp [18]int16
		var best_data [9]uint8

		blk_samp[0] = best_samp[0]
		blk_samp[1] = best_samp[1]

		for filter := 0; filter <= 3; filter++ {
			// Only use filter 0 for the start block or loop block
			if filter != 0 {
				if (wi == 0) || (wi == loop_block) {
					continue
				}
			}

			// Ranges 0, 13, 14, 15 are "invalid", so they are not used for encoding.
			// The values produced by these ranges are fully covered by the other
			// range values, so there will be no loss in quality.
			for brange := 12; brange >= 1; brange-- {
				var rhalf int32 = (1 << brange) >> 1
				var blk_err float64 = 0
				var blk_data [16]uint8

				for n := 0; n < 16; n++ {
					//int16* blk_ls = blk_samp + n;
					var filter_s int32

					switch filter {
					case 0:
						filter_s = 0

					case 1:
						filter_s = int32(blk_samp[n+1])        // add 16/16
						filter_s += int32(-blk_samp[n+1]) >> 4 // add (-1)/16

					case 2:
						filter_s = int32(blk_samp[n+1]) << 1                                   // add 64/32
						filter_s += -(int32(blk_samp[n+1]) + (int32(blk_samp[n+1]) << 1)) >> 5 // add (-3)/32
						filter_s += -int32(blk_samp[n+0])                                      // add (-16)/16
						filter_s += int32(blk_samp[n+0]) >> 4                                  // add 1/16

					case 3:
						filter_s = int32(blk_samp[n+1]) << 1                                                                 // add 128/64
						filter_s += -(int32(blk_samp[n+1]) + (int32(blk_samp[n+1]) << 2) + (int32(blk_samp[n+1]) << 3)) >> 6 // add (-13)/64
						filter_s += -int32(blk_samp[n+0])                                                                    // add (-16)/16
						filter_s += (int32(blk_samp[n+0]) + (int32(blk_samp[n+0]) << 1)) >> 4                                // add 3/16
					default:
						panic("unknown filter value")
					}

					// undo 15 -> 16 bit conversion
					var xs int32 = int32(p[n]) >> 1

					// undo 16 -> 15 bit wrapping
					// check both possible 16-bit values
					var s1 int32 = int32(int16(xs & 0x7FFF))
					var s2 int32 = int32(int16(xs | 0x8000))

					// undo filtering
					s1 -= filter_s
					s2 -= filter_s

					// restore low bit lost during range decoding
					s1 <<= 1
					s2 <<= 1

					// reduce s to range with nearest value rounding
					// range = 2, rhalf = 2
					// s(-6, -5, -4, -3) = -1
					// s(-2, -1,  0,  1) =  0
					// s( 2,  3,  4,  5) =  1
					s1 = (s1 + rhalf) >> brange
					s2 = (s2 + rhalf) >> brange

					s1 = clamp(s1, 4)
					s2 = clamp(s2, 4)

					var rs1 uint8 = uint8(s1 & 0x0F)
					var rs2 uint8 = uint8(s2 & 0x0F)

					// -16384 to 16383
					s1 = (s1 << brange) >> 1
					s2 = (s2 << brange) >> 1

					// BRR accumulates to 17 bits, saturates to 16 bits, and then wraps to 15 bits

					if filter >= 2 {
						s1 = clamp(s1+filter_s, 16)
						s2 = clamp(s2+filter_s, 16)
					} else {
						// don't clamp - result does not overflow 16 bits
						s1 += filter_s
						s2 += filter_s
					}

					// wrap to 15 bits, sign-extend to 16 bits
					s1 = int32(int16(s1<<1) >> 1)
					s2 = int32(int16(s2<<1) >> 1)

					var d1 float64 = float64(xs - s1)
					var d2 float64 = float64(xs - s2)

					d1 *= d1
					d2 *= d2

					// If d1 == d2, prefer s2 over s1.
					if d1 < d2 {
						blk_err += d1
						blk_samp[n+2] = int16(s1)
						blk_data[n] = rs1
					} else {
						blk_err += d2
						blk_samp[n+2] = int16(s2)
						blk_data[n] = rs2
					}
				} // block loop

				// Use < for comparison. This will cause the encoder to prefer
				// less complex filters and higher ranges when error rates are equal.
				// This will then result in a slightly lower average error rate.
				if blk_err < best_err {
					best_err = blk_err

					for n := 0; n < 16; n++ {
						best_samp[n+2] = blk_samp[n+2]
					}
					best_data[0] = uint8((brange << 4) | (filter << 2))

					for n := 0; n < 8; n++ {
						best_data[n+1] = (blk_data[n*2] << 4) | blk_data[n*2+1]
					}
				}
			} // range loop
		} // filter loop

		var overflow uint16 = 0

		for n := 0; n < 16; n++ {
			var b uint8 = testOverflow(best_samp[n:])
			overflow = (overflow << 1) | uint16(b)
		}

		if overflow != 0 {
			var f [16]float64

			for n := 0; n < 16; n++ {
				f[n] = adjust_rate
			}

			for n := 0; n < 16; n++ {
				if (overflow & 0x8000) != 0 {
					var t float64 = 0.05

					for i := n; i >= 0; i-- {
						f[i] *= 1.0 + t
						t *= 0.1
					}

					t = 0.05 * 0.1
					for i := n + 1; i < 16; i++ {
						f[i] *= 1.0 + t
						t *= 0.1
					}
				}
				overflow <<= 1
			}

			for n := 0; n < 16; n++ {
				p[n] = int16(float64(p[n]) * (1.0 - f[n]))
			}

			adjust_rate *= 1.1
		} else {
			adjust_rate = base_adjust_rate
			best_samp[0] = best_samp[16]
			best_samp[1] = best_samp[17]

			bc.total_error += best_err

			if best_err < bc.min_error {
				bc.min_error = best_err
			}
			if best_err > bc.max_error {
				bc.max_error = best_err
			}
			bc.brr_data = append(bc.brr_data, best_data[:]...)

			wi += 1
			bc.set_progress(uint(wi * 100 / wimax))
		}
	} // wave loop

	if wimax == 0 {
		bc.min_error = 0
	} else {
		bc.avg_error = bc.total_error / float64(wimax)
	}
	if len(bc.brr_data) == 0 || !bc.loop_enabled {
		bc.brr_data = append(bc.brr_data, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	}

	var last_header_set_bits uint8 = 1
	if bc.loop_enabled {
		last_header_set_bits |= 2
	}
	bc.brr_data[len(bc.brr_data)-9] |= last_header_set_bits

	bc.set_progress(100)

	if !bc.user_pitch_enabled {
		// 0.128 = 0x1000 / 32000
		var x uint32 = uint32(float64(bc.input_sample_rate)*0.128 + 0.5)

		if x < 1 {
			x = 1
		} else if x > 0x3FFF {
			x = 0x3FFF
		}

		bc.pitch_step_base = uint16(x)
	}
}

// void FASTCALL BrrCodec::read_brr (Stream& is)
// {
// 	brr_data.resize(((is.size() + 8) / 9) * 9, 0);
// 	is.read(&brr_data[0], is.size());
// }

// func (bc *BrrCodec) readBrr(is io.Reader) error {
// 	bc.brr_data = make([]uint8, ((is.Size()+8)/9)*9)
// 	_, err := is.Read(bc.brr_data[:is.Size()])
// 	return err
// }

func (bc *BrrCodec) readBrr(is io.Reader) {
	var err error
	bc.brr_data, err = io.ReadAll(is)
	if err != nil {
		panic(err)
	}
	for (len(bc.brr_data) % 9) != 0 {
		bc.brr_data = append(bc.brr_data, 0)
	}
}

func (bc *BrrCodec) writeBrr(os io.Writer) {
	_, err := os.Write(bc.brr_data)
	if err != nil {
		panic(err)
	}
}

// Load a wav file into the codec.
func (bc *BrrCodec) readWav(file io.ReadSeeker) {
	decoder := wav.NewDecoder(file)
	//meta := decoder.ReadMetadata()

	bc.wav_data = []int16{}

	for {
		var buffer audio.IntBuffer

		count, err := decoder.PCMBuffer(&buffer)
		if err != nil {
			panic(err)
		}
		if count == 0 {
			break
		}

	}
}
