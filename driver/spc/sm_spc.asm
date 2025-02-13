;*
;* Copyright 2009 Mukunda Johnson (mukunda.com)
;* 
;* This file is part of SNESMOD - gh.mukunda.com/snesmod
;*
;* See LICENSING.txt
;*

#define DEBUGINC inc debug \ mov SPC_PORT0, debug

.define LBYTE(z) (z & 0FFh)
.define HBYTE(z) (z >> 8)

.define SPROC TCALL 0
.define SPROC2 SPROC

;********************************************************
; PROTOCOL
;
; mm = mimic data
; id = message id
; vv = validation data (not previous value)
; v1 = nonzero validation data (not previous value)
;
; SPC PORTS:
; PORT0 = RESERVED
; PORT1 = COMMUNICATION
; PORT2 = STATUS:
;   MSB fep-cccc LSB
;   f = module volume fade[out/in] in progress
;   e = end of module reached (restarted from beginning)
;   p = module is playing (0 means not playing or preparing...)
;   cccc = cue, incremented on SF1 pattern effect
; PORT3 = MODULE POSITION
; 
; NAME	ID	DESC
;--------------------------------------------------------
; LOAD	00	Upload Module
; 
; >> id vv -- --	send message
; << -- mm -- --	message confirmed
;
; >> -- v1 DD DD	transfer module
; << -- mm -- --	DDDD = data, loop until all words xferred
;
; >> -- 00 DD DD	final word
; << -- mm -- --	okay proceed to transfer sources...
;
; for each entry in SOURCE_LIST:
;
; >> 01 vv LL LL	send loop point
; << -- mm -- --	loop point saved
; >> -- v1 DD DD	transfer source data
; << -- mm -- --	DDDD = data, loop unti all words xferred
;
; >> -- 00 DD DD	transfer last word
; << -- mm -- --	
;
; [loop until all needed sources are transferred]
;
; >> 00 vv -- --	terminate transfer
; << -- mm -- --
;
; notes:
;   this function resets the memory system
;   all sound effects will become invalid
; 
;   after final sample transferred the system may
;   be halted for some time to setup the echo delay.
;--------------------------------------------------------
; LOADE	01	Upload Sound Effect
;
; >> id vv LL LL	send message
; << -- mm -- --	source registered, ready for data
;
; >> -- v1 DD DD	transfer source data
; << -- mm -- --	loop until all words xferred
;
; >> -- 00 DD DD	send last word
; << -- mm -- --	okay, ready for playback
;
; sound effects are always one-shot
;  LLLL is not used (or maybe it is...........)
;--------------------------------------------------------
; VOL	02	Set Master Volume
;
; >> id vv VV --
; << -- mm -- --
;
; VV = master volume level (0..127)
;--------------------------------------------------------
; PLAY	03	Play Module
;
; >> id vv -- pp
; << -- mm -- --
;
; pp = start position
;--------------------------------------------------------
; STOP	04	Stop Playback
;
; >> id vv -- --
; << -- mm -- --
;--------------------------------------------------------
; MVOL	05	Set Module Volume
;
; >> id vv -- VV
; << -- mm -- --
;
; VV = 0..255 new module volume scale
;--------------------------------------------------------
; FADE	06	Fade Module Volume
;
; >> id vv tt VV
; << -- mm -- --
;
; VV = 0..255 target volume level
; tt = fade speed (added every m tick)
;--------------------------------------------------------
; RES	07	Reset
;
; >> id vv -- --
; 
; <driver unloaded>
;--------------------------------------------------------
; FX	08	Play Sound Effect
;
; >> id vv vp sh
; << -- mm -- --
; 
; s = sample index
; h = pitch ( 8 = 32000hz, h = pitch height >> 9 )
; v = volume (15 = max)
; p = panning (8 = center)
;--------------------------------------------------------
; TEST	09	Test function
;
; >> id vv -- --
; << -- mm -- --
;--------------------------------------------------------
; SSIZE	0A	Set sound region size
;
; >> id vv -- SS
; << -- mm -- --
;
; SS = size of sound region (SS*256 bytes)
;--------------------------------------------------------
; STREAM	Update digital stream
;
; previously written port data must be buffered.
;
; >> 8m -- -- --	send update flag (8m = previous data OR 80H)
; [wait for spc, this is a high-priority signal]
; << 80 -- -- --	receive ready signal
;
; >> nn mm vp hh	nn = number of blocks (9 bytes) to transfer (1..28)
; << nn -- -- --
;
; if mm <> 0 then
;   [new sample, reset sound]
;   v = volume
;   p = panning
;   hh = pitch height H byte (6 bits)
;
; length should be significantly larger than required on
; initial transfer (mm<>0)
;
; [xx is a counter starting with 'nn' written to port0 earlier]
; [add 1 before first message]
;
; transfer 1 chunk:
;  loop 3 times:
;   >> xx D2 D0 D1
;   << xx -- -- --
;   >> xx D5 D3 D4
;   << xx -- -- --
;   >> xx D8 D6 D7
;   << xx -- -- --
; loop nn times
;
;(EXIT):
; [spc will resume operation after a short period]
; [port status must be restored before the spc resumes (approx. 45us)]
; >> pp pp pp pp	restore port status
;********************************************************

;*****************************************************************************************
; registers
;*****************************************************************************************
SPC_TEST	=0F0h ; undocumented
SPC_CONTROL	=0F1h ; control register
SPC_DSP		=0F2h
SPC_DSPA	=0F2h
SPC_DSPD	=0F3h
SPC_PORT0	=0F4h ; i/o port0
SPC_PORT1	=0F5h ; i/o port1
SPC_PORT2	=0F6h ; i/o port2
SPC_PORT3	=0F7h ; i/o port3
SPC_FLAGS	=0F8h ; custom flags
SPC_TIMER0	=0FAh ; timer0 setting
SPC_TIMER1	=0FBh ; timer1 setting
SPC_TIMER2	=0FCh ; timer2 setting
SPC_COUNTER0	=0FDh ; timer0 counter
SPC_COUNTER1	=0FEh ; timer1 counter
SPC_COUNTER2	=0FFh ; timer2 counter

DEBUG_P0 = SPC_PORT0
DEBUG_P2 = SPC_PORT2

;*****************************************************************************************
; dsp registers
;*****************************************************************************************
DSPV_VOL	=00h
DSPV_VOLR	=01h
DSPV_PL		=02h
DSPV_PH		=03h
DSPV_SRCN	=04h
DSPV_ADSR1	=05h
DSPV_ADSR2	=06h
DSPV_GAIN	=07h
DSPV_ENVX	=08h
DSPV_OUTX	=09h

DSP_MVOL	=0Ch
DSP_MVOLR	=1Ch
DSP_EVOL	=2Ch
DSP_EVOLR	=3Ch
DSP_KON		=4Ch
DSP_KOF		=5Ch
DSP_FLG		=6Ch
DSP_ENDX	=7Ch

DSP_EFB		=0Dh
DSP_PMON	=2Dh
DSP_NON		=3Dh
DSP_EON		=4Dh
DSP_DIR		=5Dh
DSP_ESA		=6Dh
DSP_EDL		=7Dh

DSP_C0		=0Fh
DSP_C1		=1Fh
DSP_C2		=2Fh
DSP_C3		=3Fh
DSP_C4		=4Fh
DSP_C5		=5Fh
DSP_C6		=6Fh
DSP_C7		=7Fh

FLG_RESET	=80h
FLG_MUTE	=40h
FLG_ECEN	=20h

#define SETDSP(xx,yy) mov SPC_DSPA, #xx\ mov SPC_DSPD, #yy

;*****************************************************************************************
; module defs
;*****************************************************************************************

MOD_IV		=00H	; INITIAL VOLUME
MOD_IT		=01H	; INITIAL TEMPO
MOD_IS		=02H	; INITIAL SPEED
MOD_CV		=03H	; INITIAL CHANNEL VOLUME
MOD_CP		=0BH	; INITIAL CHANNEL PANNING
MOD_EVOL	=13H	; ECHO VOLUME (LEFT)
MOD_EVOLR	=14H	; ECHO VOLUME (RIGHT)
MOD_EDL		=15H	; ECHO DELAY
MOD_EFB		=16H	; ECHO FEEDBACK
MOD_EFIR	=17H	; ECHO FIR COEFS
MOD_EON		=1FH	; ECHO ENABLE BITS
MOD_SEQU	=20H	; SEQUENCE
MOD_PTABLE_L	=0E8H	; PATTERN TABLE
MOD_PTABLE_H	=128H	; 
MOD_ITABLE_L	=168H	; INSTRUMENT TABLE
MOD_ITABLE_H	=1A8H	; 
MOD_STABLE_L	=1E8H	; SAMPLE TABLE
MOD_STABLE_H	=228H	;

INS_FADEOUT	=00H
INS_SAMPLE	=01H
INS_GVOL	=02H
INS_SETPAN	=03H
INS_ENVLEN	=04H
INS_ENVSUS	=05H
INS_ENVLOOPST	=06H
INS_ENVLOOPEND	=07H
INS_ENVDATA	=08H

SAMP_DVOL	=00H
SAMP_GVOL	=01H
SAMP_PITCHBASE	=02H
SAMP_DINDEX	=04H
SAMP_SETPAN	=05H

;*****************************************************************************************
; zero-page memory
;*****************************************************************************************

xfer_address:	.block 2
m0:		.block 2
m1:		.block 2
m2:		.block 2
m3:		.block 2
m4:		.block 2
m5:		.block 2
m6:		.block 2
next_sample:	.block 1
comms_v:	.block 1 ;communication variable

evol_l:		.block 1
evol_r:		.block 1

module_vol:	.block 1 ;module volume
module_fadeT:	.block 1 ;module volume fade target
module_fadeR:	.block 1 ;module volume fade rate
module_fadeC:	.block 1 ;timer counter

mod_tick:	.block 1
mod_row:	.block 1
mod_position:	.block 1
mod_bpm:	.block 1
mod_speed:	.block 1
mod_active:	.block 1
mod_gvol:	.block 1

patt_addr:	.block 2
patt_rows:	.block 1
pattjump_enable: .block 1
pattjump_index:	.block 1
patt_update:	.block 1 ;PATTERN UPDATE FLAGS

ch_start:
ch_pitch_l:	.block 8
ch_pitch_h:	.block 8
ch_volume:	.block 8 ;0..64
ch_cvolume:	.block 8 ;0..128 (IT = 0..64)
ch_panning:	.block 8 ;0..64
ch_cmem:	.block 8
ch_note:	.block 8
ch_instr:	.block 8
ch_vcmd:	.block 8
ch_command:	.block 8
ch_param:	.block 8
ch_sample:	.block 8
ch_flags:	.block 8
ch_env_y_l:	.block 8
ch_env_y_h:	.block 8
ch_env_node:	.block 8
ch_env_tick:	.block 8
ch_fadeout:	.block 8
ch_end:

; channel processing variables:
t_hasdata:	.block 1
t_sampoff:	.block 1
t_volume:	.block 1
t_panning:	.block 1
t_pitch:
t_pitch_l:	.block 1
t_pitch_h:	.block 1
t_flags:	.block 1
t_env:		.block 1 ; 0..255

p_instr:	.block 2

STATUS:		.block 1
STATUS_P	=32
STATUS_E	=64
STATUS_F	=128

debug:		.block 1

CF_NOTE		=1
CF_INSTR	=2
CF_VCMD		=4
CF_CMD		=8
CF_KEYON	=16
CF_FADE		=32
CF_SURROUND	=64

TF_START	=80H
TF_DELAY	=2


;---------------------------
; sound effects
;---------------------------

sfx_mask:	.block 1
sfx_next:	.block 1

;-----------------------------------------------------------------------------------------

stream_a:		.block 1
stream_write:		.block 2
stream_rate:		.block 1
stream_volL:		.block 1
stream_volR:		.block 1
stream_gain:		.block 1
stream_initial:		.block 1
stream_size:		.block 1
stream_region:		.block 1

;*****************************************************************************************
; sample directory
;*****************************************************************************************

SampleDirectory		=0200h	; 256 bytes	(64-sample directory)
EffectDirectory		=0300h	; 16*4 bytes	(16 sound effects)
StreamAddress		=0340h  ; 4 bytes       (streaming buffer address)
PatternMemory		=0380h	; 16*8 bytes

; [extra ram]

;*****************************************************************************************
; program (load @ 400h)
;*****************************************************************************************

;--------------------------------------------------------
.org 400h
;--------------------------------------------------------
	
;--------------------------------------------------------
main:
;--------------------------------------------------------

	mov	x, #0
	mov	a, #0
_clrmem:
	mov	(X)+, a
	cmp	x, #0F0h
	bne	_clrmem
	
	mov	SPC_PORT1, #0		; reset some ports
	mov	SPC_PORT2, #0		;
	mov	SPC_PORT3, #0		;
	mov	SPC_CONTROL, #0		; reset control
	mov	SPC_TIMER1, #255	; reset fade timer
	mov	module_vol, #255	; reset mvol
	mov	module_fadeT, #255	; 
					;----------------
	call	ResetSound		;
					;----------------
	mov	SPC_DSPA, #DSP_MVOL	; reset main volume
	mov	SPC_DSPD, #80		;
	mov	SPC_DSPA, #DSP_MVOLR	;
	mov	SPC_DSPD, #80		;
					;----------------
	mov	SPC_DSPA, #DSP_DIR	; set source dir
	mov	SPC_DSPD, #SampleDirectory >> 8
	
	call	ResetMemory
	
	call	Streaming_Init
	mov	SPC_CONTROL, #%110
	
;----------------------------------------------------------------------
spc_patch_start:
	bra	spc_patch_end		; When creating SPC files, nop this branch out
					; to start the song during boot.
	call	Module_Stop		;
	mov	a, #0			;
	call	Module_Start		;
spc_patch_end:
;----------------------------------------------------------------------

;--------------------------------------------------------
main_loop:
;--------------------------------------------------------

	SPROC2
	call	ProcessComms
	SPROC
	call	ProcessFade
	SPROC
	call	Module_Update
	SPROC
	call	UpdatePorts
	SPROC
	call	SFX_Update
	bra	main_loop
	
