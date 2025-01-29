#-------------------------------------------------------------------------
# SNES Memory Specification
#-------------------------------------------------------------------------
# By default this is for 20H LoROM mode. Update the sections as necessary
# for HiROM mode.

#*************************************************************************
MEMORY {
#*************************************************************************

#-------------------------------------------------------------------------
# ZROM is used for program code (32k minus cartridge header)
# It's actually at bank 80h (the absolute address is $808000), but we use
# 0 to avoid ASM errors (it is addressed with the "zero page"
# instructions).
#	ZROM: start = $008000, size = $7FB0, fill = yes;

#-------------------------------------------------------------------------
# For HiROM, remove ZROM and use these.
# HROM is used for extended program code (32k) - not used normally.
# CROM is common program code (32k) shared across all banks.
# These are at $40, but we use $00 to avoid ASM errors.
	HROM: start = $0000, size = $8000, fill = yes;
	CROM: start = $8000, size = $7FB0, fill = yes;

#-------------------------------------------------------------------------
# CHEAD is the cartridge header space.
	CHEAD: start = $00FFB0, size = 80, fill = yes;

#-------------------------------------------------------------------------
# Insert ROM sections for program data.
# Note that some tools will say your cart is corrupt if it's smaller than
# 1 mbit, hence 3 ROM sections here by default.
#	ROM1: start = $818000, size = $8000, fill = yes;
#	ROM2: start = $820000, size = $8000, fill = yes;
#	ROM3: start = $828000, size = $8000, fill = yes;

# For HiROM, use these instead of the above (size 64k, bank $40+)
	ROM1: start = $410000, size = $10000, fill = yes;
	ROM2: start = $420000, size = $20000, fill = yes;

#-------------------------------------------------------------------------
# 128K memory
	ZPAGE:  start = $0000, size = 256;      # 8k LoRAM
	LORAM:  start = $0100, size = 7936;     #
	HIRAM1: start = $7E2000, size = 57344;  # 56k HiRAM1
	HIRAM2: start = $7F0000, size = 65536;  # 64k HiRAM2
}

#*************************************************************************
SEGMENTS {
#*************************************************************************

#-------------------------------------------------------------------------
# Standard segments
#-------------------------------------------------------------------------
	CODE:     load = CROM, type = ro;       # 32k Code
	HEADER:   load = CHEAD, type = ro;      # Cartridge header
	ZEROPAGE: load = ZPAGE, type = zp;      # Zero page
	BSS:      load = LORAM, type = bss;     # 8k (minus zeropage) BSS
	RODATA:	  load = ROM1, type = ro;       # Read only/Data

#-------------------------------------------------------------------------
# Note: when using HRAM or HRAM2 or HDATA you must use FAR addressing:
#
# .segment "HRAM": far
#
# The segment word 'data' can't be used because I can't figure out how to
# override the addressing mode.
	HRAM:  load = HIRAM1, type = bss;      # 56K+64K HRAM
	HRAM2: load = HIRAM2, type = bss;      # 

#-------------------------------------------------------------------------
# HDATA is loaded to memory.
# Default: Load in ROM1, Run in 64k HIRAM2
	HDATA: load = ROM1, run = HIRAM2, type = rw, define = yes;

#-------------------------------------------------------------------------
# Custom segments - insert as needed.
	GRAPHICS:  load = ROM1, type = ro;
	SOUNDBANK: load = ROM2, type = ro;

}
