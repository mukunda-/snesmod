/*
 * Copyright 2009 Mukunda Johnson (mukunda.com)
 * Go version 2025
 * This file is part of SNESMOD - gh.mukunda.com/snesmod
 */

package main

import (
	"flag"
	"fmt"
	"os"
)

const usage = `SNESMOD (C) 2009 Mukunda Johnson (mukunda.com)

Usage: smconv [options] [input]

Options:
  -s, --soundbank  Soundbank creation mode.
                   (Can specify multiple files with soundbank mode.)
                   (Otherwise specify only one file for SPC creation.)
                   (Default is SPC creation mode)
  -o, --output     Specify output file or file base.
                   (Specify SPC file for -b option)
                   (Specify filename base for soundbank creation)
                   (Required for soundbank mode)
  -h, --hirom      Use HIROM mapping mode for soundbank.
  -v, --verbose    Enable verbose output.
  --help           Show Help

Typical options to create soundbank for project:
  smconv -s -o build/soundbank -h input1.it input2.it

And for IT->SPC:
  smconv input.it

TIP: use -v to view how much RAM the modules will use.`

type Config struct {
	SoundbankMode bool
	OutputFile    string
	HiROM         bool
	VerboseMode   bool
	InputFiles    []string
}

func parseArgs() (*Config, error) {
	cfg := &Config{}

	// Define flags
	flag.BoolVar(&cfg.SoundbankMode, "s", false, "Soundbank creation mode")
	flag.StringVar(&cfg.OutputFile, "o", "", "Output file")
	flag.BoolVar(&cfg.HiROM, "h", false, "Use HIROM mapping")
	flag.BoolVar(&cfg.VerboseMode, "v", false, "Verbose output")

	// Custom usage
	flag.Usage = func() {
		fmt.Println(usage)
	}

	flag.Parse()

	// Get remaining arguments as input files
	cfg.InputFiles = flag.Args()

	// Validate arguments
	if len(cfg.InputFiles) == 0 {
		return nil, fmt.Errorf("no input files specified")
	}

	if cfg.SoundbankMode && cfg.OutputFile == "" {
		return nil, fmt.Errorf("output file (-o) is required for soundbank mode")
	}

	return cfg, nil
}

func main() {
	cfg, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	if cfg.VerboseMode {
		fmt.Println("Loading modules...")
	}

	// TODO: Implement module loading and conversion
	// This will be implemented in separate packages for:
	// - IT file loading (itloader)
	// - BRR conversion (brr)
	// - SPC conversion (it2spc)

	if cfg.VerboseMode {
		fmt.Println("Starting conversion...")
	}

	// TODO: Implement conversion process
	// If soundbank mode:
	//   - Load all input files
	//   - Convert to soundbank format
	//   - Export data and header files
	// Else:
	//   - Load single input file
	//   - Convert to SPC format
	//   - Export SPC file
}
