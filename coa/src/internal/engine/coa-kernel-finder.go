package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindKernelAndInitrd cerca i file di boot corretti in base alla famiglia della distro
func FindKernelAndInitrd(workPath string, familyID string) (kernelSrc string, initrdSrc string, err error) {
	liveroot := filepath.Join(workPath, "liveroot")

	switch familyID {
	case "debian", "ubuntu":
		// Debian usa solitamente i symlink nella root
		kernelSrc = filepath.Join(liveroot, "vmlinuz")
		initrdSrc = filepath.Join(liveroot, "initrd.img")

		// Se per qualche motivo mancano i symlink, cerchiamo in /boot
		if _, err := os.Stat(kernelSrc); os.IsNotExist(err) {
			kernelSrc, initrdSrc = searchInBoot(liveroot, "vmlinuz-*", "initrd.img-*")
		}

	case "archlinux", "arch":
		// Arch supporta diversi kernel (-lts, -zen, -hardened, default).
		// Usiamo il globbing per trovare automaticamente quello installato nella liveroot.
		kernelSrc, initrdSrc = searchInBoot(liveroot, "vmlinuz-linux*", "initramfs-linux*.img")

	case "fedora", "redhat":
		// Fedora usa le versioni. Dobbiamo cercare in /boot ed escludere le immagini "rescue"
		kernelSrc, initrdSrc = searchInBoot(liveroot, "vmlinuz-*", "initramfs-*.img")

	default:
		// Fallback generico
		kernelSrc, initrdSrc = searchInBoot(liveroot, "vmlinuz-*", "initrd.img-*")
	}

	// Controllo di sicurezza finale
	if _, err := os.Stat(kernelSrc); os.IsNotExist(err) {
		return "", "", fmt.Errorf("kernel non trovato in %s per la famiglia %s", kernelSrc, familyID)
	}
	if _, err := os.Stat(initrdSrc); os.IsNotExist(err) {
		return "", "", fmt.Errorf("initramfs non trovato in %s per la famiglia %s", initrdSrc, familyID)
	}

	return kernelSrc, initrdSrc, nil
}

// Helper interno per cercare file in /boot tramite wildcard
func searchInBoot(liveroot, kernelPattern, initrdPattern string) (string, string) {
	bootDir := filepath.Join(liveroot, "boot")

	kernels, _ := filepath.Glob(filepath.Join(bootDir, kernelPattern))
	var validKernel string

	// Prendiamo il primo kernel valido che non sia un file di rescue
	for _, k := range kernels {
		if !strings.Contains(k, "rescue") {
			validKernel = k
			break
		}
	}

	if validKernel == "" {
		return "", ""
	}

	// Estraiamo la versione del kernel per trovare l'initrd corrispondente
	// Es: vmlinuz-6.8.0-35 -> 6.8.0-35
	baseName := filepath.Base(validKernel)
	version := strings.TrimPrefix(baseName, strings.Split(kernelPattern, "*")[0])

	// Ricostruiamo il nome dell'initrd
	expectedInitrdName := strings.Replace(initrdPattern, "*", version, 1)
	validInitrd := filepath.Join(bootDir, expectedInitrdName)

	return validKernel, validInitrd
}
