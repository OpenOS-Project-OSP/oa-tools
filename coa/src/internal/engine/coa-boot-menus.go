package engine

import (
	"coa/src/internal/distro"
	"coa/src/internal/pilot"
	"fmt"
	"os"
	"path/filepath"
)

// GenerateBootMenus scrive i menu di avvio ricavando i parametri dal Cervello (YAML)
func GenerateBootMenus(workPath string, d *distro.Distro, profile *pilot.BrainProfile, volId string) error {
	grubCfgPath := filepath.Join(workPath, "isodir", "boot", "grub", "grub.cfg")
	isolinuxCfgPath := filepath.Join(workPath, "isodir", "isolinux", "isolinux.cfg")

	os.MkdirAll(filepath.Dir(grubCfgPath), 0755)
	os.MkdirAll(filepath.Dir(isolinuxCfgPath), 0755)

	// =================================================================
	// 1. LOGICA DI ESTRAZIONE PARAMETRI (Da pilot/boot.go)
	// =================================================================

	// Default: Debian Style
	bootParams := "boot=live components quiet splash"

	// Override 1: ArchLinux
	if d.FamilyID == "archlinux" {
		bootParams = "archisobasedir=arch archisolabel=" + volId + " rw"
	}

	// Override 2: YAML Profile (Task "boot")
	for _, t := range profile.Tasks {
		if t.Name == "boot" && len(t.Commands) > 0 {
			bootParams = t.Commands[0]
			break // Trovato, usciamo dal ciclo
		}
	}

	// =================================================================
	// 2. GENERAZIONE MENU GRUB (UEFI/Legacy)
	// =================================================================
	grubContent := fmt.Sprintf(`set timeout=5
set default=0

# Ricerca della partizione tramite Label per evitare il rescue prompt
search --no-floppy --set=root --label %s

menuentry "Start OA Live (%s)" {
    linux /live/vmlinuz %s
    initrd /live/initrd.img
}

menuentry "Start OA Live (%s) - RAM mode" {
    linux /live/vmlinuz %s toram
    initrd /live/initrd.img
}
`, volId, d.FamilyID, bootParams, d.FamilyID, bootParams)

	// =================================================================
	// 3. GENERAZIONE MENU ISOLINUX (BIOS)
	// =================================================================
	isolinuxContent := fmt.Sprintf(`UI vesamenu.c32
TIMEOUT 50
DEFAULT live
PROMPT 0

LABEL live
    MENU LABEL Start OA Live (%s)
    LINUX /live/vmlinuz
    APPEND %s
    INITRD /live/initrd.img

LABEL ram
    MENU LABEL Start OA Live (%s) - RAM mode
    LINUX /live/vmlinuz
    APPEND %s toram
    INITRD /live/initrd.img
`, d.FamilyID, bootParams, d.FamilyID, bootParams)

	// =================================================================
	// 4. SCRITTURA SU DISCO
	// =================================================================
	err := os.WriteFile(grubCfgPath, []byte(grubContent), 0644)
	if err != nil {
		return fmt.Errorf("errore scrittura grub.cfg: %w", err)
	}

	err = os.WriteFile(isolinuxCfgPath, []byte(isolinuxContent), 0644)
	if err != nil {
		return fmt.Errorf("errore scrittura isolinux.cfg: %w", err)
	}

	// Trampolino EFI in /EFI/BOOT/grub.cfg (Fisso e strutturale)
	efiTrampolinePath := filepath.Join(workPath, "isodir", "EFI", "BOOT", "grub.cfg")
	os.MkdirAll(filepath.Dir(efiTrampolinePath), 0755)
	os.WriteFile(efiTrampolinePath, []byte("search --set=root --file /boot/grub/grub.cfg\nconfigfile /boot/grub/grub.cfg\n"), 0644)

	return nil
}