;--------------------------------------------------------
UpdatePorts:
;--------------------------------------------------------
	mov	SPC_PORT2, STATUS
	mov	SPC_PORT3, mod_position
	ret
	
;--------------------------------------------------------
ResetMemory:
;--------------------------------------------------------
	mov	xfer_address, #MODULE & 0FFh	; reset transfer address
	mov	xfer_address+1, #MODULE >> 8	;
	mov	next_sample, #0		; reset sample target
	ret
	
;--------------------------------------------------------
ResetSound:
;--------------------------------------------------------
	SETDSP( DSP_KOF, 0FFh );
	SETDSP( DSP_FLG, FLG_ECEN );
	SETDSP( DSP_PMON, 0 );
	SETDSP( DSP_EVOL, 0 );
	SETDSP( DSP_EVOLR, 0 );
	SETDSP( DSP_NON, 00h );
	SETDSP( DSP_KOF, 000h ); this is weird
	
	mov	sfx_mask, #0
	ret
	
;--------------------------------------------------------
ProcessComms:
;--------------------------------------------------------
	cmp	comms_v, SPC_PORT1	; test for command
	bne	_new_message		;
	ret				; <no message>
_new_message:
	mov	comms_v, SPC_PORT1	; copy V
	mov	a, SPC_PORT0		; jump to message
	nop				; verify data
	cmp	a, SPC_PORT0		;
	bne	_new_message		;
	and	a, #127			; mask 7 bits
	asl	a			;
	mov	x, a			;
	jmp	[CommandTable+x]	;'
;--------------------------------------------------------

CommandRet:
	mov	SPC_PORT1, comms_v
	ret

;--------------------------------------------------------
CommandTable:
;--------------------------------------------------------
	.word	CMD_LOAD		; 00h - load module
	.word	CMD_LOADE		; 01h - load sound
	.word	CMD_VOL			; 02h - set volume
	.word	CMD_PLAY		; 03h - play
	.word	CMD_STOP		; 04h - stop
	.word	CMD_MVOL		; 05h - set module volume
	.word	CMD_FADE		; 06h - fade module volume
	.word	CMD_RES			; 07h - reset
	.word	CMD_FX			; 08h - sound effect
	.word	CMD_TEST		; 09h - test
	.word	CMD_SSIZE		; 0Ah - set stream size
	;	.word	CMD_PDS			; 0Ah - play streamed sound
;	.word	CMD_DDS			; 0Bh - disable digital stream
	
;********************************************************
CMD_LOAD:
;********************************************************
	call	Module_Stop
	call	ResetMemory		; reset memory system
	
	call	StartTransfer
	
	mov	m1, #0
	
_wait_for_sourcen:			;
	cmp	comms_v, SPC_PORT1	;
	beq	_wait_for_sourcen	;
	mov	comms_v, SPC_PORT1	;
	
	cmp	SPC_PORT0, #0		; if p0 != 0:
	beq	_end_of_sources		; load source
					;
	mov	y, m1			;
	clrc				;
	adc	m1, #4			;
	call	RegisterSource		;
	call	StartTransfer		;
					;
	bra	_wait_for_sourcen	; load next source
	
_end_of_sources:			; if p0 == 0:
	jmp	CommandRet		;

;-------------------------------------------------------------------
RegisterSource:
;-------------------------------------------------------------------
	mov	a, xfer_address
	mov	!SampleDirectory+y, a
	clrc
	adc	a, SPC_PORT2
	mov	!SampleDirectory+2+y, a
	
	mov	a, xfer_address+1
	mov	!SampleDirectory+1+y, a
	
	adc	a, SPC_PORT3
	mov	!SampleDirectory+3+y, a
	
	ret
	
;-------------------------------------------------------------------
StartTransfer:
;-------------------------------------------------------------------
	mov	x, comms_v		; start transfer
	mov	y, #0			;
	mov	SPC_PORT1, x		;
	
;-------------------------------------------------------------------
DoTransfer:
;-------------------------------------------------------------------
	cmp	x, SPC_PORT1		; wait for data
	beq	DoTransfer		;
	mov	x, SPC_PORT1		;
					;---------------------------
	mov	a, SPC_PORT2		; copy data
	mov	[xfer_address]+y, a	;
	mov	a, SPC_PORT3		;
	mov	SPC_PORT1, x		;<- reply to snes
	inc	y			;
	mov	[xfer_address]+y, a	;
	inc	y			;
	beq	_inc_address		; catch index overflow
_cont1:	cmp	x, #0			; loop until x=0
	bne	DoTransfer		;
	
	mov	m0, y
	clrc
	adc	xfer_address, m0
	adc	xfer_address+1, #0
	mov	comms_v, x
	ret

_inc_address:
	inc	xfer_address+1
	bra	_cont1
	
;********************************************************
CMD_LOADE:
;********************************************************
	mov	a, xfer_address
	mov	y, next_sample
	mov	!EffectDirectory+y, a
	clrc
	adc	a, SPC_PORT2
	mov	!EffectDirectory+2+y, a
	
	mov	a, xfer_address+1
	mov	!EffectDirectory+1+y, a
	
	adc	a, SPC_PORT3
	mov	!EffectDirectory+3+y, a
	
	clrc				; safety clear for invalid loop points (thanks KungFuFurby)
	adc	next_sample, #4
	call	StartTransfer
	
	jmp	CommandRet
	
;********************************************************
CMD_VOL:
;********************************************************
	mov	a, SPC_PORT2
	mov	SPC_DSPA, #DSP_MVOL
	mov	SPC_DSPD, a
	mov	SPC_DSPA, #DSP_MVOLR
	mov	SPC_DSPD, a
	call	UpdateEchoVolume
	jmp	CommandRet
	
;********************************************************
CMD_PLAY:
;********************************************************
	call	Module_Stop
	mov	a, SPC_PORT3
	and	STATUS, #~STATUS_P
	mov	SPC_PORT2, STATUS
	mov	SPC_PORT1, comms_v
	jmp	Module_Start
	
;********************************************************
CMD_STOP:
;********************************************************
	call	Module_Stop
	jmp	CommandRet
	
;********************************************************
CMD_MVOL:
;********************************************************
	mov	module_vol, SPC_PORT3
	mov	module_fadeT, SPC_PORT3
	jmp	CommandRet

;********************************************************
CMD_FADE:
;********************************************************
	or	STATUS, #STATUS_F
	mov	SPC_PORT2, STATUS
	mov	module_fadeT, SPC_PORT3
	mov	module_fadeR, SPC_PORT2
	jmp	CommandRet
	
;********************************************************
CMD_RES:
;********************************************************
	mov	SPC_DSPA, #DSP_FLG
	mov	SPC_DSPD, #11100000b
	clrp
	mov	SPC_CONTROL, #10000000b ;
	jmp	0FFC0h
	
;********************************************************
CMD_FX:
;********************************************************
	movw	ya, SPC_PORT2
	movw	m0, ya
	mov	SPC_PORT1, comms_v
	jmp	SFX_Play

;********************************************************
CMD_TEST:
;********************************************************
	SETDSP( 00h, 7fh )
	SETDSP( 01h, 7fh )
	SETDSP( 02h, 00h )
	SETDSP( 03h, 10h )
	SETDSP( 04h, 09h )
	SETDSP( 05h, 00h )
	SETDSP( 06h, 00h )
	SETDSP( 07h, 7fh )
	SETDSP( 0Ch, 70h )
	SETDSP( 1Ch, 70h )
	SETDSP( 4Ch, 01h )
	jmp	CommandRet
	
;********************************************************
CMD_SSIZE:
;********************************************************
	call	Module_Stop
	mov	a, SPC_PORT3
	call	Streaming_Resize
	jmp	CommandRet

;********************************************************
CMD_DDS:
;********************************************************
;	call	Streaming_Deactivate
;	jmp	CommandRet
	

;********************************************************
; Setup echo...
;********************************************************
SetupEcho:
	SETDSP( DSP_FLG, 00100000b );
	SETDSP( DSP_EVOL, 0 );
	SETDSP( DSP_EVOLR, 0 );
	
	mov	a, !MODULE+MOD_EVOL
	mov	evol_l, a
	mov	a, !MODULE+MOD_EVOLR
	mov	evol_r, a
	
	mov	a, !MODULE+MOD_EDL	; ESA = stream_region - EDL*8
	xcn	a			; max = stream_region -1
	lsr	a			;
	mov	m0, a			;
	mov	a, stream_region	;
	setc				;
	sbc	a, m0			;
	cmp	a, stream_region	;
	bne	_edl_not_ss		;
	dec	a			;
_edl_not_ss:				;
	mov	SPC_DSPA, #DSP_ESA	;
	mov	SPC_DSPD, a		;
	
	mov	m0+1, a			; clear memory region used by echo
	mov	m0, #0			;
	mov	a, #0			;
	mov	y, #0			;
_clearmem:				;
	mov	[m0]+y, a		;
	inc	y			;
	bne	_clearmem		;
	inc	m0+1			;
	cmp	m0+1, stream_region	;
	bne	_clearmem		;
	
	setc				; copy FIR coefficients
	mov	SPC_DSPA, #DSP_C7	;
	mov	y, #7			;
_copy_coef:				;
	mov	a, !MODULE+MOD_EFIR+y	;
	mov	SPC_DSPD, a		;
	sbc	SPC_DSPA, #10h		;
	dec	y			;
	bpl	_copy_coef		;
	
	mov	SPC_DSPA, #DSP_EFB	; copy EFB
	mov	a, !MODULE+MOD_EFB	;
	mov	SPC_DSPD, a		;
	
	mov	SPC_DSPA, #DSP_EON	; copy EON
	mov	a, !MODULE+MOD_EON	;
	mov	SPC_DSPD, a		;
	
	mov	SPC_DSPA, #DSP_EDL	; read old EDL, set new EDL
	mov	y, SPC_DSPD		;
	mov	a, !MODULE+MOD_EDL		;
	mov	SPC_DSPD, a		;
	
	;-----------------------------------------
	; delay EDL*16ms before enabling echo
	; 16384 clks * EDL
	; EDL<<14 clks
	;
	; run loop EDL<<10 times
	;-----------------------------------------
	mov	a, y			;
	asl	a			;
	asl	a			;
	inc	a
	mov	m0+1, a			;
	mov	m0, #0			;
_delay_16clks:				;
	cmp	a, [0]+y		;
	decw	m0			;
	bne	_delay_16clks		;
	
	
	
	mov	a, !MODULE+MOD_EDL
	beq	_skip_enable_echo

	call	UpdateEchoVolume
	mov	SPC_DSPA, #DSP_FLG	; clear ECEN
	mov	SPC_DSPD, #0
	ret
_skip_enable_echo:

	mov	evol_l, #0
	mov	evol_r, #0
	ret
	
;********************************************************
; set echo volume with master scale applied
;********************************************************
UpdateEchoVolume:
	
	mov	SPC_DSPA, #DSP_MVOL	; set EVOL scaled by main volume
	mov	a, SPC_DSPD		;
	asl	a			;
	mov	m0, a			;
	mov	SPC_DSPA, #DSP_EVOL	;
	mov	y, evol_l		;
	mul	ya			;
	mov	a, y			;
	mov	y, evol_l		;
	bpl	_plus			;
	setc				;
	sbc	a, m0			;
_plus:	mov	SPC_DSPD, a		;

	mov	a, m0			; set EVOLR scaled by main volume
	mov	SPC_DSPA, #DSP_EVOLR	;
	mov	y, evol_r		;
	mul	ya			;
	mov	a, y			;
	mov	y, evol_r		;
	bpl	_plusr			;
	setc				;
	sbc	a, m0			;
_plusr:	mov	SPC_DSPD, a		;
	
	ret
	
;********************************************************
; zerofill channel data
;********************************************************
Module_ResetChannels:
	mov	x, #ch_start
	mov	a, #0
_zerofill_ch:
	mov	(x)+, a
	cmp	x, #ch_end
	bne	_zerofill_ch
	ret
	
Module_Stop:
	call	ResetSound
	mov	SPC_CONTROL, #%110
	mov	mod_active, #0
	ret
	
;********************************************************
; play module...
;
; a = initial position
;********************************************************
Module_Start:
	mov	mod_position, a
	call	ResetSound
	call	Module_ResetChannels
	mov	mod_active, #1
	mov	a, !MODULE+MOD_IS
	mov	mod_speed, a
	mov	a, !MODULE+MOD_IT
	call	Module_ChangeTempo
	mov	a, !MODULE+MOD_IV
	mov	mod_gvol, a

	mov	x, #7				;
_copy_cvolume:					; copy volume levels
	mov	a, !MODULE+MOD_CV+x		;
	mov	ch_cvolume+x, a			;
	dec	x				;
	bpl	_copy_cvolume			;
	
	mov	x, #7
_copy_cpan:
	mov	a, !MODULE+MOD_CP+x
	cmp	a, #65
	bcs	_cpan_surround
	mov	ch_panning+x, a
	bra	_cpan_normal
_cpan_surround:
	mov	a, #32
	mov	ch_panning+x, a
	mov	a, #CF_SURROUND
	mov	ch_flags+x, a
_cpan_normal:
	dec	x
	bpl	_copy_cpan
	
	call	SetupEcho
	
	mov	a, mod_position
	call	Module_ChangePosition
	
	; start timer
	mov	SPC_CONTROL, #%111
	
	or	STATUS, #STATUS_P
	mov	SPC_PORT2, STATUS
	
	SETDSP( DSP_KOF, 0 );	// ??????
	ret

;********************************************************
; set sequence position
;
; a=position
;********************************************************
Module_ChangePosition:
	
	mov	y, a
_skip_pattern:
	mov	a, !MODULE+MOD_SEQU+y
	cmp	a, #254			; skip +++
	bne	_not_plusplusplus	;
	inc	y			;
	bra	_skip_pattern		;
_not_plusplusplus:
	cmp	a, #255			; restart on ---
	bne	_not_end		;
	mov	y, #0			;
	bra	_skip_pattern		;
