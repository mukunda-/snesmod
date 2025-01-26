// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package smconv

import (
	"regexp"
	"strings"

	"go.mukunda.com/modlib/common"
)

const (
	// Free SPC memory to be used by module data.
	kSpcRamSize = 58000

	// Base of module in SPC memory.
	kModuleBase = 0x1A00
)

type SourceIndex = uint16

type SoundBank struct {
	HiRom   bool
	Sources []*Source
	Modules []*SmModule
}

func (bank *SoundBank) AddModule(mod *common.Module, filename string) error {

	// usedSources indexes into the bank.Sources.
	// The sampleSourceMap indexes into usedSources, mapping samples -> used sources.

	// In other words, usedSources dictates what to load, and sampleSourceMap references
	// loaded sources.

	usedSources := []SourceIndex{}
	sampleSourceMap := []uint8{}

	for i := 0; i < len(mod.Samples); i++ {
		s, err := createSource(mod.Samples[i])
		if err != nil {
			return err
		}
		index := bank.AddSource(s)
		exists := false

		for j := 0; j < len(usedSources); j++ {
			if usedSources[j] == index {
				sampleSourceMap = append(sampleSourceMap, uint8(j))
				exists = true
				break
			}
		}

		if !exists {
			sampleSourceMap = append(sampleSourceMap, uint8(len(usedSources)))
			usedSources = append(usedSources, index)
		}
	}

	smMod := convertModule(mod, filename, usedSources, sampleSourceMap, bank.Sources)
	bank.Modules = append(bank.Modules, smMod)
	return nil
}

// Adds a source and returns the index of it. If a duplicate source exists, then the
// existing index is returned instead and nothing is added.
func (bank *SoundBank) AddSource(s *Source) SourceIndex {
	for i, es := range bank.Sources {
		if es.Hash == s.Hash {
			return SourceIndex(i)
		}
	}

	bank.Sources = append(bank.Sources, s)
	return SourceIndex(len(bank.Sources) - 1)
}

// Convert a file path into an ID for inserting into definition files.
// This strips the directory and extension, converts to uppercase, and replaces non
// alphanumeric with "_".
func pathToId(prefix string, path string) string {
	path = strings.ToUpper(path)
	path = strings.ReplaceAll(path, "\\", "/")

	slash := strings.LastIndex(path, "/")
	if slash != -1 {
		path = path[slash+1:]
	}

	dot := strings.Index(path, ".")
	if dot != -1 {
		path = path[:dot]
	}

	path = prefix + path

	stripper := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	path = stripper.ReplaceAllString(path, "_")

	return path
}
