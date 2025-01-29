
# Check if Go is installed
ifeq (, $(shell which go))
$(error "No go toolchain found in PATH, please install Go")
endif

#-----------------------------------------------------------------------------------------
.PHONY: all clean sneskit_install

#-----------------------------------------------------------------------------------------
all:
	go build -o build/

#-----------------------------------------------------------------------------------------
clean:
	rm -rf build

#-----------------------------------------------------------------------------------------
sneskit_install: all
ifeq ($(strip $(SNESKIT)),)
	$(error SNESKIT not found, please add SNESKIT=<path to sneskit> to your environment)
endif
	cp build/smconv* $(SNESKIT)/bin/
