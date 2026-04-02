package main

import (
	"fmt"
)

// Action rappresenta un singolo blocco "command" nell'array "plan"
type Action struct {
	Command         string `json:"command"`
	VolID           string `json:"volid,omitempty"`
	OutputISO       string `json:"output_iso,omitempty"`
	CryptedPassword string `json:"crypted_password,omitempty"`
}

// FlightPlan rappresenta l'intero piano da passare a oa
type FlightPlan struct {
	PathLiveFs      string   `json:"pathLiveFs"`
	Mode            string   `json:"mode"`
	InitrdCmd       string   `json:"initrd_cmd"`
	BootloadersPath string   `json:"bootloaders_path"`
	Plan            []Action `json:"plan"`
}

// GeneratePlan è il cervello che costruisce la pipeline dinamica
func GeneratePlan(d *Distro, mode string, workPath string) FlightPlan {
	plan := FlightPlan{
		PathLiveFs: workPath,
		Mode:       mode,
	}

	// 1. Astrazione Initramfs (Il Terzo Pilastro della tua Universal Strategy)
	switch d.FamilyID {
	case "debian":
		plan.InitrdCmd = "mkinitramfs -o {{out}} {{ver}}"
	case "fedora", "opensuse":
		plan.InitrdCmd = "dracut --nomadas --force {{out}} {{ver}}"
	case "archlinux":
		// mkinitcpio richiede configurazioni specifiche per chroot, ma il template è questo
		plan.InitrdCmd = "mkinitcpio -g {{out}} -k {{ver}}"
	default:
		plan.InitrdCmd = "mkinitramfs -o {{out}} {{ver}}" // Fallback di sicurezza
	}

	// 2. Astrazione Bootloader (Il Primo Pilastro)
	if d.FamilyID != "debian" {
		// Se NON siamo su Debian, diciamo a oa di usare i binari passepartout pre-estratti
		plan.BootloadersPath = "/usr/share/cova/bootloaders"
	} else {
		plan.BootloadersPath = "" // Usa quelli di sistema
	}

	// 3. Assemblaggio dinamico della catena di montaggio
	plan.Plan = []Action{
		{Command: "action_prepare"},
		{Command: "action_users"}, // oa sa già come gestirlo in base al "mode"
		{Command: "action_initrd"},
		{Command: "action_livestruct"},
		{Command: "action_isolinux"},
		{Command: "action_uefi"},
		{Command: "action_squash"},
	}

	// Inserzione modulare: se l'utente vuole l'ISO cifrata, aggiungiamo l'azione in mezzo!
	if mode == "crypted" {
		plan.Plan = append(plan.Plan, Action{
			Command:         "action_crypted",
			CryptedPassword: "evolution", // Qui potremmo passarla da linea di comando in futuro
		})
	}

	// 4. Generazione ISO e chiusura
	// Costruiamo un nome file parlante e dinamico!
	isoName := fmt.Sprintf("egg-of_%s-%s-oa_amd64.iso", d.DistroID, d.CodenameID)
	
	plan.Plan = append(plan.Plan, Action{
		Command:   "action_iso",
		VolID:     "OA_LIVE",
		OutputISO: isoName,
	})
	
	plan.Plan = append(plan.Plan, Action{Command: "action_cleanup"})

	return plan
}