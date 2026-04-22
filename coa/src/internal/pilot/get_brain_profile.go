package pilot

import (
	"coa/src/internal/distro"
)

func GetBrainProfile(d *distro.Distro, mode string, workPath string) *BrainProfile {
	// 1. Inizializziamo il Cervello con le mappe pronte
	profile := &BrainProfile{
		Tasks: []Task{},
		Areas: Areas{
			Initrd: InitrdConfig{Files: make(map[string]string)},
			Layout: LayoutConfig{Links: make(map[string]string)},
		},
	}

	// =================================================================
	// FASE 1: CARICAMENTO DATI PURI (Riempimento delle Aree)
	// Leggiamo i file YAML e li iniettiamo nelle strutture strongly-typed
	// =================================================================

	_ = readAreaConfig(d.FamilyID, "identity", &profile.Areas.Identity)
	_ = readAreaConfig(d.FamilyID, "initrd", &profile.Areas.Initrd)
	_ = readAreaConfig(d.FamilyID, "boot", &profile.Areas.Boot)     // Niente più task fittizi!
	_ = readAreaConfig(d.FamilyID, "layout", &profile.Areas.Layout) // I link sono pronti all'uso!

	return profile
}
