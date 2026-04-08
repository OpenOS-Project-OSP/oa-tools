// Copyright 2026 Piero Proietti <piero.proietti@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// Action rappresenta un singolo blocco "command" nell'array "plan"
type Action struct {
	Command         string   `json:"command"`
	VolID           string   `json:"volid,omitempty"`
	OutputISO       string   `json:"output_iso,omitempty"`
	CryptedPassword string   `json:"crypted_password,omitempty"`
	RunCommand      string   `json:"run_command,omitempty"`
	ExcludeList     string   `json:"exclude_list,omitempty"` // <--- AGGIUNTO
	Args            []string `json:"args,omitempty"`
}

// UserConfig definisce la struttura per la creazione nativa dell'utente live [cite: 319, 321]
type UserConfig struct {
	Login    string   `json:"login"`
	Password string   `json:"password"`
	Gecos    string   `json:"gecos"`
	Home     string   `json:"home"`
	Shell    string   `json:"shell"`
	Groups   []string `json:"groups"`
}

// FlightPlan è l'oggetto JSON principale inviato al motore oa [cite: 32, 146, 281]
type FlightPlan struct {
	PathLiveFs      string       `json:"pathLiveFs"`
	Mode            string       `json:"mode"`
	InitrdCmd       string       `json:"initrd_cmd"`
	BootloadersPath string       `json:"bootloaders_path"`
	Users           []UserConfig `json:"users"` // Array globale degli utenti [cite: 32]
	Plan            []Action     `json:"plan"`
}

// generateExcludeList crea il file .list dinamico per mksquashfs
func generateExcludeList(mode string) string {
	outPath := "/tmp/coa/excludes.list"
	var excludes []string

	// 1. Esclusioni Base (Pulizia della ISO)
	excludes = append(excludes,
		"boot/efi/EFI",
		"boot/loader/entries/",
		"etc/fstab",
		"var/lib/docker/",
	)

	// 2. Esclusioni specifiche per modalità
	if mode != "clone" && mode != "crypted" {
		// In standard mode pialliamo la root dell'host
		excludes = append(excludes, "root/*")
	}

	// 3. Esclusioni Utente (leggiamo dal file se esiste)
	// 3. Esclusioni Utente (Intelligenza di percorso)
	userList := "/etc/coa/exclusion.list"
	
	// Se non esiste in /etc (non installato), usiamo il file di sviluppo locale
	if _, err := os.Stat(userList); os.IsNotExist(err) {
		userList = "conf/exclusion.list"
	}

	if data, err := os.ReadFile(userList); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			// Ignoriamo righe vuote e commenti
			if line != "" && !strings.HasPrefix(line, "#") {
				excludes = append(excludes, line)
			}
		}
	}

	// 4. Scriviamo il file finale per il motore C
	os.MkdirAll("/tmp/coa", 0755)
	os.WriteFile(outPath, []byte(strings.Join(excludes, "\n")+"\n"), 0644)

	return outPath
}

