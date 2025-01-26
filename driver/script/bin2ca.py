#!/usr/bin/env python3
#*****************************************************************************************
# SNESMOD
# (C) 2025 Mukunda Johnson (mukunda.com)
# Licensed under MIT
#*****************************************************************************************
import argparse

#-----------------------------------------------------------------------------------------
def bin2ca():
   parser = argparse.ArgumentParser(prog="bin2ca.py", description='Convert binary files to assembly files for ca65.')
   parser.add_argument('input_file', type=str, help='The binary file to convert.')
   parser.add_argument('output_file', type=str, help='The output assembly file.')
   parser.add_argument('--label', type=str, help='A label to emit for the assembly file. LABEL and LABEL_end will be produced if this argument is given.')
   parser.add_argument('--segment', type=str, help='The segment to emit for the assembly file.')
   parser.add_argument('--bytesperline', type=int, default=64, help='The number of bytes to emit per line.')
   args = parser.parse_args()

   with open(args.input_file, "rb") as f:
      data = f.read()
   
   with open(args.output_file, "w") as f:
      f.write(f"""
; bin2ca.py converted binary data
; total size: {len(data)} bytes""".strip())
      f.write("\n\n")

      if args.label:
         f.write(f"\t.global {args.label}, {args.label}_end\n")

      if args.segment:
         f.write(f"\t.segment \"{args.segment}\"\n")
      
      f.write("\n")

      if args.label:
         f.write(f"{args.label}:\n")
      
      bpl = args.bytesperline
      for i in range(0, len(data), bpl):
         f.write("\t.byte ")
         f.write(",".join([f"${byte:02x}" for byte in data[i:i+bpl]]))
         f.write("\n")
      
      if args.label:
         f.write(f"{args.label}_end:\n")

if __name__ == "__main__":
   bin2ca()
