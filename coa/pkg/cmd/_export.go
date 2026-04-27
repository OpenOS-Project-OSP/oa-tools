package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"coa/pkg/distro" // Assicurati che il path sia corretto per il tuo progetto

	"github.com/spf13/cobra"
)

// --- CONFIGURAZIONE ESPORTAZIONE ---
const (
	remoteUserHost = "root@192.168.1.2"
	remoteIsoPath  = "/var/lib/vz/template/iso/"
	remotePkgPath  = "/eggs/"
	isoSrcDir      = "/home/eggs"
)

var cleanExport bool

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export artifacts (iso, pkg) to a remote Proxmox storage",
}

var exportIsoCmd = &cobra.Command{
	Use:   "iso",
	Short: "Export the latest ISO to a remote Proxmox storage",
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements(cmd.Name(), false)
		handleExportIso(cleanExport)
	},
}

var exportPkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Export native packages (.deb, .rpm, .pkg.tar.zst) to Proxmox",
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements(cmd.Name(), false)
		handleExportPkg(cleanExport)
	},
}

func init() {
	exportCmd.PersistentFlags().BoolVar(&cleanExport, "clean", false, "Clean old versions on remote server before exporting")
	exportCmd.AddCommand(exportIsoCmd)
	exportCmd.AddCommand(exportPkgCmd)
	rootCmd.AddCommand(exportCmd)
}

// =====================================================================
// LOGICA DI ESPORTAZIONE
// =====================================================================

// handleExportPkg esporta solo i pacchetti della distro corrente
func handleExportPkg(clean bool) {
	myDistro := distro.NewDistro()
	family := myDistro.FamilyID

	LogCoala("Famiglia rilevata: %s. Ricerca pacchetti pertinenti...", family)

	var pattern string
	var extension string

	// Filtriamo per estensione in base alla famiglia
	switch family {
	case "debian", "ubuntu", "devuan":
		pattern = "oa-tools*.deb"
		extension = ".deb"
	case "arch":
		pattern = "oa-tools*.pkg.tar.zst"
		extension = ".pkg.tar.zst"
	case "fedora", "redhat", "suse":
		pattern = "oa-tools*.rpm"
		extension = ".rpm"
	default:
		// Se la famiglia non è riconosciuta, usiamo LogCoala per avvisare
		LogCoala("Nessuna regola di esportazione specifica per la famiglia: %s", family)
		return
	}

	foundFiles, _ := filepath.Glob(pattern)
	if len(foundFiles) == 0 {
		LogError("Nessun pacchetto %s trovato per l'esportazione.", extension)
		return
	}

	// SSH Multiplexing
	socketPath := "/tmp/coa-ssh-mux-pkg"
	muxArgs := []string{"-o", "ControlMaster=auto", "-o", "ControlPath=" + socketPath, "-o", "ControlPersist=2m"}
	defer func() {
		exec.Command("ssh", "-O", "exit", "-o", "ControlPath="+socketPath, remoteUserHost).Run()
		os.Remove(socketPath)
	}()

	if clean {
		LogCoala("Pulizia remota vecchi pacchetti %s...", extension)
		cleanCmdStr := fmt.Sprintf("rm -f %soa-tools*%s", remotePkgPath, extension)
		sshArgs := append(muxArgs, remoteUserHost, cleanCmdStr)

		if err := exec.Command("ssh", sshArgs...).Run(); err != nil {
			LogCoala("Pulizia remota non necessaria o fallita (nessun file trovato).")
		} else {
			LogSuccess("Vecchi pacchetti %s rimossi dal server.", extension)
		}
	}

	for _, pkg := range foundFiles {
		LogCoala("Esportazione: %s", pkg)
		dstStr := fmt.Sprintf("%s:%s", remoteUserHost, remotePkgPath)
		scpArgs := append(muxArgs, pkg, dstStr)

		scpCmd := exec.Command("scp", scpArgs...)
		scpCmd.Stdout, scpCmd.Stderr = os.Stdout, os.Stderr

		if err := scpCmd.Run(); err != nil {
			LogError("Trasferimento fallito per %s: %v", pkg, err)
		} else {
			LogSuccess("%s inviato con successo.", pkg)
		}
	}
}

func handleExportIso(clean bool) {
	// 1. Otteniamo il prefisso dinamico (egg-of-distro-host-arch-)
	d := distro.NewDistro()
	prefixBase := d.GetISOPrefix()
	isoPattern := prefixBase + "*.iso"

	// Ricerca nel nido locale
	allFiles, _ := filepath.Glob(filepath.Join(isoSrcDir, isoPattern))
	if len(allFiles) == 0 {
		LogError("Il nido è vuoto per il prefisso: %s", prefixBase)
		return
	}

	// 2. Identificazione dell'ultima ISO (basata su ModTime)
	var latestFile string
	var latestTime time.Time

	for _, path := range allFiles {
		if info, err := os.Stat(path); err == nil {
			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latestFile = path
			}
		}
	}

	targetFileName := filepath.Base(latestFile)
	LogCoala("Ultima ISO trovata: %s", targetFileName)

	// 3. Setup SSH Multiplexing
	socketPath := "/tmp/coa-ssh-mux"
	muxArgs := []string{"-o", "ControlPath=" + socketPath}

	// Start Master Connection
	exec.Command("ssh", "-M", "-f", "-N", "-o", "ControlPath="+socketPath, remoteUserHost).Run()
	defer exec.Command("ssh", "-O", "exit", "-o", "ControlPath="+socketPath, remoteUserHost).Run()

	// 4. Logica di Pulizia su Proxmox
	// Se clean è true, usiamo il prefisso per cancellare le versioni precedenti di QUESTA macchina
	if clean {
		LogCoala("Pulizia su Proxmox: rimozione versioni precedenti con prefisso %s", prefixBase)
		// Il comando remoto userà il prefisso per cancellare solo i file coerenti
		rmCmdStr := fmt.Sprintf("rm -f %s/%s*.iso", remoteIsoPath, prefixBase)
		sshCmd := exec.Command("ssh", append(muxArgs, remoteUserHost, rmCmdStr)...)
		if err := sshCmd.Run(); err != nil {
			LogCoala("Nessuna vecchia ISO rimossa su Proxmox.")
		}
	}

	// 5. Invio effettivo
	LogCoala("Inviando %s verso Proxmox...", targetFileName)
	dst := fmt.Sprintf("%s:%s", remoteUserHost, remoteIsoPath)
	scpCmd := exec.Command("scp", append(muxArgs, latestFile, dst)...)
	scpCmd.Stdout = os.Stdout
	scpCmd.Stderr = os.Stderr

	if err := scpCmd.Run(); err != nil {
		LogError("Errore durante il trasferimento: %v", err)
	} else {
		LogSuccess("Esportazione completata con successo!")
	}
}