_not_end:
	mov	mod_position, y
	mov	y, a
	mov	a, !MODULE+MOD_PTABLE_L+y
	mov	patt_addr, a
	mov	a, !MODULE+MOD_PTABLE_H+y
	mov	patt_addr+1, a
	mov	y, #0
	mov	a, [patt_addr]+y
	mov	patt_rows, a
	
	incw	patt_addr
	
	mov	pattjump_enable, #0
	mov	mod_tick, #0
	mov	mod_row, #0
	ret
	
;********************************************************
; a = new BPM value
;********************************************************
Module_ChangeTempo:
	push	x
	mov	mod_bpm, a
	mov	SPC_CONTROL, #%110
	
	mov	x, a
	mov	y, #50h
	mov	a, #00h
	div	ya, x
	mov	SPC_TIMER0, a
	pop	x
	ret
	
;********************************************************
; process module fading
;********************************************************
ProcessFade:
	mov	a, SPC_COUNTER1
	beq	_skipfade
	or	STATUS, #STATUS_F
	mov	a, module_vol
	cmp	a, module_fadeT
	beq	_nofade
	bcc	_fadein
;--------------------------------------------
_fadeout:
;--------------------------------------------
	sbc	a, module_fadeR
	bcs	_fade_satL
	mov	module_vol, module_fadeT
	ret
_fade_satL:
	cmp	a, module_fadeT
	bcs	_fadeset
	mov	module_vol, module_fadeT
	ret
;--------------------------------------------
_fadein:
;--------------------------------------------
	adc	a, module_fadeR
	bcc	_fade_satH
	mov	module_vol, module_fadeT
	ret
_fade_satH:
	cmp	a, module_fadeT
	bcc	_fadeset
	mov	module_vol, module_fadeT
	ret
_fadeset:
	mov	module_vol, a
	ret
_nofade:
	and	STATUS, #~STATUS_F
_skipfade:
	ret

;********************************************************
; Update module playback
;
;********************************************************
Module_Update:
	mov	a, mod_active
	beq	_no_tick
	mov	a, SPC_COUNTER0		; check for a tick
	beq	_no_tick		;

	call	Module_OnTick		;
_no_tick:				;
	ret				;

;********************************************************
; module tick!!!
;********************************************************
Module_OnTick:
	cmp	mod_tick, #0
	bne	_skip_read_pattern
	call	Module_ReadPattern
_skip_read_pattern:

	call	Module_UpdateChannels

	inc	mod_tick		; increment tick until >= SPEED
	cmp	mod_tick, mod_speed	;
	bcc	_exit_tick		;
	mov	mod_tick, #0		;
	
	cmp	pattjump_enable, #0	; catch pattern jump...
	beq	_no_pattjump		;
	mov	a, pattjump_index	;
	jmp	Module_ChangePosition	;
_no_pattjump:				;
	
	inc	mod_row			; increment row until > PATTERN_ROWS
	beq	_adv_pos
	cmp	mod_row, patt_rows	;
	beq	_exit_tick
	bcc	_exit_tick		;
_adv_pos:
	
	mov	a, mod_position		; advance position
	inc	a			;
	jmp	Module_ChangePosition	;
_exit_tick:
	ret

;********************************************************
; read pattern data
;********************************************************
Module_ReadPattern:
	
	mov	y, #1			; skip hints
	mov	a, [patt_addr]+y	; copy update flags
	inc	y			;
	mov	patt_update, a		;
	mov	m1, a			;
	mov	x, #0
	
	lsr	m1			; test first bit
	bcc	_no_channel_data	;
_read_pattern_data:
	SPROC
	mov	a, [patt_addr]+y	; read maskvar
	inc	y			;
	mov	m0, a			;
	
	bbc4	m0, _skip_read_note	; test/read new note
	mov	a, [patt_addr]+y	;
	inc	y			;
	mov	ch_note+x, a		;
_skip_read_note:			;

	bbc5	m0, _skip_read_instr	; test/read new instrument
	mov	a, [patt_addr]+y	;
	inc	y			;
	mov	ch_instr+x, a		;
_skip_read_instr:			;

	bbc6	m0, _skip_read_vcmd	; test/read new vcmd
	mov	a, [patt_addr]+y	;
	inc	y			;
	mov	ch_vcmd+x, a		;
_skip_read_vcmd:			;

	bbc7	m0, _skip_read_cmd	; test/read new cmd+param
	mov	a, [patt_addr]+y	;
	inc	y			;
	mov	ch_command+x, a		;
	mov	a, [patt_addr]+y	;
	inc	y			;
	mov	ch_param+x, a		;
_skip_read_cmd:				;

	and	m0, #0Fh		; set flags (lower nibble)
	mov	a, ch_flags+x		;
	and	a, #0F0h		;
	or	a, m0			;
	mov	ch_flags+x, a		;
	
_no_channel_data:			;
_rp_nextchannel:
	inc	x			; increment index
	lsr	m1			; shift out next bit
	bcs	_read_pattern_data	; process if set
	bne	_no_channel_data	; loop if bits remain (upto 8 iterations)
	;-------------------------------

	mov	m0, y			; add offset to pattern address
	clrc				;
	adc	patt_addr, m0		;
	adc	patt_addr+1, #0		;
	
	ret
	
BITS:
	.byte 1, 2, 4, 8, 16, 32, 64, 128
	
;********************************************************
; update module channels...
;********************************************************
Module_UpdateChannels:
	mov	x, #0
	mov	a, patt_update
	
_muc_loop:
	lsr	a
	push	a
	mov	a, #0
	rol	a
	mov	t_hasdata, a
	
	call	Module_UpdateChannel
	
	pop	a
	
	inc	x
	cmp	x, #8
	bne	_muc_loop
	
	ret
	
;********************************************************
; update module channel
;********************************************************
Module_UpdateChannel:
	SPROC
	
	;--------------------------------------
	; get data pointers
	;--------------------------------------
	mov	y, ch_instr+x
	dec	y
	mov	a, !MODULE+MOD_ITABLE_L+y
	mov	p_instr, a
	mov	a, !MODULE+MOD_ITABLE_H+y
	mov	p_instr+1, a
	

	mov	t_flags, #0
	cmp	t_hasdata, #0
	beq	_muc_nopatterndata
	
	call	Channel_ProcessData
	bra	_muc_pa
_muc_nopatterndata:
	call	Channel_CopyTemps
_muc_pa:
	
	call	Channel_ProcessAudio
	ret

;********************************************************	
Channel_ProcessData:
;********************************************************

	cmp	mod_tick, #0		; skip tick0 processing on other ticks
	bne	_cpd_non0		;
	
	mov	a, ch_flags+x
	mov	m6, a
	
	bbc0	m6, _cpd_no_note	; test for note
	mov	a, ch_note+x		;
	cmp	a, #254			; test notecut/noteoff
	beq	_cpd_notecut		;
	bcs	_cpd_noteoff		;
	
_cpd_note:				; dont start note on glissando
	bbc3	m6, _cpdn_test_for_glis	;
	mov	a, ch_command+x		;
	cmp	a, #7			;
	beq	_cpd_note_next		;
_cpdn_test_for_glis:			;
					;
	call	Channel_StartNewNote	;
	bra	_cpd_note_next		;
	
_cpd_notecut:				;notecut:
	mov	a, #0			; cut volume
	mov	ch_volume+x, a		;
	and	m6, #~CF_NOTE		; clear note flag
	bra	_cpd_note_next		;
	
_cpd_noteoff:				;noteoff:
	and	m6, #~(CF_NOTE|CF_KEYON); clear note and keyon flags
	
_cpd_note_next:
	
	bbc1	m6, _cpdn_no_instr	; apply instrument SETPAN
	mov	y, #INS_SETPAN		;
	mov	a, [p_instr]+y		;
	bmi	_cpdi_nsetpan		;
	mov	ch_panning+x, a		;
_cpdi_nsetpan:				;
	
	mov	y, ch_sample+x		; apply sample SETPAN
;	beq	_cpdi_nosample		;
	mov	a, !MODULE+MOD_STABLE_L+y	;
	mov	m0, a			;
	mov	a, !MODULE+MOD_STABLE_H+y	;
	mov	m0+1, a			;
	mov	y, #SAMP_DVOL		; copy default volume
	mov	a, [m0]+y		;
	mov	ch_volume+x, a		;
	mov	y, #SAMP_SETPAN		;
	mov	a, [m0]+y		;
	bmi	_cpdi_nsetpan_s		;
	mov	ch_panning+x, a		;
_cpdi_nsetpan_s:
_cpdi_nosample:
_cpdn_no_instr:

	and	m6, #~CF_NOTE
	
_cpd_no_note:				;

	mov	a, m6			; save flag mods
	mov	ch_flags+x, a		;
	
	and	a, #(CF_NOTE|CF_INSTR)	; test for note or instrument
	beq	_no_note_or_instr	;
	call	Channel_ResetVolume	; and reset volume things
_no_note_or_instr:			;

_cpd_non0:				; nonzero ticks: just update audio

	SPROC
	
	mov	a, ch_flags+x		; test and process volume command
	and	a, #CF_VCMD		;
	beq	_skip_vcmd		;
	call	Channel_ProcessVolumeCommand
_skip_vcmd:
	SPROC
	call	Channel_CopyTemps	; copy t values
	
	mov	a, ch_flags+x		; test and process command
	and	a, #CF_CMD		;
	beq	_skip_cmd		;
	call	Channel_ProcessCommand	;
_skip_cmd:
	
	ret

;********************************************************
Channel_CopyTemps:
;********************************************************

	mov	a, ch_pitch_l+x		; prepare for effects processing.....
	mov	y, ch_pitch_h+x		;
	movw	t_pitch, ya		;
	mov	a, ch_volume+x		;
	mov	y, ch_panning+x		;
	movw	t_volume, ya		;
	mov	t_sampoff, #0		;
	
	
	ret

;********************************************************
Channel_StartNewNote:
;********************************************************
	
	mov	a, ch_note+x		; pitch = note * 64
	mov	y, #64			;
	mul	ya			;
	mov	ch_pitch_l+x, a		;
	mov	ch_pitch_h+x, y		;
	
	mov	a, ch_instr+x		; test for instrument and copy sample!
	beq	_csnn_no_instr		;
	mov	y, #INS_SAMPLE		;
	mov	a, [p_instr]+y		;
	mov	ch_sample+x, a		;
_csnn_no_instr:

	or	t_flags, #TF_START	; set start flag
	ret
	
;********************************************************
Channel_ResetVolume:
;********************************************************
	mov	a, #255			; reset fadeout
	mov	ch_fadeout+x, a		;----------------
	mov	a, #0			; reset envelope
	mov	ch_env_node+x, a	;
	mov	ch_env_tick+x, a	;----------------
	mov	ch_cmem+x, a		; reset CMem
					;----------------
	mov	a, ch_flags+x		; set KEYON
	or	a, #CF_KEYON		; clear FADE
	and	a, #~CF_FADE		;
	mov	ch_flags+x, a		;----------------
	ret
	
;********************************************************
Channel_ProcessAudio:
;********************************************************

	SPROC
	mov	y, ch_sample+x			; m5 = sample address
;	beq	_cpa_nsample			;
	mov	a, !MODULE+MOD_STABLE_L+y	;
	mov	m5, a				;
	mov	a, !MODULE+MOD_STABLE_H+y	;
	mov	m5+1, a				;
_cpa_nsample:					;
	
	call	Channel_ProcessEnvelope
	
	mov	a, ch_flags+x			; process FADE
	and	a, #CF_FADE			;
	beq	_skip_fade			;
	mov	a, ch_fadeout+x			;
	setc					;
	mov	y, #INS_FADEOUT			;
	sbc	a, [p_instr]+y			;
	bcs	_subfade_noverflow		;	
	mov	a, #0				;
_subfade_noverflow:				;
	mov	ch_fadeout+x, a			;
_skip_fade:					;

	mov	a, !BITS+x
	and	a, sfx_mask
	bne	_sfx_override

	mov	a, t_flags			; exit if 'note delay' is set
	and	a, #TF_DELAY			;
	beq	_cpa_ndelay			;
_sfx_override:
	ret					;
_cpa_ndelay:					;

	;----------------------------------------
	; COMPUTE VOLUME:
	; V*CV*SV*GV*VEV*FADE
	; m0 = result (0..255)
	;----------------------------------------
	
	mov	y, #INS_GVOL
	mov	a, [p_instr]+y
	push	a
	mov	y, #SAMP_GVOL
	mov	a, [m5]+y
	push	a
	
	mov	a, t_volume			; y = 8-BIT VOLUME
	asl	a				;
	asl	a				;		
	bcc	_cpa_clamp_vol			;	
	mov	a, #255				;
_cpa_clamp_vol					;
	mov	y, a				;
	
	mov	a, ch_cvolume+x			; *= CV
	asl	a				;
	asl	a
	bcs	_calcvol_skip_cv		;
	mul	ya				;
_calcvol_skip_cv:				;

	pop	a				; *= SV
	asl	a				;
	asl	a
	bcs	_calcvol_skip_sv		;
	mul	ya				;
_calcvol_skip_sv:				;

	pop	a				;
	asl	a				;
	bcs	_calcvol_skip_iv		;
	mul	ya				;
_calcvol_skip_iv:
	
	mov	a, mod_gvol			; *= GV
	asl	a				;
	bcs	_calcvol_skip_gvol		;
	mul	ya				;
_calcvol_skip_gvol:				;

	mov	a, t_env			; *= VEV
	mul	ya				;
	
	mov	a, ch_fadeout+x			; *= FADE
	mul	ya				;
	
	mov	a, module_vol
	mul	ya
	
	mov	a, y				; store 7bit result
	lsr	a				; 
	mov	m2, a
	
	cmp	t_flags, #80h
	bcs	_dont_hack_gain
	cmp	a, #0
	bne	_gain_not_zero			; map value 0 to fast linear decrease
	mov	a, #%10011100			; (8ms)
_gain_not_zero:					;
	cmp	a, #127				; map value 127 to fast linear increase
	bne	_gain_not_max			; (8ms)
	mov	a, #%11011100			;
