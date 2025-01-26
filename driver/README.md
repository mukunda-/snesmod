## sm-spc

This is the SPC-700 driver used by SNESMOD. Also included here is the API that the SNES
can use to communicate with the driver.

### Build Requirements

* Make
* Python3
* SNESKIT must be installed. $(SNESKIT) must be added to the path.
* Telemark Assembler (TASM).

TASM is shareware and not included. Copy the TASM binary into /tasm for building.

### Building
```
# Compile the driver.
make

# Add the driver and include files to the SNESKIT directories.
make install
```
