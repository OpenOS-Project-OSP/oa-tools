package engine

import (
	"coa/src/internal/distro"
	"coa/src/internal/pilot"
	"fmt"
	"path/filepath"
)

func generateRemasterPlan(d *distro.Distro, mode string, workPath string) FlightPlan {
	plan := FlightPlan{
		PathLiveFs:      workPath,
		Mode:            mode,
		Family:          d.FamilyID,
		BootloadersPath: BootloaderRoot,
	}

	// =================================================================
	// FASE 1: THE BRAIN (Operazioni Interne - Chroot)
	// Leggiamo il profilo YAML tramite il pacchetto pilot
	// =================================================================
	profile := pilot.GetBrainProfile(d, mode, workPath)

	// Chiamiamo esplicitamente i task nell'ordine corretto, senza hardcoding!
	// 1.1 Iniezione identità (Pulizia, utenti, skel)
	appendTaskActions(&plan, profile, "identity")

	// 1.2 Rigenerazione dell'Initramfs
	appendTaskActions(&plan, profile, "initrd")

	// =================================================================
	// FASE 2: LIVE STRUCT (Operazioni Esterne - Host)
	// =================================================================
	kernelSrc := filepath.Join(workPath, "liveroot", "vmlinuz")
	initrdSrc := filepath.Join(workPath, "liveroot", "initrd.img")

	lsActions, err := generateLiveStruct(workPath, kernelSrc, initrdSrc)
	if err == nil {
		plan.Plan = append(plan.Plan, lsActions...)
	}

	// =================================================================
	// FASE 3: BOOTLOADERS
	// =================================================================
	blActions, err := LiveBootloader(workPath)
	if err == nil {
		plan.Plan = append(plan.Plan, blActions...)
	}

	// ---> NOVITÀ: Generazione dei Menu di Avvio! <---
	fmt.Println("\033[1;36m[coa]\033[0m Scrittura dei menu GRUB e ISOLINUX...")
	err = GenerateBootMenus(workPath, d, profile, "OA_LIVE")
	if err != nil {

		fmt.Printf("\033[1;33m[WARNING]\033[0m Impossibile scrivere i menu di boot: %v\n", err)
	}

	// =================================================================
	// FASE 4: SQUASHFS
	// =================================================================
	excludeFilePath := generateExcludeList(mode)
	sqActions, err := generateSquashfs(workPath, "xz", excludeFilePath)
	if err == nil {
		plan.Plan = append(plan.Plan, sqActions...)
	}

	// =================================================================
	// FASE 5: XORRISO
	// =================================================================
	isoName := getIsoName(d)
	isoActions, err := GenerateIso(workPath, isoName, "OA_LIVE")
	if err == nil {
		plan.Plan = append(plan.Plan, isoActions...)
	}

	// =================================================================
	// FASE 6: CLEANUP
	// =================================================================
	plan.Plan = append(plan.Plan, Action{
		Command: "oa_remaster_cleanup",
		Info:    "Smontaggio filesystem virtuali e pulizia finale",
	})

	return plan
}

// appendTaskActions cerca un task specifico nel profilo e accoda i suoi comandi al piano
func appendTaskActions(plan *FlightPlan, profile *pilot.BrainProfile, taskName string) {
	for _, t := range profile.Tasks {
		if t.Name == taskName {
			for _, cmd := range t.Commands {
				plan.Plan = append(plan.Plan, Action{
					Command:    "oa_shell",
					Info:       t.Description,
					RunCommand: cmd,
					Chroot:     t.Chroot,
				})
			}
			return // Task trovato e accodato, usciamo dal ciclo
		}
	}
}
