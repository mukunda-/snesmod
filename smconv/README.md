## smconv

SNESMOD conversion tool

### Overview

This tool converts module files into SNESMOD soundbanks or SPC files. See
doc/snesmod_music.txt for restrictions and guidelines on making compatible music for the
SPC driver.

### Usage

Create a soundbank for a game.
```
smconv -s -o build/soundbank input1.it input2.it
```

Create an SPC file.
```
smconv -o song.spc song.it
```

See --help for more options.

### Additional notes

The soundbank is a continuous block of data that can span multiple ROM banks. The default
settings use LoROM mapping, meaning that addresses are mapped to 0x8000-0xFFFF range in
each bank. The -h flag enables HiROM mapping, where addresses are mapped to 0x0000-0xFFFF.