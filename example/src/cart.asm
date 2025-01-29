;*****************************************************************************************
; SNES Cartridge Header
;
; Customize as needed
; https://snes.nesdev.org/wiki/ROM_header
;*****************************************************************************************

;-----------------------------------------------------------------------------------------
.segment "HEADER"
;-----------------------------------------------------------------------------------------

.import _start
.import _nmi

;-----------------------------------------------------------------------------------------
; Expansion RAM size
;-----------------------------------------------------------------------------------------
XRAM_NONE		=00H
XRAM_16KBIT		=01H
XRAM_64KBIT		=02H
XRAM_256KBIT		=03H
XRAM_512KBIT		=04H
XRAM_1MBIT		=05H

;-----------------------------------------------------------------------------------------
; Memory mapping mode
;-----------------------------------------------------------------------------------------
MODE_20			=20H	; mode 20, 2.68 mhz (LoROM)
MODE_21			=21H	; mode 21, 2.68 mhz (HiROM)
MODE_23			=23H	; mode 23, 2.68 mhz (?)
MODE_25			=25H	; mode 25, 2.68 mhz (?)
MODE_20_FAST		=30H	; mode 20, 3.58 mhz (LoROM, FastROM)
MODE_21_FAST		=31H	; mode 21, 3.58 mhz (HiROM, FastROM)
MODE_25_FAST		=35H	; mode 25, 3.58 mhz (?)

;-----------------------------------------------------------------------------------------
; Cartridge type
;-----------------------------------------------------------------------------------------
CART_ROM		=00H	; ROM Only
CART_ROM_RAM		=01H	; ROM+RAM
CART_ROM_RAM_BATT	=02H	; ROM+RAM+BATTERY

;-----------------------------------------------------------------------------------------
; Destination code
;-----------------------------------------------------------------------------------------
DEST_JAPAN		=00H
DEST_USA_CANADA		=01H
DEST_EUROPE		=02H
DEST_SCANDANAVIA	=03H
DEST_FRENCH_EUROPE	=06H
DEST_DUTCH		=07H
DEST_SPANISH		=08H
DEST_GERMAN		=09H
DEST_ITALIAN		=0AH
DEST_CHINESE		=0BH
DEST_KOREAN		=0DH
DEST_COMMON		=0EH
DEST_CANADA		=0FH
DEST_BRAZIL		=10H
DEST_AUSTRALIA		=11H
DEST_OTHERX		=12H
DEST_OTHERY		=13H
DEST_OTHERZ		=14H

;-----------------------------------------------------------------------------------------
; Expanded Cartridge Header
;-----------------------------------------------------------------------------------------
.byte	'0', '0'		; B0 - Maker code
.byte	"BREW"			; B2 - Game code
.byte	0, 0, 0, 0, 0, 0, 0	; B6 - Reserved
.byte	XRAM_NONE		; BD - Expansion ram size
.byte	00h			; BE - Special version
.byte	00h			; BF - Cartridge sub-number

;-----------------------------------------------------------------------------------------
; Standard Cartridge Header
;-----------------------------------------------------------------------------------------
.byte	"SNESMOD DEMO         "	; C0 - Game title
.byte	MODE_21		       	; D5 - Map mode (Update if using HiROM or other)
.byte	CART_ROM		; D6 - Cartridge type (ROM only)
.byte	00h			; D7 - ROM size (set with sneschk later)
.byte	XRAM_NONE		; D8 - RAM size
.byte	DEST_USA_CANADA		; D9 - Destination code (USA/Canada) (NTSC)
.byte	33h			; DA - Fixed byte
.byte	00h			; DB - Mask ROM version
.word	0FFFFh			; DC - Complement check (set with sneschk later)
.word	0000h			; DE - Checksum (set with sneschk later)

;-----------------------------------------------------------------------------------------
; Native mode vectors
;-----------------------------------------------------------------------------------------
VEC_E0:	.word	0		; -
VEC_E2:	.word	0		; -
VEC_E4:	.word	0		; -
VEC_E6:	.word	0		; BRK
VEC_E8:	.word	0		; ABORT
VEC_EA:	.word	_nmi		; NMI
VEC_EC:	.word	0		; RESET
VEC_EE:	.word	0		; IRQ

;-----------------------------------------------------------------------------------------
; Emulation mode vectors
;-----------------------------------------------------------------------------------------
VEC_F0:	.word	0		; -
VEC_F2:	.word	0		; -
VEC_F4:	.word	0		; COP
VEC_F6:	.word	0		; -
VEC_F8:	.word	0		; ABORT
VEC_FA:	.word	0		; NMI
VEC_FC:	.word	_start		; RES (entry point)
VEC_FE:	.word	0		; IRQ/BRK
;-----------------------------------------------------------------------------------------