_gain_not_max:					;
	mov	m2, a				;
_dont_hack_gain:
	
	;---------------------------------------
	; compute PANNING
	;---------------------------------------
	mov	a, t_panning			; a = panning 0..127	
	asl	a				;	
	bpl	_clamppan			;
	dec	a				;
_clamppan:					;	
	mov	m1+1, a				; store panning (volume) levels
	eor	a, #127				;
	mov	m1, a				;
	
	mov	a, ch_flags+x			; apply surround (R = -R)
	and	a, #CF_SURROUND			;
	beq	_cpa_nsurround			;
	eor	m1+1, #255			;
	inc	m1+1				;
_cpa_nsurround:					;
	
	;---------------------------------------
	; compute PITCH
	;---------------------------------------
	cmp	x, #1

	mov	y, #SAMP_PITCHBASE		; m3 = t_pitch PITCHBASE
	mov	a, [m5]+y			;
	clrc					;
	adc	a, t_pitch_l			;
	mov	m3, a				;
	inc	y				;
	mov	a, [m5]+y			;
	adc	a, t_pitch_h			;
	mov	m3+1, a				;
	
	mov	y, a				; m0 = octave
	mov	a, !LUT_DIV3+y			;
	mov	m0, a				;
	
	asl	a				; m3 -= (oct*3) << 8
	adc	a, m0				;
	mov	m0+1, a				;
	mov	a, m3+1				;
	setc					;
	sbc	a, m0+1				;
	
	
	asl	m3				; m3 = m3*2 + LUT_FTAB base
	rol	a				;
	adc	m3, #(LUT_FTAB&0FFh)		;
	adc	a, #(LUT_FTAB>>8)			; 
	mov	m3+1, a				;
	
	mov	y, #0				; read ftab[f]
	mov	a, [m3]+y			;
	mov	m4, a				;
	inc	y				;
	mov	a, [m3]+y			;
	push	a				;
	
	mov	a, #8				; y = 8-oct
	setc					;
	sbc	a, m0				;
	mov	y, a				;
	
	pop	a				; a,m4 = ftab value
	beq	_no_pitch_shift			; skip shift if 0
	
	lsr	a				; shift by (8-oct)
	ror	m4				;
	dec	y				;
	beq	_no_pitch_shift			;
	lsr	a				;
	ror	m4				;
	dec	y				;
	beq	_no_pitch_shift			;
	lsr	a				;
	ror	m4				;
	dec	y				;
	beq	_no_pitch_shift			;
	lsr	a				;
	ror	m4				;
	dec	y				;
	beq	_no_pitch_shift			;
	lsr	a				;	
	ror	m4				;	
	dec	y				;
	beq	_no_pitch_shift			;
	lsr	a				;
	ror	m4				;
	dec	y				;
	beq	_no_pitch_shift			;
	lsr	a				;
	ror	m4				;
	dec	y				;
	beq	_no_pitch_shift			;
	lsr	a				;
	ror	m4				;
	
_no_pitch_shift:
	
	mov	m4+1, a
	
	;----------------------------------------
	; m1 = VOL/VOLR
	; m2 = GAIN
	; m4 = PITCH
	;----------------------------------------
	mov	a, x				; DSPA = voices[x]
	xcn	a				;
	mov	SPC_DSPA, a			;
						;------------------------------
	mov	a, t_flags			; test for KEYON
	and	a, #TF_START			;
	beq	_cpa_nstart			;------------------------------
						;keyon:
	mov	y, #SAMP_DINDEX			; set SRCN
	mov	a, [m5]+y			;
	or	SPC_DSPA, #DSPV_SRCN		;
	mov	SPC_DSPD, a			;------------------------------
	;----------------------------------------
	; **TODO: SAMPLE OFFSET
	;----------------------------------------
	mov	SPC_DSPA, #DSP_KON		; set KON bit
	mov	a, !BITS+x			;
	mov	SPC_DSPD, a			;------------------------------
	mov	a, x				; restore DSPA = voices[x]
	xcn	a				;
	mov	SPC_DSPA, a			;
;------------------------------------------------
_cpa_nstart:
;------------------------------------------------
	
	
	mov	SPC_DSPD, m1			; set VOLUME
	inc	SPC_DSPA			;
	mov	SPC_DSPD, m1+1			;
	inc	SPC_DSPA			;------------------------------
	mov	SPC_DSPD, m4			; set PITCH
	inc	SPC_DSPA			;
	mov	SPC_DSPD, m4+1			;
	inc	SPC_DSPA			;
	inc	SPC_DSPA			;------------------------------
	mov	SPC_DSPD, #00h			; disable ADSR
	or	SPC_DSPA, #07h			; set GAIN
	mov	SPC_DSPD, m2			;------------------------------

	;----------------------------------------
	; **TODO: RESTORE SAMPLE OFFSET
	;----------------------------------------
	
	SPROC
	ret
	
	
;********************************************************
Channel_ProcessEnvelope:
;********************************************************
	mov	y, #INS_ENVLEN			; test for envelope
	mov	a, [p_instr]+y			;
	mov	m0, a

	bne	_envelope_valid			;if no envelope:
	mov	t_env, #255			; set to max
	
	mov	a, ch_flags+x			; start fade on KEYOFF
	and	a, #CF_KEYON			;
	beq	_env_quit			;
	jmp	_env_setfade			;
_env_quit:
	ret					;
_envelope_valid:				;
	
	mov	a, ch_env_node+x		; read envelope node data
	
	clrc					; m1/m2
	adc	a, #INS_ENVDATA			;
	mov	y, a				;
	mov	a, [p_instr]+y			;
	mov	m1, a				;
	inc	y				;
	mov	a, [p_instr]+y			;
	mov	m1+1, a				;
	inc	y				;
	mov	a, [p_instr]+y			;
	mov	m2, a				;
	inc	y				;
	mov	a, [p_instr]+y			;
	mov	m2+1, a				;
	
	SPROC
	mov	a, ch_env_tick+x		; test zero/nonzero tick
	bne	_env_nonzero_tick		;
						;ZEROTICK:
	mov	a, m1				; copy Y level
	mov	ch_env_y_h+x, a			;
	mov	a, #0				;
	mov	ch_env_y_l+x, a			;
	bra	_env_zerotick			;
	
_env_nonzero_tick:				;NONZERO:
	mov	a, ch_env_y_l+x
	clrc
	adc	a, m2
	mov	ch_env_y_l+x, a
	mov	a, ch_env_y_h+x
	adc	a, m2+1
	
	bpl	_catch_negative			; clamp result 0.0->64.0
	mov	a, #0				;
	mov	ch_env_y_h+x, a			;
	mov	ch_env_y_l+x, a			;
	bra	_env_zerotick			;
_catch_negative:				;
	cmp	a, #64				;
	bcc	_catch_plus			;
	mov	a, #64				;
	mov	ch_env_y_h+x, a			;
	mov	a, #0				;
	mov	ch_env_y_l+x, a			;
	bra	_env_zerotick			;
_catch_plus:					;
						;
	mov	ch_env_y_h+x, a			;
	
_env_zerotick:

	mov	a, ch_env_y_l+x			; t_env = env << 2
	mov	m1, a				;
	mov	a, ch_env_y_h+x			;
	asl	m1				;
	rol	a				;
	asl	m1				;
	rol	a				;
	
	bcc	_env_shift_clamp		; clamp to 255
	mov	a, #255				;
_env_shift_clamp:				;
	mov	t_env, a			;
	
	mov	a, ch_flags+x			; dont advance if "keyon" and node=sustain
	and	a, #CF_KEYON			;
	beq	_env_nsustain			;
	mov	y, #INS_ENVSUS			;
	mov	a, [p_instr]+y			;
	cmp	a, ch_env_node+x		;
	bne	_env_nsustain			;
	ret					;
_env_nsustain:					;
	
	inc	ch_env_tick+x			; increment tick
	mov	a, ch_env_tick+x		;
	cmp	a, m1+1				; exit if < duration
	bcc	_env_exit			;
	
	mov	a, #0				; reset tick
	mov	ch_env_tick+x, a		;
	
	mov	y, #INS_ENVLOOPEND		; turn on FADE if keyoff and loop
	mov	a, [p_instr]+y			;
	cmp	a, #255				;
	beq	_env_no_loop			;
	mov	a, ch_flags+x			;	
	and	a, #CF_KEYON			;	
	bne	_env_no_fade			;	
	mov	a, ch_flags+x			;
	or	a, #CF_FADE			;
	mov	ch_flags+x, a			;
_env_no_fade:
	
	mov	a, ch_env_node+x		; test for loop point
;	mov	y, #INS_ENVLOOPEND		;
	cmp	a, [p_instr]+y			;
	bne	_env_loop_test			;
	mov	y, #INS_ENVLOOPST
	mov	a, [p_instr]+y
	mov	ch_env_node+x, a
	ret
_env_loop_test:					;
_env_no_loop:
	
	mov	a, ch_env_node+x
	setc					; suspicious...
	sbc	m0, #4
	cmp	a, m0				; test for envelope end
	beq	_env_setfade			;
	clrc					; increment node
	adc	a, #4				;
	mov	ch_env_node+x, a		;
	
	ret
	
_env_setfade:
	mov	a, ch_flags+x
	or	a, #CF_FADE
	mov	ch_flags+x, a
_env_exit:					;
	ret

;********************************************************
Channel_ProcessVolumeCommand:
;********************************************************
	mov	a, ch_volume+x
	mov	y, ch_vcmd+x
	mov	m0, y
	call	do_vcmd
	mov	ch_volume+x, a
	ret
	
do_vcmd:
	cmp	y, #65
	bcc	vcmd_setvol
	cmp	y, #75
	bcc	vcmd_finevolup
	cmp	y, #85
	bcc	vcmd_finevoldown
	cmp	y, #95
	bcc	vcmd_volup
	cmp	y, #105
	bcc	vcmd_voldown
	cmp	y, #193
	bcs	vcmd_invalid
	cmp	y, #128
	bcs	vcmd_pan
vcmd_invalid:
	ret
	
;--------------------------------------------------------
; 00-64 set volume
;--------------------------------------------------------
vcmd_setvol:
	cmp	mod_tick, #0		; a = volume
	bne	exit_vcmd		;
	mov	a, y			;
exit_vcmd:				;
	ret				;
	
;--------------------------------------------------------
; 65-74 fine vol up
;--------------------------------------------------------
vcmd_finevolup:
	sbc	m0, #65			; m0 = rate (-1)
	
	cmp	mod_tick, #0
	bne	exit_vcmd
	
_vcmd_add_sat64
	adc	a, m0			; a += rate (+1)
	cmp	a, #65			; saturate to 64
	bcc	exit_vcmd		;
	mov	a, #64			;
	ret				;
	
;--------------------------------------------------------
; 75-84 fine vol down
;--------------------------------------------------------
vcmd_finevoldown:
	sbc	m0, #75-1		; m0 = rate [carry is cleared]

	cmp	mod_tick, #0
	bne	exit_vcmd

_vcmd_sub_sat0:	
	sbc	a, m0			; a -= rate
	bcs	exit_vcmd		; saturate lower bound to 0
	mov	a, #0			;
	ret				;
	
;--------------------------------------------------------
; 85-94 vol up
;--------------------------------------------------------
vcmd_volup:
	sbc	m0, #85			; m0 = rate (-1)
	cmp	mod_tick, #0
	beq	exit_vcmd
	bra	_vcmd_add_sat64
	
;--------------------------------------------------------
; 95-104 vol down
;--------------------------------------------------------
vcmd_voldown:
	sbc	m0, #95-1
	cmp	mod_tick, #0
	beq	exit_vcmd
	bra	_vcmd_sub_sat0
	
;--------------------------------------------------------
; 128-192 set pan
;--------------------------------------------------------
vcmd_pan:
	cmp	mod_tick, #0		; set panning
	bne	exit_vcmd		;
	push	a			;
	mov	a, y			;
	sbc	a, #128			;
	mov	ch_panning+x, a		;
	pop	a			;
	ret				;

command_memory_map:	
	.byte 00h, 00h, 00h, 10h, 20h, 20h, 30h, 70h, 00h
	;       A    B    C    D    E    F    G    H    I
	.byte 40h, 10h, 10h, 00h, 10h, 50h, 10h, 80h, 70h
	;       J    K    L    M    N    O    P    Q    R
	.byte 60h, 00h, 70h, 00h, 10h, 00h, 70h, 00h
	;       S    T    U    V    W    X    Y    Z
	
;********************************************************
Channel_ProcessCommandMemory:
;********************************************************
	
	mov	y, ch_command+x
	
	mov	a, !command_memory_map-1+y
	beq	_cpc_quit		; 0 = no memory!
	mov	m0, x
	clrc
	adc	a, m0
	mov	y, a
	
	
	cmp	y, #70h			; <7 : single param
	bcc	_cpcm_single		;
;--------------------------------------------------------
_cpcm_double:				; >=7: double param
;--------------------------------------------------------

	mov	a, !PatternMemory-10h+y
	mov	m0, a
	mov	a, ch_param+x
	cmp	a, #10h
	bcc	_cpcmd_h_clr
	push	a
	and	m0, #0Fh
	or	a, m0
	mov	m0, a
	pop	a
_cpcmd_h_clr:
	and	a, #0Fh
	beq	_cpcmd_l_clr
	and	m0, #0F0h
	or	a, m0
	mov	m0, a
_cpcmd_l_clr:
	mov	a, m0
	mov	ch_param+x, a
	mov	!PatternMemory-10h+y, a
	ret
;--------------------------------------------------------
_cpcm_single:
;--------------------------------------------------------

	mov	a, ch_param+x
	beq	_cpcms_clear
	mov	!PatternMemory-10h+y, a
	ret
_cpcms_clear:
	mov	a, !PatternMemory-10h+y
	mov	ch_param+x, a	
_cpc_quit:
	ret

;********************************************************
Channel_ProcessCommand:
;********************************************************
	
	mov	a, ch_command+x		; exit if cmd = 0 
	beq	_cpc_quit		;
	
	cmp	mod_tick, #0		; process MEMORY on t0
	bne	_cpc_nott0		;
	call	Channel_ProcessCommandMemory
