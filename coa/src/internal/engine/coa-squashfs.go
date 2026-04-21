package engine

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// generateSquashfs crea l'azione per comprimere la liveroot nel filesystem.squashfs
func generateSquashfs(workPath string, compType string, excludeFile string) ([]Action, error) {
	var actions []Action

	// 1. Definiamo i percorsi esatti usando workPath
	liveroot := filepath.Join(workPath, "liveroot")
	squashDest := filepath.Join(workPath, "isodir", "live", "filesystem.squashfs")

	// 2. Interroghiamo il sistema per sapere quanti thread (core logici) abbiamo
	cores := runtime.NumCPU()

	// 3. Creazione del comando per mksquashfs
	// -comp zstd: Usa l'algoritmo Zstandard
	// -Xcompression-level 3: Livello 3 (il default perfetto tra velocità e compressione)
	// -processors %d: Forza mksquashfs a usare tutti i core rilevati da Go
	cmd := fmt.Sprintf(
		"mksquashfs %s %s -comp zstd -Xcompression-level 3 -b 1M -processors %d -noappend -wildcards -ef %s",
		liveroot, squashDest, cores, excludeFile,
	)

	// 4. Aggiungiamo l'azione da passare al motore C (usando il comando unificato oa_shell)
	actions = append(actions, Action{
		Command:    "oa_shell",
		Info:       fmt.Sprintf("Compressione SquashFS (zstd livello 3, %d thread) - Decollo in corso...", cores),
		RunCommand: cmd,
		Chroot:     false,
	})

	return actions, nil
}