// GeneratePlan costruisce il piano di volo dinamico in base alla distribuzione rilevata [cite: 333, 334]
func GeneratePlan(d *Distro, mode string, workPath string) FlightPlan {
	plan := FlightPlan{
		PathLiveFs: workPath,
		Mode:       mode,
	}

	// 1. Astrazione Initramfs e Bootloaders (Il Terzo Pilastro) [cite: 80, 313, 324]
	// Delega il comando di generazione dell'initrd all'orchestratore [cite: 325, 328]
	switch d.FamilyID {
	case "debian":
		plan.InitrdCmd = "mkinitramfs -o {{out}} {{ver}}"
		plan.BootloadersPath = "" // Su Debian usiamo quelli di sistema [cite: 314]
	case "archlinux":
		// MODIFICA: Utilizziamo il flag -c per caricare la configurazione live
		// bridge-ata fisicamente da coa in /etc/mkinitcpio-live.conf.
		// Spostato da /tmp a /etc per evitare problemi di permessi/mount nel chroot.
		plan.InitrdCmd = "mkinitcpio -g {{out}} -k {{ver}}"
		plan.BootloadersPath = BootloaderRoot // Utilizza bootloader esterni [cite: 36, 72]
	case "fedora", "opensuse":
		plan.InitrdCmd = "dracut --nomadas --force {{out}} {{ver}}"
		plan.BootloadersPath = BootloaderRoot
	default:
		plan.InitrdCmd = "mkinitramfs -o {{out}} {{ver}}"
		plan.BootloadersPath = ""
	}

	// 2. Configurazione Utenti (Globale) [cite: 237, 319, 452]
	if mode == "standard" {
		// Gestione dinamica dei gruppi admin (sudo vs wheel) [cite: 34, 274, 275]
		adminGroup := "sudo"
		if d.FamilyID == "archlinux" || d.FamilyID == "fedora" {
			adminGroup = "wheel"
		}

		plan.Users = []UserConfig{
			{
				Login:    "live",
				Password: "$6$wM.wY0QtatvbQMHZ$QtIKXSpIsp2Sk57.Ny.JHk7hWDu.lxPtUYaTOiBnP4WBG5KS6JpUlpXj2kcSaaMje7fr01uiGmxZhE8kfZRqv.",
				Gecos:    "live,,,",
				Home:     "/home/live",
				Shell:    "/bin/bash",
				Groups:   []string{"cdrom", "audio", "video", "plugdev", "netdev", "autologin", adminGroup},
			},
		}
	} else {
		plan.Users = []UserConfig{}
	}

	// 3. Assemblaggio dinamico della catena di montaggio
	// NOTA: lay_prepare viene omesso qui perché eseguito preventivamente
	// in handleProduce per permettere il bridging dei file
	plan.Plan = []Action{
		{Command: "lay_users"}, // Identità nativa Yocto-style
	}

	// --- Task di "Vestizione" (Patching configurazioni) ---
	if d.FamilyID == "fedora" {
		plan.Plan = append(plan.Plan, Action{
			Command:    "sys_run",
			RunCommand: "cp",
			Args:       []string{"/tmp/coa/configs/dracut/fedora.conf", "/etc/dracut.conf.d/coa.conf"},
		})
	}
	
	// --- Generiamo la lista di esclusioni dinamica ---
	excludeFilePath := generateExcludeList(mode)

	// Proseguiamo con il resto del piano standard
	plan.Plan = append(plan.Plan,
		Action{Command: "lay_initrd"},     // Generazione ramdisk
		Action{Command: "lay_livestruct"}, // Kernel extraction
		Action{Command: "lay_isolinux"},   // BIOS bootloader
		Action{Command: "lay_uefi"},       // UEFI bootloader
		Action{
			Command:     "lay_squash",
			ExcludeList: excludeFilePath,  // <--- PASSIAMO IL FILE AL MOTORE C
		},     // Compressione Turbo SquashFS		
	)

	// Inserzione modulare per cifratura
	if mode == "crypted" {
		plan.Plan = append(plan.Plan, Action{
			Command:         "lay_crypted",
			CryptedPassword: "evolution",
		})
	}

	// --- definizione di isoname ---

	// 1. Recuperiamo l'hostname (es. colibri)
	hostname, _ := os.Hostname()

	// 2. Generiamo il timestamp nel formato richiesto (2026-04-07_0930)
	timestamp := time.Now().Format("2006-01-02_1504")

	// 3. Rileviamo l'architettura della CPU
	arch := runtime.GOARCH
	if arch == "amd64" {
		// arch = "x86_64"
	}

	// 4. Prepariamo i componenti del nome
	var nameParts []string
	nameParts = append(nameParts, d.DistroID)

	// Priorità: Codename > Release
	if d.CodenameID != "" {
		nameParts = append(nameParts, d.CodenameID)
	} else if d.ReleaseID != "" {
		nameParts = append(nameParts, d.ReleaseID)
	}

	// Aggiungiamo l'hostname
	if hostname != "" {
		nameParts = append(nameParts, hostname)
	}

	// 5. Uniamo i componenti base (es. arch-rolling-colibri)
	distroTag := strings.Join(nameParts, "-")

	// 6. Assembliamo il nome finale con timestamp e architettura
	// Formato: egg-of_distro-info-host_timestamp_arch.iso
	isoName := fmt.Sprintf("egg-of_%s_%s_%s.iso", distroTag, arch, timestamp)

	// --- Inserimento dell'azione nel piano per il motore oa ---
	plan.Plan = append(plan.Plan, Action{
		Command:   "lay_iso", 
		VolID:     "OA_LIVE",
		OutputISO: isoName,
	})

	plan.Plan = append(plan.Plan, Action{Command: "lay_cleanup"})

	return plan
}