_cpc_nott0:

	mov	y, ch_command+x		; setup jump address
	mov	a, !CMD_JUMPTABLE_L-1+y	;
	mov	!cpc_jump+1, a		;
	mov	a, !CMD_JUMPTABLE_H-1+y	;
	mov	!cpc_jump+2, a		;
	
	mov	a, ch_param+x		; preload data
	mov	y, mod_tick		;
	
	;-------------------------------
	; a = param
	; y = tick
	; Z = tick=0
	;-------------------------------
	
cpc_jump:
	jmp	$0011

;-----------------------------------------------------------------------
CMD_JUMPTABLE_L:
;-----------------------------------------------------------------------
	.byte	Command_SetSpeed & 0FFh			; Axx
	.byte	Command_SetPosition & 0FFh		; Bxx
	.byte	Command_PatternBreak & 0FFh		; Cxx
	.byte	Command_VolumeSlide & 0FFh		; Dxy
	.byte	Command_PitchSlideDown & 0FFh		; Exy
	.byte	Command_PitchSlideUp & 0FFh		; Fxy
	.byte	Command_Glissando & 0FFh		; Gxx
	.byte	Command_Vibrato & 0FFh			; Hxy
	.byte	Command_Tremor & 0FFh			; Ixy
	.byte	Command_Arpeggio & 0FFh			; Jxy
	.byte	Command_VolumeSlideVibrato & 0FFh	; Kxy
	.byte	Command_VolumeSlideGliss & 0FFh		; Lxy
	.byte	Command_SetChannelVolume & 0FFh		; Mxx
	.byte	Command_ChannelVolumeSlide & 0FFh	; Nxy
	.byte	Command_SampleOffset & 0FFh		; Oxx
	.byte	Command_PanningSlide & 0FFh		; Pxy
	.byte	Command_RetriggerNote & 0FFh		; Qxy
	.byte	Command_Tremolo & 0FFh			; Rxy
	.byte	Command_Extended & 0FFh			; Sxy
	.byte	Command_Tempo & 0FFh			; Txy
	.byte	Command_FineVibrato & 0FFh		; Uxy
	.byte	Command_SetGlobalVolume & 0FFh		; Vxx
	.byte	Command_GlobalVolumeSlide & 0FFh	; Wxy
	.byte	Command_SetPanning & 0FFh		; Xxx
	.byte	Command_Panbrello & 0FFh		; Yxy
	.byte	Command_MidiMacro & 0FFh		; Zxy
;-----------------------------------------------------------------------
CMD_JUMPTABLE_H:
;-----------------------------------------------------------------------
	.byte	Command_SetSpeed >> 8			; Axx
	.byte	Command_SetPosition >> 8		; Bxx
	.byte	Command_PatternBreak >> 8		; Cxx
	.byte	Command_VolumeSlide >> 8		; Dxy
	.byte	Command_PitchSlideDown >> 8		; Exy
	.byte	Command_PitchSlideUp >> 8		; Fxy
	.byte	Command_Glissando >> 8			; Gxx
	.byte	Command_Vibrato >> 8			; Hxy
	.byte	Command_Tremor >> 8			; Ixy
	.byte	Command_Arpeggio >> 8			; Jxy
	.byte	Command_VolumeSlideVibrato >> 8		; Kxy
	.byte	Command_VolumeSlideGliss >> 8		; Lxy
	.byte	Command_SetChannelVolume >> 8		; Mxx
	.byte	Command_ChannelVolumeSlide >> 8		; Nxy
	.byte	Command_SampleOffset >> 8		; Oxx
	.byte	Command_PanningSlide >> 8		; Pxy
	.byte	Command_RetriggerNote >> 8		; Qxy
	.byte	Command_Tremolo >> 8			; Rxy
	.byte	Command_Extended >> 8			; Sxy
	.byte	Command_Tempo >> 8			; Txy
	.byte	Command_FineVibrato >> 8		; Uxy
	.byte	Command_SetGlobalVolume >> 8		; Vxx
	.byte	Command_GlobalVolumeSlide >> 8		; Wxy
	.byte	Command_SetPanning >> 8			; Xxx
	.byte	Command_Panbrello >> 8			; Yxy
	.byte	Command_MidiMacro >> 8			; Zxy

;=======================================================================
Command_SetSpeed:
;=======================================================================
	bne	cmd_exit1			;on tick0:
	cmp	a, #0				; if param != 0
	beq	cmd_exit1			; mod_speed = param
	mov	mod_speed, a			;
cmd_exit1:					;
	ret					;
;=======================================================================
Command_SetPosition:
;=======================================================================
	bne	cmd_exit1			;on tick0:
	mov	pattjump_index, a		; set jump index
	mov	pattjump_enable, #1		; enable pattern jump
	ret					;
;=======================================================================
Command_PatternBreak:
;=======================================================================
	; nonzero params are not supported
	;
	bne	cmd_exit1			;on tick0:
	mov	pattjump_index, mod_position	; index = position+1
	inc	pattjump_index			; enable pattern jump(break)
	mov	pattjump_enable, #1		;
	ret
;=======================================================================
Command_VolumeSlide:
;=======================================================================
	mov	m0, t_volume			; slide volume
	mov	m0+1, #64			;
	call	DoVolumeSlide			;
	mov	t_volume, a			;
	mov	ch_volume+x, a			;
	ret					;
;=======================================================================
Command_PitchSlideDown:
;=======================================================================
	call	PitchSlide_Load			; m0 = slide amount
	movw	ya, t_pitch			; pitch -= m0
	subw	ya, m0				;
	bmi	_exx_zero			; saturate lower to 0
	movw	t_pitch, ya			;
	mov	ch_pitch_l+x, a			;
	mov	ch_pitch_h+x, y			;
	ret					;
;---------------------------------------------------------------------
_exx_zero:
;---------------------------------------------------------------------
	mov	a, #0				; zero pitch
	mov	y, #0				;
	movw	t_pitch, ya			;
	mov	ch_pitch_l+x, a			;
	mov	ch_pitch_h+x, a			;
	ret					;
;=======================================================================
Command_PitchSlideUp:
;=======================================================================
	call	PitchSlide_Load			; m0 = slide amount
	movw	ya, t_pitch			;
	addw	ya, m0				;
	cmp	y, #01Ah			;
	bcs	_fxx_max			; clamp upper bound to 1A00H
	movw	t_pitch, ya			;
	mov	ch_pitch_l+x, a			;
	mov	ch_pitch_h+x, y			;
	ret					;
;-----------------------------------------------------------------------
_fxx_max:
;-----------------------------------------------------------------------
	mov	y, #01Ah			; max pitch
	mov	a, #0				;
	movw	t_pitch, ya			;
	mov	ch_pitch_l+x, a			;
	mov	ch_pitch_h+x, y			;
	ret					;
;=======================================================================
Command_Glissando:
;=======================================================================
	beq	cmd_exit1			; on tickn:
	
	mov	m0+1, #0			; m0 = xx*4 (slide amount)
	asl	a				;
	rol	m0+1				;
	asl	a				;
	rol	m0+1				;
	mov	m0, a				;
	
	mov	a, ch_note+x			; m1 = slide target
	mov	m1, #0				;
	lsr	a				;
	ror	m1				;
	lsr	a				;
	ror	m1				;
	mov	m1+1, a				;
	
	movw	ya, t_pitch			; test slide direction
	cmpw	ya, m1				;
	bcc	_gxx_slideup
;-----------------------------------------------
_gxx_slidedown:
;-----------------------------------------------
	subw	ya, m0				; subtract xx*4 from pitch
	bmi	_gxx_set			; saturate lower to target pitch
	cmpw	ya, m1				;
	bcc	_gxx_set			;
_gxx_set2:					;
	movw	t_pitch, ya			;
	mov	ch_pitch_l+x, a			;
	mov	ch_pitch_h+x, y			;
	ret					;
;-----------------------------------------------
_gxx_slideup:
;-----------------------------------------------
	addw	ya, m0				; add xx*4 to pitch
	cmpw	ya, m1				; saturate upper to target pitch
	bcs	_gxx_set			;
	bra	_gxx_set2			;
;-----------------------------------------------
_gxx_set:					; pitch = target
;-----------------------------------------------
	movw	ya, m1				;
	bra	_gxx_set2			;
	
;=======================================================================
Command_Vibrato:
;=======================================================================
	mov	a, #70h
	mov	m0, x
	clrc
	adc	a, m0
	mov	y, a
	mov	a, !PatternMemory-10h+y
	
	mov	m0, a
	and	m0, #0Fh
	
	lsr	a				; cmem += x*4
	lsr	a				;
	and	a, #111100b			;
	clrc					;
	adc	a, ch_cmem+x			;
	mov	ch_cmem+x, a			;
	
	mov	y, a				; a = sine[cmem]
	mov	a, !IT_FineSineData+y		;
	bpl	_hxx_plus
	
_hxx_neg:
	eor	a, #255
	inc	a
	mov	y, m0
	mul	ya
	mov	m0+1, y
	lsr	m0+1
	ror	a
	lsr	m0+1
	ror	a
	lsr	m0+1
	ror	a
	lsr	m0+1
	ror	a
	mov	m0, a
	movw	ya, t_pitch
	subw	ya, m0
	bmi	_hxx_zero
	movw	t_pitch, ya
	ret
_hxx_plus:
	mov	y, m0
	mul	ya
	mov	m0+1, y
	lsr	m0+1
	ror	a
	lsr	m0+1
	ror	a
	lsr	m0+1
	ror	a
	lsr	m0+1
	ror	a
	mov	y, m0+1
	addw	ya, t_pitch			; warning: might break something on highest note
	movw	t_pitch, ya
	ret
	
_hxx_zero:
	mov	t_pitch, #0
	mov	t_pitch+1, #0
	ret
	
;=======================================================================
Command_Tremor:					; unimplemented
;=======================================================================
	ret

;=======================================================================
Command_Arpeggio:
;=======================================================================
	bne	_jxx_other
	mov	a, #0
	mov	ch_cmem+x, a
	ret
_jxx_other:
	mov	a, ch_cmem+x
	inc	a
	cmp	a, #3
	bcc	_jxx_less3
	mov	a, #0
_jxx_less3:
	mov	ch_cmem+x, a
	
	cmp	a, #1
	beq	_jxx_x
	bcs	_jxx_y
	ret
	
_jxx_x:
	mov	a, ch_param+x
	
_jxx_add:
	
	and	a, #0F0h
	asl	a
	mov	m0+1, #0
	rol	m0+1
	asl	a
	rol	m0+1
	mov	m0, a
	movw	ya, t_pitch
	addw	ya, m0
	movw	t_pitch, ya
	ret
_jxx_y:
	mov	a, ch_param+x
	xcn	a
	bra	_jxx_add

;=======================================================================
Command_VolumeSlideVibrato:
;=======================================================================
	call	Command_Vibrato
	
	mov	a, ch_param+x
	mov	y, mod_tick
	mov	m0, t_volume			; slide volume
	mov	m0+1, #64			;
	call	DoVolumeSlide			;
	mov	t_volume, a			;
	mov	ch_volume+x, a			;
cmd_exit2:
	ret					;

;=======================================================================
Command_VolumeSlideGliss:			; unimplemented
;=======================================================================
	ret

;=======================================================================
Command_SetChannelVolume:
;=======================================================================
	bne	cmd_exit2			; on tick0:
	cmp	a, #65				;  cvolume = param > 64 ? 64 : param
	bcc	cscv_under65			;
	mov	a, #64				;
cscv_under65:					;
	mov	ch_cvolume+x, a			;
	ret					;

;=======================================================================
Command_ChannelVolumeSlide:
;=======================================================================
	mov	a, ch_cvolume+x			; slide channel volume
	mov	m0, a				; 
	mov	m0+1, #64			;
	mov	a, ch_param+x			;
	call	DoVolumeSlide			;
	mov	ch_cvolume+x, a			;
	ret					;
	
;=======================================================================
Command_SampleOffset:
;=======================================================================
	bne	cmd_exit2			; on tick0:
	mov	t_sampoff, a			;   set sampoff data
	ret					;
	
;=======================================================================
Command_PanningSlide:
;=======================================================================
	xcn	a
	mov	m0, t_panning			; slide panning
	mov	m0+1, #64			;
	call	DoVolumeSlide			;
	mov	t_panning, a			;
	mov	ch_panning+x, a			;
	ret					;
	
;=======================================================================
Command_RetriggerNote:
;=======================================================================
	
	and	a, #0Fh				; m0 = y == 0 ? 1 : x
	bne	_crn_x1				;
	inc	a				;
_crn_x1:					;	
	mov	m0, a				;
	
	mov	a, ch_cmem+x			;if cmem is 0:
	bne	_crn_cmem_n0			;  cmem = m0
	mov	a, m0				;
_crn_count_ret:
	mov	ch_cmem+x, a			;
	ret					;	
_crn_cmem_n0:					;else:
	dec	a				; dec cmem until 0
	bne	_crn_count_ret			;
						;RETRIGGER NOTE:
	mov	a, m0				; cmem = m0
	mov	ch_cmem+x, a			;
	
	;----------------------------------------
	; affect volume
	;----------------------------------------
	mov	a, ch_param+x
	xcn	a
	and	a, #0Fh
	mov	m1, a
	asl	a
	push	x
	mov	x, a
	mov	a, t_volume
	clrc
	jmp	[rnvtable+x]
rnvtable:
	.word	rnv_0
	.word	rnv_1
	.word	rnv_2
	.word	rnv_3
	.word	rnv_4
	.word	rnv_5
	.word	rnv_6
	.word	rnv_7
	.word	rnv_8
	.word	rnv_9
	.word	rnv_A
	.word	rnv_B
	.word	rnv_C
	.word	rnv_D
	.word	rnv_E
	.word	rnv_F
	
rnv_1:	dec	a
	bra	_rnv_sat0
rnv_2:	sbc	a, #2-1
	bra	_rnv_sat0
