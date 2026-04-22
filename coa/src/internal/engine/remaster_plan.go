package engine

import (
	"coa/src/internal/distro"
	"coa/src/internal/pilot"
	"fmt"
	"os"
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

	// =================================================================
	// IDENTITÀ E UTENTI (Prima di chiudere lo squashfs)
	// =================================================================
	identityActions := GenerateIdentityActions(mode, profile)
	plan.Plan = append(plan.Plan, identityActions...)

	// ---> Generazione Initrd dinamica <---
	initrdActions, err := GenerateInitrd(workPath, profile)
	if err != nil {
		fmt.Printf("\033[1;31m[coa-FATAL]\033[0m Errore durante la preparazione dell'Initrd: %v\n", err)
		os.Exit(1)
	}

	// Aggiungiamo l'azione di rigenerazione initrd al piano
	plan.Plan = append(plan.Plan, initrdActions...)

	// =================================================================
	// FASE 2: LIVE STRUCT (Operazioni Esterne - Host)
	// =================================================================
	// Cerchiamo dinamicamente il Kernel in base alla distro!
	kernelSrc, initrdSrc, err := FindKernelAndInitrd(workPath, d.FamilyID)
	if err != nil {
		fmt.Printf("\033[1;31m[coa-FATAL]\033[0m %v\n", err)
		os.Exit(1)
	}

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
	// LAYOUT E SYMLINK DELLA ISO (Es. la cartella arch/x86_64/)
	// =================================================================
	layoutActions := GenerateLayoutActions(workPath, profile)
	plan.Plan = append(plan.Plan, layoutActions...)

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
		Command: "oa_umount",
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
