#!/usr/bin/env python3
# SNESMOD
# (C) 2025 Mukunda Johnson (mukunda.com)
# Licensed under MIT

# Frequency Look Up Table generator for the SPC driver.

print("LUT_FTAB:", end="")

for i in range(0,768):
    if (i % 16) == 0:
        print("\n\t.word ", end="")
    print(hex( int(round((1070.464*8) * 2**(i/768.0)) )).replace("0x","").upper().zfill(5) + "h", end="")
    if (i % 16) != 15:
        print(", ", end="")

print("")
print("")