rnv_3:	sbc	a, #4-1
	bra	_rnv_sat0
rnv_4:	sbc	a, #8-1
	bra	_rnv_sat0
rnv_5:	sbc	a, #16-1
	bra	_rnv_sat0
rnv_6:	mov	y, #170
	mul	ya
	mov	a, y
	bra	_rnv_set
rnv_7:	lsr	a
	bra	_rnv_set
rnv_8:
rnv_0:	bra	_rnv_set
rnv_9:	inc	a
	bra	_rnv_sat64
rnv_A:	adc	a, #2
	bra	_rnv_sat64
rnv_B:	adc	a, #4
	bra	_rnv_sat64
rnv_C:	adc	a, #8
	bra	_rnv_sat64
rnv_D:	adc	a, #16
	bra	_rnv_sat64
rnv_E:	mov	y, #3
	mul	ya
	lsr	a
	bra	_rnv_sat64
rnv_F:	asl	a
	bra	_rnv_sat64
	
_rnv_sat0:
	bpl	_rnv_set
	mov	a, #0
	bra	_rnv_set
_rnv_sat64:
	cmp	a, #65
	bcc	_rnv_set
	mov	a, #64
_rnv_set:
	pop	x
	mov	t_volume, a
	mov	ch_volume+x, a
	or	t_flags, #TF_START
	
	
	ret
	
;=======================================================================
Command_Tremolo:				; unimplemented
;=======================================================================
	ret

;=======================================================================
Command_Extended:
;=======================================================================
	xcn	a				; setup jump to:
	and	a, #0Fh				; CmdExTab[x]
	mov	y, a				;
	mov	a, !CmdExTab_L+y		;
	mov	!cmdex_jmp+1, a			;
	mov	a, !CmdExTab_H+y		;
	mov	!cmdex_jmp+2, a			;
	
	mov	a, ch_param+x			; a = y
	and	a, #0Fh				; y = tick
	mov	y, mod_tick			; z = tick0
	
cmdex_jmp:
	jmp	0a0bh
	
SCommand_Null:
	ret
	
CmdExTab_L:
	.byte	SCommand_Echo & 0FFh
	.byte	SCommand_Null & 0FFh
	.byte	SCommand_Null & 0FFh
	.byte	SCommand_Null & 0FFh
	.byte	SCommand_Null & 0FFh
	.byte	SCommand_Null & 0FFh
	.byte	SCommand_Null & 0FFh
	.byte	SCommand_Null & 0FFh
	.byte	SCommand_Panning & 0FFh
	.byte	SCommand_SoundControl & 0FFh
	.byte	SCommand_Null & 0FFh
	.byte	SCommand_Null & 0FFh
	.byte	SCommand_NoteCut & 0FFh
	.byte	SCommand_NoteDelay & 0FFh
	.byte	SCommand_Null & 0FFh
	.byte	SCommand_Cue & 0FFh
CmdExTab_H:
	.byte	SCommand_Echo >> 8
	.byte	SCommand_Null >> 8
	.byte	SCommand_Null >> 8
	.byte	SCommand_Null >> 8
	.byte	SCommand_Null >> 8
	.byte	SCommand_Null >> 8
	.byte	SCommand_Null >> 8
	.byte	SCommand_Null >> 8
	.byte	SCommand_Panning >> 8
	.byte	SCommand_SoundControl >> 8
	.byte	SCommand_Null >> 8
	.byte	SCommand_Null >> 8
	.byte	SCommand_NoteCut >> 8
	.byte	SCommand_NoteDelay >> 8
	.byte	SCommand_Null >> 8
	.byte	SCommand_Cue >> 8

; S01 = turn on echo
; S02 = turn off echo
; S03 = turn on echo for all
; S04 = turn off echo for all
;=======================================================================
SCommand_Echo:
;=======================================================================
	mov	SPC_DSPA, #DSP_EON
	cmp	a, #1
	beq	_sce_enable_one
	bcc	cmd_exit3
	cmp	a, #3
	bcc	_sce_disable_one
	beq	_sce_disable_all
	cmp	a, #4
	beq	_sce_enable_all
cmd_exit3:
	ret
_sce_enable_one:
	mov	a, !BITS+x
	or	a, SPC_DSPD
	mov	SPC_DSPD, a
	ret
_sce_disable_one:
	mov	a, !BITS+x
	eor	a, #255
	and	a, SPC_DSPD
	mov	SPC_DSPD, a
	ret
_sce_enable_all:
	mov	SPC_DSPD, #0FFh
	ret
_sce_disable_all:
	mov	SPC_DSPD, #0
	ret
;=======================================================================
SCommand_Panning:
;=======================================================================
	bne	cmd_exit3			; on tick0:
	mov	m0, a				; panning = (y << 2) + (y >> 2)
	asl	a				;
	asl	a				;
	lsr	m0				;
	lsr	m0				;
	adc	a, m0				;
	mov	t_panning, a			;
	mov	ch_panning+x, a			;
	ret					;
;=======================================================================
SCommand_SoundControl:
;=======================================================================
	bne	cmd_exit3
	cmp	a, #1
	bne	cmd_exit3
	mov	a, ch_flags+x
	or	a, #CF_SURROUND
	mov	ch_flags+x, a
	mov	a, #32
	mov	ch_panning+x, a
	mov	t_panning, a
	ret
;=======================================================================
SCommand_NoteCut:
;=======================================================================
	cmp	a, mod_tick			; on tick Y:
	bne	cmd_exit3			;
	mov	a, #0				; zero volume
	mov	t_volume, a			;
	mov	ch_volume+x, a			;
	ret					;
;=======================================================================
SCommand_NoteDelay:
;=======================================================================
	cmp	a, mod_tick
	beq	scdelay_equ
	bcs	scdelay_lower
	
	ret
scdelay_lower:
	or	t_flags, #TF_DELAY
	ret
scdelay_equ:
	or	t_flags, #TF_START
	ret
;=======================================================================
SCommand_Cue:
;=======================================================================
	bne	cmd_exit3			;on tick0:
	inc	STATUS				; increment CUE value
	and	STATUS, #11101111b		; in status and send to
	mov	SPC_PORT2, STATUS		; snes
	ret					;
;=======================================================================
Command_Tempo:
;=======================================================================
	cmp	a, #20h
	bcc	_temposlide
	cmp	a, #80
	bcs	_txxu1
	mov	a, #80
_txxu1:	cmp	a, #200
	bcc	_txxu2
	mov	a, #200
_txxu2:	call	Module_ChangeTempo
	mov	SPC_CONTROL, #%111
	ret
_temposlide:
	cmp	a, #10h
	bcc	_txx_down
	and	a, #0Fh
	clrc
	adc	a, mod_bpm
	cmp	a, #200
	bcc	_txx_satH
	mov	a, #200
_txx_satH:
	call	Module_ChangeTempo
	mov	SPC_CONTROL, #%111
	ret
_txx_down:
	mov	m0, a
	mov	a, mod_bpm
	setc
	sbc	a, m0
	cmp	a, #80
	bcs	_txx_satH
	mov	a, #80
	call	Module_ChangeTempo
	mov	SPC_CONTROL, #%111
	ret
;=======================================================================
Command_FineVibrato:				; unimplemented
;=======================================================================
	ret

;=======================================================================
Command_SetGlobalVolume:
;=======================================================================
	bne	cmd_exit4			; set global volume on tick0
	cmp	a, #80h				;
	bcc	_vxx_nsat			; saturate to 80h
	mov	a, #80h				;
_vxx_nsat:					;
	mov	mod_gvol, a			;
cmd_exit4:					;
	ret					;
;=======================================================================
Command_GlobalVolumeSlide:
;=======================================================================
	mov	m0, mod_gvol			; slide global volume
	mov	m0+1, #128			; max 128
	call	DoVolumeSlide			;
	mov	mod_gvol, a			;
	ret					;
;=======================================================================
Command_SetPanning:
;=======================================================================
	bne	cmd_exit4			; set panning on tick0	
	lsr	a				;
	lsr	a				;
	mov	t_panning, a			;
	mov	ch_panning+x, a			;
	mov	a, ch_flags+x
	and	a, #~CF_SURROUND
	mov	ch_flags+x, a
	ret					;
;=======================================================================
Command_Panbrello:				; unimplemented
;=======================================================================
	ret
;=======================================================================
Command_MidiMacro:				; ?
;=======================================================================
	ret

;-----------------------------------------------------------------------
; a = param
; y = tick
; m0 = value
; m0+1 = upper bound
;
; return: a = result
;-----------------------------------------------------------------------
DoVolumeSlide:
;-----------------------------------------------------------------------
	mov	m1, a			; test param for slide behavior
					;-------------------------------
	and	a, #0Fh			; Dx0 : slide up
	beq	_dvs_up			;-------------------------------
	mov	a, m1			; D0y : slide down
	and	a, #0F0h		;
	beq	_dvs_down		;-------------------------------
	mov	a, m1			; DxF : slide up fine
	and	a, #0Fh			;
	cmp	a, #0Fh			;
	beq	_dvs_fineup		;-------------------------------
	mov	a, m1			; DFy : slide down fine
	cmp	a, #0F0h		;
	bcs	_dvs_finedown		;
_dvs_quit:				;-------------------------------
	mov	a, m0			; (invalid)
_dvs_exit:				;
	ret				;
;-----------------------------------------------------------------------
_dvs_finedown:				; DFy
;-----------------------------------------------------------------------
	cmp	y, #0			;on tick0:
	bne	_dvs_quit		;
	mov	a, m0			; a = volume - y
	and	m1, #0Fh		;
	sbc	a, m1			;
	bcs	_dvs_exit		; saturate lower bound to 0
	mov	a, #0			;
	ret				;
;-----------------------------------------------------------------------
_dvs_fineup:				; DxF
;-----------------------------------------------------------------------
	cmp	y, #0			;on tick0:
	bne	_dvs_quit		;
	mov	a, m1			; a = x + volume
	xcn	a			;
	and	a, #0Fh			;
	clrc				;
	adc	a, m0			;
	cmp	a, m0+1			; saturate upper to [m0.h]
	bcc	_dvs_exit		;
	mov	a, m0+1			;
	ret				;
;-----------------------------------------------------------------------
_dvs_down:				; D0y
;-----------------------------------------------------------------------
	cmp	m1,#0Fh			;on tick0 OR y == 15
	beq	_dvsd_15		;
	cmp	y, #0			;
	beq	_dvs_quit		;
_dvsd_15:				;
	mov	a, m0			; a = volume - param
	setc				;
	sbc	a, m1			;
	bcs	_dvs_exit		; saturate lower to 0
	mov	a, #0			;
	ret				;
;-----------------------------------------------------------------------
_dvs_up:				;
;-----------------------------------------------------------------------
	cmp	m1, #0F0h		;on tick0 OR x == 15
	beq	_dvsu_15		;
	cmp	y, #0			;
	beq	_dvs_quit		;
_dvsu_15:				;
	mov	a, m1			; a = x + volume
	xcn	a			;
	and	a, #0Fh			;
	clrc				;
	adc	a, m0			;
	cmp	a, m0+1			; saturte upper to [m0.h]
	bcc	_dvs_exit		;
	mov	a, m0+1			;
	ret				;
;-----------------------------------------------------------------------

;=======================================================================
; a = param
; y = tick
; return m0:word = slide amount
;=======================================================================
PitchSlide_Load:
;=======================================================================
	cmp	a, #0F0h			; Fx: fine slide
	bcs	_psl_fine			;
	cmp	a, #0E0h			; Ex: extra fine slide
	bcs	_psl_exfine			;
;-----------------------------------------------------------------------
_psl_normal:
;-----------------------------------------------------------------------
	cmp	y, #0				; no slide on tick0
	beq	_psl_zero			;
	mov	m0+1, #0			; m0 = a*4
	asl	a				;	
	rol	m0+1				;
	asl	a				;
	rol	m0+1				;
	mov	m0, a				;
	ret					;
;-----------------------------------------------------------------------
_psl_fine:
;-----------------------------------------------------------------------
	cmp	y, #0				; no slide on not tick0
	bne	_psl_zero			;
	mov	m0+1, #0			; m0 = y*4
	and	a, #0Fh				;	
	asl	a				;
	asl	a				;
	mov	m0, a				;
	ret					;
;-----------------------------------------------------------------------
_psl_exfine:
;-----------------------------------------------------------------------
	cmp	y, #0				; no slide on not tick0
	bne	_psl_zero			;
	mov	m0+1, #0			; m0 = y
	and	a, #0Fh				;	
	mov	m0, a				;
	ret					;
;-----------------------------------------------------------------------
_psl_zero:
;-----------------------------------------------------------------------
	mov	m0, #0
	mov	m0+1, #0
	ret

;************************************************************************************************************************************************

LUT_DIV3:
	.byte 0, 0, 0, 1, 1, 1, 2, 2, 2
	.byte 3, 3, 3, 4, 4, 4, 5, 5, 5
	.byte 6, 6, 6, 7, 7, 7, 8, 8, 8
	.byte 9, 9, 9,10,10
	
