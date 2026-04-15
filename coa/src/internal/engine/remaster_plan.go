// Copyright 2026 Piero Proietti <piero.proietti@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package engine

import (
	"coa/src/internal/distro"
	"coa/src/internal/pilot"
	"fmt"
)

// GeneratePlan costruisce il piano di volo dinamico basato sul Cervello (pilot)
func GeneratePlan(d *distro.Distro, mode string, workPath string) FlightPlan {
	// 1. Recuperiamo i dati dal Pilota (il Cervello Modulare)
	task := pilot.GetInitrdTask(d.FamilyID)

	// Valori di default (fallback) se il pilota non trova nulla
	bootParams := "boot=live components"
	adminGroup := "sudo"
	userGroups := []string{"audio", "video", "autologin"}

	if task != nil {
		if task.Remaster.BootParams != "" {
			bootParams = task.Remaster.BootParams
		}
		if task.Remaster.AdminGroup != "" {
			adminGroup = task.Remaster.AdminGroup
		}
		if len(task.Remaster.UserGroups) > 0 {
			userGroups = task.Remaster.UserGroups
		}
	}

	// Uniamo i gruppi utente con il gruppo amministrativo
	allGroups := append(userGroups, adminGroup)

	plan := FlightPlan{
		PathLiveFs:      workPath,
		Mode:            mode,
		Family:          d.FamilyID,
		BootloadersPath: BootloaderRoot,
	}

	// 2. Gestione Identità (Solo in modalità standard)
	if mode == "standard" {
		plan.Users = []UserConfig{
			{
				Login:    "live",
				Password: "$6$wM.wY0QtatvbQMHZ$QtIKXSpIsp2Sk57.Ny.JHk7hWDu.lxPtUYaTOiBnP4WBG5KS6JpUlpXj2kcSaaMje7fr01uiGmxZhE8kfZRqv.",
				Gecos:    "live,,,",
				Home:     "/home/live",
				Shell:    "/bin/bash",
				Groups:   allGroups,
			},
		}

		plan.Plan = append(plan.Plan, Action{Command: "oa_remaster_users"})

		// Configurazione sudoers dinamica basata sul gruppo admin del Cervello
		sudoersCmd := fmt.Sprintf("mkdir -p /etc/sudoers.d && echo '%%%s ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/00-oa-live && chmod 0440 /etc/sudoers.d/00-oa-live", adminGroup)
		plan.Plan = append(plan.Plan, Action{
			Command:    "oa_sys_shell",
			RunCommand: sudoersCmd,
			Chroot:     true,
		})
	} else {
		plan.Users = []UserConfig{}
		plan.Plan = append(plan.Plan, Action{Command: "oa_remaster_users"})
	}

	// 3. Gestione Initrd (Pilotaggio dinamico)
	if task != nil {
		// Scrittura file di configurazione (es. coa-mkinitcpio.conf)
		for path, content := range task.SetupFiles {
			writeCmd := fmt.Sprintf("mkdir -p $(dirname %s) && echo -e '%s' > %s", path, content, path)
			plan.Plan = append(plan.Plan, Action{
				Command:    "oa_sys_shell",
				RunCommand: writeCmd,
				Chroot:     true,
			})
		}

		// Esecuzione comando di rigenerazione (Protezione contro comandi vuoti!)
		if task.Command != "" {
			plan.Plan = append(plan.Plan, Action{
				Command:    "oa_sys_shell",
				RunCommand: task.Command,
				Chroot:     true,
			})
		}

		// Pulizia file temporanei di configurazione
		for path := range task.SetupFiles {
			plan.Plan = append(plan.Plan, Action{
				Command:    "oa_sys_shell",
				RunCommand: fmt.Sprintf("rm -f %s", path),
				Chroot:     true,
			})
		}
	}

	// 4. Struttura ISO e Bootloaders
	excludeFilePath := generateExcludeList(mode)

	plan.Plan = append(plan.Plan,
		Action{Command: "oa_remaster_livestruct"},
		Action{Command: "oa_remaster_isolinux", BootParams: bootParams},
		Action{Command: "oa_remaster_uefi", BootParams: bootParams},
	)

	// 5. IL PONTE: Creazione Link Simbolici (Layout)
	// Essenziale per Arch Linux (es. airootfs.sfs -> filesystem.squashfs)
	if task != nil && len(task.Remaster.IsoLinks) > 0 {
		for dst, src := range task.Remaster.IsoLinks {
			linkCmd := fmt.Sprintf("mkdir -p $(dirname %s/iso/%s) && ln -sf %s %s/iso/%s",
				workPath, dst, src, workPath, dst)
			plan.Plan = append(plan.Plan, Action{
				Command:    "oa_sys_shell",
				RunCommand: linkCmd,
				Chroot:     false, // Azione fuori dal chroot, agisce sulla directory di lavoro
			})
		}
	}

	// 6. Chiusura: Squashfs e ISO
	plan.Plan = append(plan.Plan,
		Action{
			Command:     "oa_remaster_squash",
			ExcludeList: excludeFilePath,
		},
	)

	if mode == "crypted" {
		plan.Plan = append(plan.Plan, Action{
			Command:         "oa_remaster_crypted",
			CryptedPassword: "evolution",
		})
	}

	isoName := getIsoName(d)
	plan.Plan = append(plan.Plan, Action{
		Command:   "oa_remaster_iso",
		VolID:     "OA_LIVE",
		OutputISO: isoName,
	})

	plan.Plan = append(plan.Plan, Action{Command: "oa_remaster_cleanup"})

	return plan
}
