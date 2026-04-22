package engine

import (
	"coa/src/internal/pilot"
)

// GenerateIdentityActions crea il payload JSON per l'azione nativa oa_users.c
func GenerateIdentityActions(mode string, profile *pilot.BrainProfile) []Action {

	// =================================================================
	// 1. PREPARAZIONE DATI UTENTE
	// Assembliamo i gruppi secondari letti dal tuo YAML (es. wheel, audio, ecc.)
	// =================================================================
	var liveGroups []string
	if profile.Areas.Identity.AdminGroup != "" {
		liveGroups = append(liveGroups, profile.Areas.Identity.AdminGroup)
	}
	if len(profile.Areas.Identity.UserGroups) > 0 {
		liveGroups = append(liveGroups, profile.Areas.Identity.UserGroups...)
	}

	// Costruiamo l'oggetto Utente esattamente come lo vuole la tua struct C
	liveUser := UserDef{
		Login:    "live",
		Password: "$6$wM.wY0QtatvbQMHZ$QtIKXSpIsp2Sk57.Ny.JHk7hWDu.lxPtUYaTOiBnP4WBG5KS6JpUlpXj2kcSaaMje7fr01uiGmxZhE8kfZRqv.", // Hash per "live"
		Home:     "/home/live",
		Shell:    "/bin/bash",
		Gecos:    "live,,,",
		Uid:      1000,
		Gid:      1000,
		Groups:   liveGroups, // Il C (yocto_add_user_to_groups) li processerà nativamente
	}

	// =================================================================
	// 2. AZIONE PRINCIPALE: IL CHIRURGO (C)
	// Generiamo l'azione che innesca oa_users.c per fare Purge e Iniezione
	// =================================================================
	actions := []Action{
		{
			Command: "oa_users",
			Info:    "Gestione nativa identità (Purge & Inject stile Yocto)",
			Mode:    mode, // Passando il mode, il C sa in automatico se fare il Purge host
			Users:   []UserDef{liveUser},
			Chroot:  false, // Il C lavora sui file host iniettando /liveroot da solo
		},
	}

	// =================================================================
	// 3. IL FIX PER SUDO IN ARCH LINUX (BASH)
	// Creiamo un file in sudoers.d per dare i poteri al gruppo wheel.
	// Usiamo NOPASSWD per comodità nella Live (standard nelle ISO).
	// =================================================================
	if mode != "clone" && mode != "crypted" {
		actions = append(actions, Action{
			Command: "oa_shell",
			Info:    "Abilitazione poteri sudo per il gruppo wheel",
			// Crea il file e imposta i permessi stretti a 440 (vitale per non far crashare sudo)
			RunCommand: "echo '%wheel ALL=(ALL:ALL) NOPASSWD: ALL' > /etc/sudoers.d/10-live-wheel && chmod 440 /etc/sudoers.d/10-live-wheel",
			Chroot:     true, // Questo lo facciamo in chroot perché usa i comandi di sistema
		})
	}

	return actions
}