LUT_FTAB:
        .word 02174h, 0217Bh, 02183h, 0218Bh, 02193h, 0219Ah, 021A2h, 021AAh, 021B2h, 021BAh, 021C1h, 021C9h, 021D1h, 021D9h, 021E1h, 021E8h
        .word 021F0h, 021F8h, 02200h, 02208h, 02210h, 02218h, 0221Fh, 02227h, 0222Fh, 02237h, 0223Fh, 02247h, 0224Fh, 02257h, 0225Fh, 02267h
        .word 0226Fh, 02277h, 0227Fh, 02287h, 0228Fh, 02297h, 0229Fh, 022A7h, 022AFh, 022B7h, 022BFh, 022C7h, 022CFh, 022D7h, 022DFh, 022E7h
        .word 022EFh, 022F7h, 022FFh, 02307h, 0230Fh, 02317h, 0231Fh, 02328h, 02330h, 02338h, 02340h, 02348h, 02350h, 02358h, 02361h, 02369h
        .word 02371h, 02379h, 02381h, 0238Ah, 02392h, 0239Ah, 023A2h, 023AAh, 023B3h, 023BBh, 023C3h, 023CBh, 023D4h, 023DCh, 023E4h, 023EDh
        .word 023F5h, 023FDh, 02406h, 0240Eh, 02416h, 0241Fh, 02427h, 0242Fh, 02438h, 02440h, 02448h, 02451h, 02459h, 02462h, 0246Ah, 02472h
        .word 0247Bh, 02483h, 0248Ch, 02494h, 0249Dh, 024A5h, 024AEh, 024B6h, 024BEh, 024C7h, 024CFh, 024D8h, 024E0h, 024E9h, 024F2h, 024FAh
        .word 02503h, 0250Bh, 02514h, 0251Ch, 02525h, 0252Dh, 02536h, 0253Fh, 02547h, 02550h, 02559h, 02561h, 0256Ah, 02572h, 0257Bh, 02584h
        .word 0258Ch, 02595h, 0259Eh, 025A7h, 025AFh, 025B8h, 025C1h, 025C9h, 025D2h, 025DBh, 025E4h, 025ECh, 025F5h, 025FEh, 02607h, 0260Fh
        .word 02618h, 02621h, 0262Ah, 02633h, 0263Ch, 02644h, 0264Dh, 02656h, 0265Fh, 02668h, 02671h, 0267Ah, 02682h, 0268Bh, 02694h, 0269Dh
        .word 026A6h, 026AFh, 026B8h, 026C1h, 026CAh, 026D3h, 026DCh, 026E5h, 026EEh, 026F7h, 02700h, 02709h, 02712h, 0271Bh, 02724h, 0272Dh
        .word 02736h, 0273Fh, 02748h, 02751h, 0275Ah, 02763h, 0276Dh, 02776h, 0277Fh, 02788h, 02791h, 0279Ah, 027A3h, 027ACh, 027B6h, 027BFh
        .word 027C8h, 027D1h, 027DAh, 027E4h, 027EDh, 027F6h, 027FFh, 02809h, 02812h, 0281Bh, 02824h, 0282Eh, 02837h, 02840h, 0284Ah, 02853h
        .word 0285Ch, 02865h, 0286Fh, 02878h, 02882h, 0288Bh, 02894h, 0289Eh, 028A7h, 028B0h, 028BAh, 028C3h, 028CDh, 028D6h, 028E0h, 028E9h
        .word 028F2h, 028FCh, 02905h, 0290Fh, 02918h, 02922h, 0292Bh, 02935h, 0293Eh, 02948h, 02951h, 0295Bh, 02965h, 0296Eh, 02978h, 02981h
        .word 0298Bh, 02995h, 0299Eh, 029A8h, 029B1h, 029BBh, 029C5h, 029CEh, 029D8h, 029E2h, 029EBh, 029F5h, 029FFh, 02A08h, 02A12h, 02A1Ch
        .word 02A26h, 02A2Fh, 02A39h, 02A43h, 02A4Dh, 02A56h, 02A60h, 02A6Ah, 02A74h, 02A7Eh, 02A87h, 02A91h, 02A9Bh, 02AA5h, 02AAFh, 02AB9h
        .word 02AC3h, 02ACCh, 02AD6h, 02AE0h, 02AEAh, 02AF4h, 02AFEh, 02B08h, 02B12h, 02B1Ch, 02B26h, 02B30h, 02B3Ah, 02B44h, 02B4Eh, 02B58h
        .word 02B62h, 02B6Ch, 02B76h, 02B80h, 02B8Ah, 02B94h, 02B9Eh, 02BA8h, 02BB2h, 02BBCh, 02BC6h, 02BD1h, 02BDBh, 02BE5h, 02BEFh, 02BF9h
        .word 02C03h, 02C0Dh, 02C18h, 02C22h, 02C2Ch, 02C36h, 02C40h, 02C4Bh, 02C55h, 02C5Fh, 02C69h, 02C74h, 02C7Eh, 02C88h, 02C93h, 02C9Dh
        .word 02CA7h, 02CB2h, 02CBCh, 02CC6h, 02CD1h, 02CDBh, 02CE5h, 02CF0h, 02CFAh, 02D04h, 02D0Fh, 02D19h, 02D24h, 02D2Eh, 02D39h, 02D43h
        .word 02D4Dh, 02D58h, 02D62h, 02D6Dh, 02D77h, 02D82h, 02D8Ch, 02D97h, 02DA1h, 02DACh, 02DB7h, 02DC1h, 02DCCh, 02DD6h, 02DE1h, 02DECh
        .word 02DF6h, 02E01h, 02E0Bh, 02E16h, 02E21h, 02E2Bh, 02E36h, 02E41h, 02E4Bh, 02E56h, 02E61h, 02E6Ch, 02E76h, 02E81h, 02E8Ch, 02E97h
        .word 02EA1h, 02EACh, 02EB7h, 02EC2h, 02ECCh, 02ED7h, 02EE2h, 02EEDh, 02EF8h, 02F03h, 02F0Eh, 02F18h, 02F23h, 02F2Eh, 02F39h, 02F44h
        .word 02F4Fh, 02F5Ah, 02F65h, 02F70h, 02F7Bh, 02F86h, 02F91h, 02F9Ch, 02FA7h, 02FB2h, 02FBDh, 02FC8h, 02FD3h, 02FDEh, 02FE9h, 02FF4h
        .word 02FFFh, 0300Ah, 03015h, 03020h, 0302Ch, 03037h, 03042h, 0304Dh, 03058h, 03063h, 0306Eh, 0307Ah, 03085h, 03090h, 0309Bh, 030A7h
        .word 030B2h, 030BDh, 030C8h, 030D4h, 030DFh, 030EAh, 030F5h, 03101h, 0310Ch, 03117h, 03123h, 0312Eh, 0313Ah, 03145h, 03150h, 0315Ch
        .word 03167h, 03173h, 0317Eh, 03189h, 03195h, 031A0h, 031ACh, 031B7h, 031C3h, 031CEh, 031DAh, 031E5h, 031F1h, 031FCh, 03208h, 03213h
        .word 0321Fh, 0322Bh, 03236h, 03242h, 0324Dh, 03259h, 03265h, 03270h, 0327Ch, 03288h, 03293h, 0329Fh, 032ABh, 032B7h, 032C2h, 032CEh
        .word 032DAh, 032E5h, 032F1h, 032FDh, 03309h, 03315h, 03320h, 0332Ch, 03338h, 03344h, 03350h, 0335Ch, 03367h, 03373h, 0337Fh, 0338Bh
        .word 03397h, 033A3h, 033AFh, 033BBh, 033C7h, 033D3h, 033DFh, 033EBh, 033F7h, 03403h, 0340Fh, 0341Bh, 03427h, 03433h, 0343Fh, 0344Bh
        .word 03457h, 03463h, 0346Fh, 0347Bh, 03488h, 03494h, 034A0h, 034ACh, 034B8h, 034C4h, 034D1h, 034DDh, 034E9h, 034F5h, 03502h, 0350Eh
        .word 0351Ah, 03526h, 03533h, 0353Fh, 0354Bh, 03558h, 03564h, 03570h, 0357Dh, 03589h, 03595h, 035A2h, 035AEh, 035BAh, 035C7h, 035D3h
        .word 035E0h, 035ECh, 035F9h, 03605h, 03612h, 0361Eh, 0362Bh, 03637h, 03644h, 03650h, 0365Dh, 03669h, 03676h, 03683h, 0368Fh, 0369Ch
        .word 036A8h, 036B5h, 036C2h, 036CEh, 036DBh, 036E8h, 036F4h, 03701h, 0370Eh, 0371Bh, 03727h, 03734h, 03741h, 0374Eh, 0375Ah, 03767h
        .word 03774h, 03781h, 0378Eh, 0379Ah, 037A7h, 037B4h, 037C1h, 037CEh, 037DBh, 037E8h, 037F5h, 03802h, 0380Eh, 0381Bh, 03828h, 03835h
        .word 03842h, 0384Fh, 0385Ch, 03869h, 03876h, 03884h, 03891h, 0389Eh, 038ABh, 038B8h, 038C5h, 038D2h, 038DFh, 038ECh, 038FAh, 03907h
        .word 03914h, 03921h, 0392Eh, 0393Bh, 03949h, 03956h, 03963h, 03970h, 0397Eh, 0398Bh, 03998h, 039A6h, 039B3h, 039C0h, 039CEh, 039DBh
        .word 039E8h, 039F6h, 03A03h, 03A11h, 03A1Eh, 03A2Bh, 03A39h, 03A46h, 03A54h, 03A61h, 03A6Fh, 03A7Ch, 03A8Ah, 03A97h, 03AA5h, 03AB2h
        .word 03AC0h, 03ACEh, 03ADBh, 03AE9h, 03AF6h, 03B04h, 03B12h, 03B1Fh, 03B2Dh, 03B3Bh, 03B48h, 03B56h, 03B64h, 03B72h, 03B7Fh, 03B8Dh
        .word 03B9Bh, 03BA9h, 03BB6h, 03BC4h, 03BD2h, 03BE0h, 03BEEh, 03BFCh, 03C09h, 03C17h, 03C25h, 03C33h, 03C41h, 03C4Fh, 03C5Dh, 03C6Bh
        .word 03C79h, 03C87h, 03C95h, 03CA3h, 03CB1h, 03CBFh, 03CCDh, 03CDBh, 03CE9h, 03CF7h, 03D05h, 03D13h, 03D21h, 03D2Fh, 03D3Eh, 03D4Ch
        .word 03D5Ah, 03D68h, 03D76h, 03D85h, 03D93h, 03DA1h, 03DAFh, 03DBDh, 03DCCh, 03DDAh, 03DE8h, 03DF7h, 03E05h, 03E13h, 03E22h, 03E30h
        .word 03E3Eh, 03E4Dh, 03E5Bh, 03E6Ah, 03E78h, 03E86h, 03E95h, 03EA3h, 03EB2h, 03EC0h, 03ECFh, 03EDDh, 03EECh, 03EFAh, 03F09h, 03F18h
        .word 03F26h, 03F35h, 03F43h, 03F52h, 03F61h, 03F6Fh, 03F7Eh, 03F8Dh, 03F9Bh, 03FAAh, 03FB9h, 03FC7h, 03FD6h, 03FE5h, 03FF4h, 04002h
        .word 04011h, 04020h, 0402Fh, 0403Eh, 0404Dh, 0405Bh, 0406Ah, 04079h, 04088h, 04097h, 040A6h, 040B5h, 040C4h, 040D3h, 040E2h, 040F1h
        .word 04100h, 0410Fh, 0411Eh, 0412Dh, 0413Ch, 0414Bh, 0415Ah, 04169h, 04178h, 04188h, 04197h, 041A6h, 041B5h, 041C4h, 041D3h, 041E3h
        .word 041F2h, 04201h, 04210h, 04220h, 0422Fh, 0423Eh, 0424Eh, 0425Dh, 0426Ch, 0427Ch, 0428Bh, 0429Ah, 042AAh, 042B9h, 042C9h, 042D8h

IT_FineSineData:
	.byte   0,  2,  3,  5,  6,  8,  9, 11, 12, 14, 16, 17, 19, 20, 22, 23
	.byte  24, 26, 27, 29, 30, 32, 33, 34, 36, 37, 38, 39, 41, 42, 43, 44
	.byte  45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 56, 57, 58, 59
	.byte  59, 60, 60, 61, 61, 62, 62, 62, 63, 63, 63, 64, 64, 64, 64, 64
	.byte  64, 64, 64, 64, 64, 64, 63, 63, 63, 62, 62, 62, 61, 61, 60, 60
	.byte  59, 59, 58, 57, 56, 56, 55, 54, 53, 52, 51, 50, 49, 48, 47, 46
	.byte  45, 44, 43, 42, 41, 39, 38, 37, 36, 34, 33, 32, 30, 29, 27, 26
	.byte  24, 23, 22, 20, 19, 17, 16, 14, 12, 11,  9,  8,  6,  5,  3,  2
	.byte   0, -2, -3, -5, -6, -8, -9,-11,-12,-14,-16,-17,-19,-20,-22,-23
	.byte -24,-26,-27,-29,-30,-32,-33,-34,-36,-37,-38,-39,-41,-42,-43,-44
	.byte -45,-46,-47,-48,-49,-50,-51,-52,-53,-54,-55,-56,-56,-57,-58,-59
	.byte -59,-60,-60,-61,-61,-62,-62,-62,-63,-63,-63,-64,-64,-64,-64,-64
	.byte -64,-64,-64,-64,-64,-64,-63,-63,-63,-62,-62,-62,-61,-61,-60,-60
	.byte -59,-59,-58,-57,-56,-56,-55,-54,-53,-52,-51,-50,-49,-48,-47,-46
	.byte -45,-44,-43,-42,-41,-39,-38,-37,-36,-34,-33,-32,-30,-29,-27,-26
	.byte -24,-23,-22,-20,-19,-17,-16,-14,-12,-11, -9, -8, -6, -5, -3, -2
	
;****************************************************************************************
;* Sound Effects
;****************************************************************************************

map_15_127:
	 .byte 0,  8, 17, 25,
	 .byte 34, 42, 51, 59,
	 .byte 68, 76, 85, 93,
	 .byte 102, 110, 119, 127

;*************************************************************************
;* play sound effect
;*
;* m0 = params
;* vp sh
;* 
;* s = sample index
;* h = pitch ( 8 = 32000hz, h = pitch height >> 9 )
;* v = volume (15 = max)
;* p = panning (8 = center)
;*************************************************************************
SFX_Play:
;-------------------------------------------------------------------------
	mov	a, m0			; m1 = GAIN (0-15 = 0-127)
	xcn	a			;
	and	a, #0Fh			;
	mov	y, a			;
	mov	a, !map_15_127+y	;
	mov	m1, a			;---------------------------------
	mov	a, m0			; m2 = volumes
	and	a, #0Fh			;
	mov	y, a			;
	mov	a, !map_15_127+y	;
	mov	m2+1, a			;
	eor	a, #127			;
	mov	m2, a			;---------------------------------
	mov	a, m0+1			; m1.h = src
	and	a, #0F0h		;
	xcn	a			;
	clrc				;
	adc	a, #64			;
	mov	m1+1, a			;---------------------------------
	mov	a, m0+1			; m3 = pitch.h
	and	a, #0Fh			; (pitch.l = 0)
	asl	a			;
	mov	m3, a			;---------------------------------
	mov	a, sfx_mask		; test for unused channels
	asl	a			;
	bcc	_sfx_use1		;
	beq	_sfx_use0		;---------------------------------
	eor	sfx_next, #1		; otherwise alternate channels
	bne	_sfx_use1		;
