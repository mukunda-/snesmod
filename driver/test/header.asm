;----------------------------------------------------------------------
.segment "HEADER"
;----------------------------------------------------------------------

.import _start
.import _nmi

.export CART_HEADER

;----------------------------------------------------------------------
; ROM registration data
;----------------------------------------------------------------------
CART_HEADER:
.byte	0, 0			; B0 - maker code
.byte	"BREW"			; B2 - game code
.byte	0, 0, 0, 0, 0, 0, 0	; B6 - reserved
.byte	0			; BD - expansion ram size
.byte	0			; BE - special version
.byte	0			; BF - cartridge sub-number
.byte	"IT2SPC TEST     "	; C0 - game title
.byte	"     "			; D0 - game title (cont)
.byte	21h			; D5 - mode 20, 2.68mhz
.byte	00h			; D6 - cartridge type (ROM only)
.byte	09h			; D7 - ROM size (3-4mbit)
.byte	00h			; D8 - RAM size (no ram)
.byte	01h			; D9 - destination code (usa/canada)
.byte	33h			; DA - fixed byte
.byte	00h			; DB - mask ROM version
.word	0000h			; DC - complement check
.word	0000h			; DE - check sum
;----------------------------------------------------------------------
; native mode vectors
;----------------------------------------------------------------------
VEC_E0:	.word	0		; -
VEC_E2:	.word	0		; -
VEC_E4:	.word	0		; -
VEC_E6:	.word	0		; BRK
VEC_E8:	.word	0		; ABORT
VEC_EA:	.word	_nmi		; NMI
VEC_EC:	.word	0		; RESET
VEC_EE:	.word	0		; IRQ
;----------------------------------------------------------------------
; emulation mode vectors
;----------------------------------------------------------------------
VEC_F0:	.word	0		; -
VEC_F2:	.word	0		; -
VEC_F4:	.word	0		; COP
VEC_F6:	.word	0		; -
VEC_F8:	.word	0		; ABORT
VEC_FA:	.word	0		; NMI
VEC_FC:	.word	_start		; RES (entry point)
VEC_FE:	.word	0		; IRQ/BRK
;----------------------------------------------------------------------
