// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

/*
smconv is a part of SNESMOD. It converts music files to SPC format or a "soundbank" to be
included in a game.
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"go.mukunda.com/modlib"
	"go.mukunda.com/snesmod/smconv/clog"
	"go.mukunda.com/snesmod/smconv/smconv"
)

const shortUsage = `Usage: smconv [options] input
Use --help for more info.`

const usage = `SNESMOD (C) 2025 Mukunda Johnson (mukunda.com)

Usage: smconv [options] input

Options
-------

-s, --soundbank
   Soundbank creation mode. Multiple input files can be
   specified with this option. If this option is not used,
   it is "SPC" creation mode and one input file can be
   specified to be converted to SPC.
  
-o, --output
   For soundbank creation mode, this is the base filename.
   Required for soundbank creation.
   For SPC creation, this is the SPC filename to create.
   For SPC, default is the input with the extension changed
   to .spc.
	
-h, --hirom
   Use HIROM mapping mode for the soundbank.

-v, --verbose
   Enable verbose output.

--help
   Show Help

Example to create a soundbank for a project:
  smconv -s -o build/soundbank -h input1.it input2.it

Example to convert IT to SPC:
  smconv input.it`

type programArgs struct {
	Help          bool
	SoundbankMode bool
	OutputFile    string
	HiRom         bool
	VerboseMode   bool
	InputFiles    []string
}

func parseArgs(argStrings []string) (*programArgs, error) {
	cfg := &programArgs{}

	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Define flags
	flags.BoolVar(&cfg.SoundbankMode, "s", false, "Soundbank creation mode")
	flags.BoolVar(&cfg.SoundbankMode, "soundbank", false, "Soundbank creation mode")
	flags.StringVar(&cfg.OutputFile, "o", "", "Output file")
	flags.StringVar(&cfg.OutputFile, "output", "", "Output file")
	flags.BoolVar(&cfg.HiRom, "h", false, "Use HIROM mapping (larger banks)")
	flags.BoolVar(&cfg.HiRom, "hirom", false, "Use HIROM mapping (larger banks)")
	flags.BoolVar(&cfg.VerboseMode, "v", false, "Verbose output")
	flags.BoolVar(&cfg.VerboseMode, "verbose", false, "Verbose output")
	flags.BoolVar(&cfg.Help, "?", false, "Show help")
	flags.BoolVar(&cfg.Help, "help", false, "Show help")

	err := flags.Parse(argStrings)
	if err != nil {
		return nil, err
	}

	// Get remaining arguments as input files
	cfg.InputFiles = flags.Args()

	return cfg, nil
}

// Main entry point.
func smconvCli(args []string) int {
	cfg, err := parseArgs(args)

	if err != nil {
		clog.Errorf("%v", err)
		fmt.Fprintln(os.Stderr, shortUsage)
		return 1
	}

	if cfg.Help {
		fmt.Println(usage)
		return 0
	}

	// Validate arguments
	if len(cfg.InputFiles) == 0 {
		clog.Errorln("No input files specified.")
		return 1
	}

	if cfg.SoundbankMode && cfg.OutputFile == "" {
		clog.Errorln("Output file (-o) is required for soundbank mode.")
		return 1
	}

	bank := smconv.SoundBank{}
	if !cfg.SoundbankMode && len(cfg.InputFiles) != 1 {
		clog.Errorln("SPC conversion mod requires exactly one input file.")
		return 1
	}

	if cfg.SoundbankMode {
		clog.Infoln("Loading input files for soundbank.")
	} else {
		clog.Infoln("Loading input file.")
	}

	for _, inputFile := range cfg.InputFiles {
		clog.Infoln("Loading module:", inputFile)
		mod, err := modlib.LoadModule(inputFile)
		if err != nil {
			clog.Errorf("Error loading module %s: %v\n", inputFile, err)
			return 1
		}

		err = bank.AddModule(mod, inputFile)
		if err != nil {
			clog.Errorf("Error converting module %s: %v\n", inputFile, err)
			return 1
		}
	}

	if cfg.SoundbankMode {
		clog.Infoln("Exporting sound bank.")
		outputFile := strings.TrimSuffix(cfg.OutputFile, ".smbank")

		bank.Export(outputFile+".smbank", cfg.HiRom)
		bank.ExportAssembly(outputFile+".asm", outputFile+".smbank")
		bank.ExportAssemblyInclude(outputFile + ".inc")

	} else {
		clog.Infoln("Writing SPC file.")
		// Export to SPC
		err := bank.WriteSpcFile(cfg.OutputFile)
		if err != nil {
			clog.Errorf("Error writing SPC file: %v\n", err)
			return 1
		}
	}

	return 0
}

// Entry point forwards to our inner function so we can overwrite args during testing.
func main() {
	os.Exit(smconvCli(os.Args))
}
