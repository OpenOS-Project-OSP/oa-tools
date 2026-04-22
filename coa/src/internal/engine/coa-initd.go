package engine

import (
	"coa/src/internal/pilot"
	"fmt"
	"os"
	"path/filepath"
)

// GenerateInitrdActions prepara l'ambiente e restituisce le azioni per rigenerare l'initramfs
func GenerateInitrd(workPath string, profile *pilot.BrainProfile) ([]Action, error) {
	var actions []Action

	// Se lo YAML non prevede nulla per l'initrd, restituiamo un array vuoto
	if profile.Areas.Initrd.Command == "" {
		return actions, nil
	}

	fmt.Println("\033[1;36m[coa]\033[0m Preparazione ambiente Initrd dal profilo YAML...")

	liveroot := filepath.Join(workPath, "liveroot")

	// =================================================================
	// 1. INIEZIONE DEI FILE DI SETUP (es. mkinitcpio.conf)
	// Li scriviamo fisicamente dal lato host prima di lanciare il comando
	// =================================================================
	for targetPath, content := range profile.Areas.Initrd.Files {
		// Puliamo il percorso. Se inizia con "/", lo rimuoviamo per evitare
		// che filepath.Join scavalchi la liveroot e scriva nell'host vero.
		cleanPath := filepath.Clean(targetPath)
		if len(cleanPath) > 0 && cleanPath[0] == '/' {
			cleanPath = cleanPath[1:]
		}

		fullPath := filepath.Join(liveroot, cleanPath)

		// Creiamo le cartelle genitrici se non esistono
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return nil, fmt.Errorf("impossibile creare le directory per %s: %w", fullPath, err)
		}

		// Scriviamo il file
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("impossibile scrivere il file %s: %w", fullPath, err)
		}
	}

	// =================================================================
	// 2. GENERAZIONE DELL'AZIONE CHROOT
	// =================================================================
	actions = append(actions, Action{
		Command:    "oa_shell",
		Info:       "Generazione Initramfs custom",
		RunCommand: profile.Areas.Initrd.Command,
		Chroot:     true, // Fondamentale: deve girare DENTRO la liveroot
	})

	return actions, nil
}
