package engine

import (
	"coa/src/internal/pilot"
	"fmt"
	"path/filepath"
)

// GenerateLayoutActions crea le azioni per i link simbolici della ISO
func GenerateLayoutActions(workPath string, profile *pilot.BrainProfile) []Action {
	var actions []Action

	// Se la mappa Links è vuota, non c'è nulla da fare
	if len(profile.Areas.Layout.Links) == 0 {
		return actions
	}

	for dst, src := range profile.Areas.Layout.Links {
		// Calcoliamo i percorsi sicuri lato host.
		// Assumiamo che la directory di destinazione sia "isodir" (come visto nei log)
		targetFile := filepath.Join(workPath, "isodir", dst)
		targetDir := filepath.Dir(targetFile)

		// Costruiamo il comando bash (eseguito sull'host, NON in chroot)
		linkCmd := fmt.Sprintf("mkdir -p %s && ln -sf %s %s", targetDir, src, targetFile)

		actions = append(actions, Action{
			Command:    "oa_shell",
			Info:       fmt.Sprintf("ISO Layout link: %s", dst),
			RunCommand: linkCmd,
			Chroot:     false,
		})
	}

	return actions
}