;-------------------------------------------------------------------------
_sfx_use0:
;-------------------------------------------------------------------------
	mov	sfx_next, #0		;
	mov	SPC_DSPA, #064h		; set SRCN value for channel
	mov	SPC_DSPD, m1+1		;---------------------------------
	mov	SPC_DSPA, #DSP_KON	; set KON bit
	mov	SPC_DSPD, #%01000000	;
	or	sfx_mask, #%01000000	; set SFX flag
	mov	SPC_DSPA, #060h		; setup dsp pointer
	bra	_sfx_start		;
;-------------------------------------------------------------------------
_sfx_use1:
;-------------------------------------------------------------------------
;	cmp	stream_active, #0	; [STREAMING reserves channel7]
;	bne	_sfx_use0		;
	mov	sfx_next, #1
	mov	SPC_DSPA, #074h
	mov	SPC_DSPD, m1+1
	mov	SPC_DSPA, #DSP_KON
	mov	SPC_DSPD, #%10000000
	or	sfx_mask, #%10000000
	mov	SPC_DSPA, #070h
;-------------------------------------------------------------------------
_sfx_start:
;-------------------------------------------------------------------------
	mov	SPC_DSPD, m2		; VOLUME L
	inc	SPC_DSPA		;
	mov	SPC_DSPD, m2+1		; VOLUME R
	inc	SPC_DSPA		;
	mov	SPC_DSPD, #0		; PITCH L
	inc	SPC_DSPA		;
	mov	SPC_DSPD, m3		; PITCH H
	inc	SPC_DSPA		;
	inc	SPC_DSPA		;
	mov	SPC_DSPD, #0		; ADSR1
	or	SPC_DSPA, #7		;
	mov	SPC_DSPD, m1		; GAIN
	ret				;
;-------------------------------------------------------------------------

;*************************************************************************
;* update sound effects
;*************************************************************************
SFX_Update:
;-------------------------------------------------------------------------
	mov	SPC_DSPA, #DSP_ENDX	; reset SFX mask flags with ENDX
	mov	a, SPC_DSPD		;
	mov	SPC_DSPD, a		; <- clear endx
;	cmp	stream_active, #0
;	beq	_sfxu_nstreaming
;	and	a, #127
;_sfxu_nstreaming:
	eor	a, sfx_mask		;
	and	a, sfx_mask		;
	mov	sfx_mask, a		;
	ret				;
;-------------------------------------------------------------------------

;*************************************************************************
;*
;* Streaming
;*
;*************************************************************************

;**************************************************************************************
;* setup streaming system
;**************************************************************************************
Streaming_Init:
;--------------------------------------------------------------------------------------
	mov	a, #0				; reset region size
	call	Streaming_Resize		;
;--------------------------------------------------------------------------------------
	mov	a, #__BRK_ROUTINE__ & 0FFh	; set BRK/TCALL0 vector
	mov	!0FFDEH, a			;
	mov	a, #__BRK_ROUTINE__ >> 8	;
	mov	!0FFDFH, a			;
;--------------------------------------------------------------------------------------
	ret
	
;**************************************************************************************
;* RESIZE STREAM
;* a = newsize
;**************************************************************************************
Streaming_Resize:
;--------------------------------------------------------------------------------------
;	call	Streaming_CancelActive
;--------------------------------------------------------------------------------------
	mov	stream_size, a			;
	mov	a, #0FFh			; calc streaming region address H
	setc					;
	sbc	a, stream_size			;
	mov	stream_region, a		;
;--------------------------------------------------------------------------------------
	mov	a, #0			; copy stream buffer address
	mov	!StreamAddress, a	;
	mov	!StreamAddress+2, a	;
	mov	a, stream_region	;
	mov	!StreamAddress+1, a	;
	mov	!StreamAddress+3, a	;
;--------------------------------------------------------------------------------------
	ret
	
;Streaming_CancelActive:
;	mov	a, sfx_mask
;	and	a, #80h
;	beq	streaming_is_inactive
;	mov	y, #70h|DSPV_GAIN
;	mov	a, #0
;	movw	SPC_DSP, ya
;	
;streaming_is_inactive:
;	ret
	
;**************************************************************************************
;* START STREAM
;**************************************************************************************
Streaming_Activate:
;--------------------------------------------------------------------------------------
	mov	a, SPC_PORT2			; compute volume from panning
	and	a, #15				;
	asl	a				;
	asl	a				;
	asl	a				;
	mov	stream_volR, a			;
	eor	a, #127				;
	mov	stream_volL, a			;
;--------------------------------------------------------------------------------------
	mov	a, SPC_PORT2			; compute GAIN (v<<3)
	and	a, #0F0h			;
	lsr	a				;
	mov	stream_gain, a			;
;--------------------------------------------------------------------------------------
	mov	stream_rate, SPC_PORT3		; copy rate/PITCH
;--------------------------------------------------------------------------------------
	mov	stream_initial, #1		; set initial flag for data routine
;--------------------------------------------------------------------------------------
	call	StreamResetAddress		;
;--------------------------------------------------------------------------------------
	ret
	
;======================================================================================
StreamStartChannel:
;======================================================================================
	mov	stream_initial, #0	; reset flag
	or	sfx_mask, #80h		; patch sfx system
	mov	sfx_next, #1		; 
;--------------------------------------------------------------------------------------
	mov	SPC_DSPA, #074h		; SRCN = stream
	mov	SPC_DSPD, #80		;
;--------------------------------------------------------------------------------------
	mov	SPC_DSPA, #DSP_KON	; KEYON channel
	mov	SPC_DSPD, #80h		;
;--------------------------------------------------------------------------------------
	mov	SPC_DSPA, #070h		; copy volume (panning)
	mov	SPC_DSPD, stream_volL	; 
	inc	SPC_DSPA		;
	mov	SPC_DSPD, stream_volR	;
	inc	SPC_DSPA		;
;--------------------------------------------------------------------------------------
	mov	SPC_DSPD, #00H		; copy pitch
	inc	SPC_DSPA		;
	mov	SPC_DSPD, stream_rate	;
	inc	SPC_DSPA		;
	inc	SPC_DSPA		;
;--------------------------------------------------------------------------------------
	mov	SPC_DSPD, #0		; clear ADSR
	inc	SPC_DSPA		;
	inc	SPC_DSPA		;
;--------------------------------------------------------------------------------------
	mov	SPC_DSPD, stream_gain	; copy gain
;--------------------------------------------------------------------------------------

	ret
	
;**************************************************************************************
;* UPDATE STREAM
;**************************************************************************************
Streaming_Run:
;--------------------------------------------------------------------------------------
	mov	SPC_PORT0, #80h		; respond to SNES
;--------------------------------------------------------------------------------------
	push	a			; preserve regs
	push	x			;
	push	y			;
;--------------------------------------------------------------------------------------
_srw1:	cmp	SPC_PORT0, #80h		; wait for snes
	bcs	_srw1			;
;--------------------------------------------------------------------------------------
	mov	a, SPC_PORT0		; copy nchunks
	mov	stream_a, a		;
	mov	a, SPC_PORT1		; check for new note
	beq	_sr_nstart		;	
	call	Streaming_Activate	;
_sr_nstart:				;
	mov	x, SPC_PORT0		;
	mov	SPC_PORT0, x		; respond to snes
;--------------------------------------------------------------------------------------
_sr_start:
	mov	y, #0			; prepare COPYING...
	inc	x
_sr_wait_for_snes:			;
	cmp	x, SPC_PORT0		;
	bne	_sr_wait_for_snes	;
;--------------------------------------------------------------------------------------
	bra	_sr_copy

_sr_nextcopy:
	inc	x
_sr_wait3:
	cmp	x, SPC_PORT0
	bne	_sr_wait3
	
;--------------------------------------------------------------------------------------
_sr_copy:				; copy 9 bytes (16 SAMPLES)
;--------------------------------------------------------------------------------------
	mov	a, SPC_PORT2		; copy first 3 bytes
STRC0:	mov	!0FE00h+0+y, a	;
	mov	a, SPC_PORT3		;
STRC1:	mov	!0FE00h+1+y, a	;
	mov	SPC_PORT0, x		;-signal
	mov	a, SPC_PORT1		;
STRC2:	mov	!0FE00h+2+y, a	;
	inc	x			;
_wait1:					; wait for data
	cmp	x, SPC_PORT0		;
	bne	_wait1			;
;--------------------------------------------------------------------------------------
	mov	a, SPC_PORT2		; copy next 3 bytes
STRC3:	mov	!0FE00h+3+y, a	;
	mov	a, SPC_PORT3		;
STRC4:	mov	!0FE00h+4+y, a	;
	mov	SPC_PORT0, x		;-signal
	mov	a, SPC_PORT1		;
STRC5:	mov	!0FE00h+5+y, a	;
	inc	x			;
_wait2:					; wait for data
	cmp	x, SPC_PORT0		;
	bne	_wait2			;
;--------------------------------------------------------------------------------------
	mov	a, SPC_PORT2		; copy last 3 bytes
STRC6:	mov	!0FE00h+6+y, a	;
	mov	a, SPC_PORT3		;
STRC7:	mov	!0FE00h+7+y, a	;
	mov	SPC_PORT0, x		;-signal
	mov	a, SPC_PORT1		;
STRC8:	mov	!0FE00h+8+y, a	; wait for data
;--------------------------------------------------------------------------------------
	mov	a, y			; wr += 9
	clrc
	adc	a, #9			;
	mov	y, a			;
;--------------------------------------------------------------------------------------
	dec	stream_a		; decrement chunk counter
	bne	_sr_nextcopy		; loop until all blocks transferred
;--------------------------------------------------------------------------------------
_sr_exit:				; update write address
	mov	a, y			;
	mov	y, #0			;
	addw	ya, stream_write	;
	movw	stream_write, ya	;
	call	StreamSetupAddress	;
	cmp	stream_initial, #0
	beq	_sr_nstart2
	call	StreamStartChannel
_sr_nstart2:
;--------------------------------------------------------------------------------------
	pop	y			;4
	pop	x			;4
	pop	a			;4
	ret				;6
	
__BRK_ROUTINE__:
	asl	SPC_PORT0
	bcs	_brk_pass
	ret
_brk_pass:
	jmp	Streaming_Run
	
; (faster version without overflow checks)
;======================================================================================
StreamResetAddress:
;======================================================================================
	mov	y, stream_region
	mov	a, #0 
	movw	stream_write, ya
do_fast_ssa:
	mov	!STRC0+1, a
	inc	a
	mov	!STRC1+1, a
	inc	a
	mov	!STRC2+1, a
	inc	a
	mov	!STRC3+1, a
	inc	a
	mov	!STRC4+1, a
	inc	a
	mov	!STRC5+1, a
	inc	a
	mov	!STRC6+1, a
	inc	a
	mov	!STRC7+1, a
	inc	a
	mov	!STRC8+1, a
	mov	!STRC0+2, y
	mov	!STRC1+2, y
	mov	!STRC2+2, y
	mov	!STRC3+2, y
	mov	!STRC4+2, y
	mov	!STRC5+2, y
	mov	!STRC6+2, y
	mov	!STRC7+2, y
	mov	!STRC8+2, y
	ret
	
;======================================================================================
StreamSetupAddress:
;======================================================================================
	movw	ya, stream_write
;--------------------------------------------------------------------------------------
	cmp	a, #240				; do fast setup if akku won't overflow
	bcc	do_fast_ssa
	mov	!STRC0+1, a			; 1st address
	mov	!STRC0+2, y			;
	inc	a				;
	beq	_ssa_over_1			;
_ssa1:	mov	!STRC1+1, a			; 2nd
	mov	!STRC1+2, y			;
	inc	a				;
	beq	_ssa_over_2			;
_ssa2:	mov	!STRC2+1, a			; 3rd
	mov	!STRC2+2, y			;
	inc	a				;
	beq	_ssa_over_3			;
_ssa3:	mov	!STRC3+1, a			; 4th
	mov	!STRC3+2, y			;
	inc	a				;
	beq	_ssa_over_4			;
_ssa4:	mov	!STRC4+1, a			; 5th
	mov	!STRC4+2, y			;
	inc	a				;
	beq	_ssa_over_5			; 
_ssa5:	mov	!STRC5+1, a			; 6th
	mov	!STRC5+2, y			;
	inc	a				;
	beq	_ssa_over_6			;
_ssa6:	mov	!STRC6+1, a			; 7th
	mov	!STRC6+2, y			;
	inc	a				;
	beq	_ssa_over_7			;
_ssa7:	mov	!STRC7+1, a			; 8th
	mov	!STRC7+2, y			;
	inc	a				;
	beq	_ssa_over_8			;
_ssa8:	mov	!STRC8+1, a			; 9th
	mov	!STRC8+2, y			;
;--------------------------------------------------------------------------------------
	ret
	
_ssa_over_1:
	inc	y
	jmp	_ssa1
_ssa_over_2:
	inc	y
	jmp	_ssa2
_ssa_over_3:
	inc	y
	jmp	_ssa3
_ssa_over_4:
	inc	y
	jmp	_ssa4
_ssa_over_5:
	inc	y
	jmp	_ssa5
_ssa_over_6:
	inc	y
	jmp	_ssa6
_ssa_over_7:
	inc	y
	jmp	_ssa7
_ssa_over_8:
	inc	y
	jmp	_ssa8

;--------------------------------------------------------
MODULE = 1A00h
;--------------------------------------------------------
	
;--------------------------------------------------------
.END
;--------------------------------------------------------
