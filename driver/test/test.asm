;---------------------------------------------------
; it2spc test bench
;---------------------------------------------------

.include "snes.inc"
.include "snesmod.inc"
.include "snes_joypad.inc"
.include "soundbank.inc"

.global main, _nmi

.import gfx_ifontTiles
.import DecompressDataVram

.code

;==============================================================================
SoundTable:
;==============================================================================
SND_TEST = 0
	.byte	4
	.byte	8
	.byte	15
	.word	(TEST66_DATA_END-TEST66_DATA)/9
	.word	.LOWORD(TEST66_DATA)
	.byte	^TEST66_DATA
;------------------------------------------------------------------------------

TEST66_DATA:
.incbin "../sound/tada.brr"
TEST66_DATA_END:


bgText = 1024

.i16
.a8

;---------------------------------------------------
.zeropage
;---------------------------------------------------

colour: .res 1

;---------------------------------------------------
.code
;---------------------------------------------------

main:
	
	lda	#80h
	sta	REG_INIDISP
	
	stz	REG_CGADD	
	stz	REG_CGDATA	; blue
	lda	#80		;
	sta	REG_CGDATA	;
	lda	#0FFh		; white
	sta	REG_CGDATA	;	
	sta	REG_CGDATA	;
	
	ldx	#.LOWORD(gfx_ifontTiles)
	lda	#^gfx_ifontTiles
	ldy	#32*16
	jsr	DecompressDataVram
	
	stz	REG_VMAIN
	
	lda	#1<<2
	sta	REG_BG1SC
	
	stz	REG_BGMODE
	lda	#1
	sta	REG_TM
	
	ldx	#bgText+1+32
	ldy	#test_message
	jsr	DrawString
	
	lda	#81h
	sta	REG_NMITIMEN
	
	wai
	lda	#0Fh
	sta	REG_INIDISP
	
	sei
	jsr	spcBoot
	cli
	
	lda	#^__SOUNDBANK__
	jsr	spcSetBank
	
	ldx	#MOD_TEST
	jsr	spcLoad
	
	ldx	#0
	jsr	spcPlay
	
	lda	#39
	jsr	spcAllocateSoundRegion
	
	jsr	spcFlush	
	
	ldx	#0
	ldy	#4

	lda	#^SoundTable|80h
	ldy	#.LOWORD(SoundTable)
	jsr	spcSetSoundTable
	
	lda	#0
main_loop:

	stz	REG_CGADD
	inc	colour
	lda	colour
	and	#31
	sta	REG_CGDATA
	stz	REG_CGDATA
	
	lda	joy1_down
	bit	#JOYPAD_A
	beq	@nkeypress_a

	;spcPlaySoundM SND_TEST
	ldx	#MOD_POLLEN8
	jsr	spcLoad
	
@nkeypress_a:

	lda	joy1_down+1
	bit	#JOYPADH_B
	beq	@nkeypress_b
	
	ldx	#0
	jsr	spcPlay
@nkeypress_b:

	lda	joy1_down
	bit	#JOYPAD_X
	beq	@nkeypress_x
	
	jsr	spcStop
@nkeypress_x:

;	lda	joy1_down
;	bit	#JOYPAD_Y
;	beq	@nkeypress_y
	
;	jsr	spc
;@nkeypress_y:

	jsr	spcProcess
	wai
	
	jmp	main_loop
	
; x = offset
; y = string
DrawString:
	stx	REG_VMADDL
	
_drawstring:
	lda	0, y
	beq	_eos
	iny
	sta	REG_VMDATAL
	bra	_drawstring
_eos:
	rts
	
; x = offset
; a = byte
DrawByte:
	stx	REG_VMADDL
	
	pha
	lsr
	lsr
	lsr
	lsr
	cmp	#10
	bcs	:+
	adc	#'0'
	bra	_digit2
:	adc	#'A'-10-1
_digit2:
	sta	REG_VMDATAL
	pla
	and	#0Fh
	cmp	#10
	bcs	:+
	adc	#'0'
	bra	_digit3
:	adc	#'A'-10-1
_digit3:
	sta	REG_VMDATAL
	rts
	
.i16
.a8

_nmi:
	rep	#30h
	pha
	phy
	phx
	sep	#20h
	
	ldx	#bgText+1+96
	lda	REG_APUIO0
	jsr	DrawByte
	
	ldx	#bgText+1+99
	lda	REG_APUIO1
	jsr	DrawByte
	
	ldx	#bgText+1+102
	lda	REG_APUIO2
	jsr	DrawByte
	
	ldx	#bgText+1+105
	lda	REG_APUIO3
	jsr	DrawByte
	
	jsr	joyRead
	
	rep	#30h
	plx
	ply
	pla
	rti

test_message:
	.byte "Hello World", 0
	
str_spc_ports:
	.byte "SPC Ports", 0

.segment "HDATA"
.segment "HRAM"
.segment "HRAM2"
.segment "XCODE"
