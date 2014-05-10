.SECONDARY:

#---------------------------------------------------------------------------------
.SUFFIXES:
#---------------------------------------------------------------------------------

ifeq ($(strip $(SNESKIT)),)
$(error SNESKIT path not found, please add it to your environment)
endif

include	$(SNESKIT)/snes_rules

#---------------------------------------------------------------------------------
# TARGET is the name of the output
# BUILD is the directory where object files & intermediate files will be placed
# SOURCES is a list of directories containing source code
# INCLUDES is a list of directories containing extra header files
#---------------------------------------------------------------------------------
TARGET		:=	test_driver
BUILD		:=	build
SOURCES		:=	snes test
INCLUDES	:=	include
GRAPHICS	:=	graphics
AUDIO		:=	audio
SOUND		:=	sound

TASM		:=	/c/tasm301/TASM

MAPMODE		:=	hirom

#---------------------------------------------------------------------------------
# options for code generation
#---------------------------------------------------------------------------------
STARTFILES	:= snes_init snes_decompress snes_joypad snes_zvars
ASFLAGS		:=	--cpu 65816 -s -D HIROM
SMCONVFL	:= --hirom --soundbank --output soundbank

TASMFLAGS	:=	-07 -b -l
 
#---------------------------------------------------------------------------------
ifneq ($(BUILD),$(notdir $(CURDIR)))
#---------------------------------------------------------------------------------

export INCLUDE	:=	$(foreach dir,$(INCLUDES),-I $(CURDIR)/$(dir)) -I $(CURDIR)/$(BUILD) -I $(SNESKIT)/include -I $(CURDIR)

export OUTPUT	:=	$(CURDIR)/$(TARGET)
export	SPCBIN	:=	$(CURDIR)/sm_spc.obj

export LDFLAGS	:=	-C $(CURDIR)/config.ld -o $(OUTPUT).smc

export VPATH	:=	$(foreach dir,$(SOURCES),$(CURDIR)/$(dir)) \
					$(foreach dir,$(GRAPHICS),$(CURDIR)/$(dir)) \
					$(foreach dir,$(AUDIO),$(CURDIR)/$(dir)) \
					$(SNESKIT)/source

SFILES		:=	$(STARTFILES:=.asm) sm_spc_image.asm $(foreach dir,$(SOURCES),$(notdir $(wildcard $(dir)/*.asm))) 
PNGFILES	:=	$(foreach dir,$(GRAPHICS),$(notdir $(wildcard $(dir)/*.png)))

export AUDIOFILES :=	$(foreach dir,$(AUDIO),$(notdir $(wildcard $(dir)/*.*)))
export SOUNDFILES :=	$(foreach dir,$(SOUND),$(notdir $(wildcard $(dir)/*.wav)))

ifneq ($(AUDIOFILES),)
SFILES		:=	soundbank.asm $(SFILES)
#export SBXML	:=	$(CURDIR)/audio.xml
endif

export OFILES	:=	$(PNGFILES:.png=.o) $(SFILES:.asm=.o)
 
.PHONY: $(BUILD) clean install
 
#---------------------------------------------------------------------------------
$(BUILD):
	@[ -d $@ ] || mkdir -p $@
	@make --no-print-directory -C $(BUILD) -f $(CURDIR)/Makefile
 
all	: $(BUILD)
#---------------------------------------------------------------------------------
clean:
	@echo clean ...
	@rm -fr $(BUILD) $(TARGET).smc $(TARGET).map $(SPCBIN) $(SPCBIN:.obj=.lst)
	
install:
	cp snes/snesmod_dev.asm $(SNESKIT)/source/snesmod.asm
	cp build/sm_spc_image.asm $(SNESKIT)/source/sm_spc.asm
	cp include/snesmod.inc $(SNESKIT)/include/snesmod.inc

#---------------------------------------------------------------------------------
else
 
#---------------------------------------------------------------------------------
# main targets
#---------------------------------------------------------------------------------
$(OUTPUT).smc	: $(BRRFILES) $(OFILES)
	@echo Linking...
	@$(LD) $(LDFLAGS) $(OFILES) -o $@ -m $(@:.smc=.map)
	@echo ROM Checkup...
	@$(SNESCHK) $(MAPMODE) $@
	@echo Okay! :: $(notdir $@)

	
%.o : %.asm
%.o : %.asm $(SPCBIN)
	@echo Assembling ... $(notdir $<)
	@$(AS) $(ASFLAGS) $(INCLUDE) $< -o $@
	
sm_spc_image.asm : $(SPCBIN)
	$(SNESTOOLS)/bin2ca $< RODATA SNESMOD_SPC $@

$(SPCBIN) : ../spc/sm_spc.asm
	cp ../tasm07.tab tasm07.tab
	$(TASM) $(TASMFLAGS) $< ../SM_SPC.OBJ ../SM_SPC.LST
	../bin2h ../SM_SPC.OBJ ../../convert/include/spc_program.h spc_program
	
%.asm : %.png %.grit 
	@echo Converting ... $(notdir $<)
	@$(SNESGRIT) $< -fta -fh!
	
#---------------------------------------------------------------------------------
# rule for converting audio files
#---------------------------------------------------------------------------------
soundbank.asm : $(AUDIOFILES)
	@echo Compiling Soundbank ...
	$(SMCONV) $(SMCONVFL) $^ -v
	


#---------------------------------------------------------------------------------------
endif
#---------------------------------------------------------------------------------------
