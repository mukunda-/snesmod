.SECONDARY:
.SUFFIXES:

ifeq ($(OS),Windows_NT)
PYTHON := py
else
PYTHON := python3
endif

#-----------------------------------------------------------------------------------------
.PHONY: all clean install
#-----------------------------------------------------------------------------------------

all: build/sm_spc_image.asm

#-----------------------------------------------------------------------------------------
clean:
	@rm -rf build

#-----------------------------------------------------------------------------------------
build/sm_spc.bin: spc/sm_spc.asm
	@mkdir -p build
	tasm/tasm -07 -b -l spc/sm_spc.asm build/sm_spc.bin build/sm_spc.lst

#-----------------------------------------------------------------------------------------
build/sm_spc_image.asm: build/sm_spc.bin
	$(PYTHON) script/bin2ca.py $< $@ --label SNESMOD_SPC --segment RODATA --bytesperline 32

#-----------------------------------------------------------------------------------------
install:
ifeq ($(strip $(SNESKIT)),)
	$(error SNESKIT path not found, please add it to your environment)
endif
	cp snes/snesmod_dev.asm $(SNESKIT)/source/snesmod.asm
	cp build/sm_spc_image.asm $(SNESKIT)/source/sm_spc.asm
	cp include/snesmod.inc $(SNESKIT)/include/snesmod.inc
