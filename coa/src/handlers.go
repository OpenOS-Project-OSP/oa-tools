package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

// getOaPath cerca il braccio operativo (oa) nel sistema o nel percorso relativo
func getOaPath() string {
	path, err := exec.LookPath("oa")
	if err == nil {
		return path
	}
	if _, err := os.Stat("../oa/oa"); err == nil {
		return "../oa/oa"
	}
	if _, err := os.Stat("./oa/oa"); err == nil {
		return "./oa/oa"
	}
	return "oa"
}

// bridgeConfigs sovrascrive la configurazione mkinitcpio nella liveroot per Arch
func bridgeConfigs(d *Distro, workPath string) error {
	if d.FamilyID != "archlinux" {
		return nil
	}

	presetName := "live-arch.conf"
	if d.DistroID == "manjaro" || d.DistroID == "biglinux" {
		presetName = fmt.Sprintf("live-%s.conf", d.DistroID)
	}

	src := fmt.Sprintf("/tmp/coa/configs/mkinitcpio/%s", presetName)
	dst := filepath.Join(workPath, "liveroot", "etc", "mkinitcpio.conf")

	fmt.Printf("\033[1;34m[coa]\033[0m Overwriting liveroot /etc/mkinitcpio.conf with %s...\n", presetName)

	if err := copyFile(src, dst); err != nil {
		return err
	}
	return os.Chmod(dst, 0644)
}

// copyFile è una utility per la copia fisica tra percorsi host
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// executePlan trasforma il piano in JSON e lo dà in pasto a oa
func executePlan(plan FlightPlan) {
	jsonData, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		log.Fatalf("\033[1;31m[coa]\033[0m JSON Error: %v", err)
	}

	tmpJsonPath := "/tmp/plan_coa_tmp.json"
	err = os.WriteFile(tmpJsonPath, jsonData, 0644)
	if err != nil {
		log.Fatalf("\033[1;31m[coa]\033[0m Temp file error: %v", err)
	}
	defer os.Remove(tmpJsonPath)

	oaPath := getOaPath()
	cmd := exec.Command("sudo", oaPath, tmpJsonPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("\n\033[1;31m[coa]\033[0m Engine error: %v", err)
	}
}

// handleProduce gestisce la creazione della ISO
func handleProduce(mode string, workPath string, d *Distro) {
	tempConfigPath := "/tmp/coa/configs"
	fmt.Printf("\033[1;32m[coa]\033[0m Extracting internal configurations to %s...\n", tempConfigPath)
	if err := ExtractConfigs(tempConfigPath); err != nil {
		log.Fatalf("\033[1;31m[coa]\033[0m Asset extraction failed: %v", err)
	}

	fmt.Printf("\033[1;32m[coa]\033[0m Ensuring bootloaders are present in %s...\n", BootloaderRoot)
	if _, err := EnsureBootloaders(); err != nil {
		log.Fatalf("\033[1;31m[coa]\033[0m Bootloader retrieval failed: %v", err)
	}

	fmt.Printf("\033[1;32m[coa]\033[0m Preparing environment...\n")
	prePlan := FlightPlan{
		PathLiveFs: workPath,
		Mode:       mode,
		Plan:       []Action{{Command: "lay_prepare"}},
	}
	executePlan(prePlan)

	if err := bridgeConfigs(d, workPath); err != nil {
		log.Fatalf("\033[1;31m[coa]\033[0m Config bridging failed: %v", err)
	}

	fmt.Printf("\033[1;32m[coa]\033[0m Starting production flight...\n")
	flightPlan := GeneratePlan(d, mode, workPath)

	if len(flightPlan.Plan) > 0 && flightPlan.Plan[0].Command == "action_prepare" {
		flightPlan.Plan = flightPlan.Plan[1:]
	}
	executePlan(flightPlan)
}

// handleKill gestisce la pulizia
func handleKill() {
	fmt.Println("\033[1;33m[coa]\033[0m Freeing the nest...")
	oaPath := getOaPath()
	cmd := exec.Command("sudo", oaPath, "cleanup")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("\033[1;31m[coa]\033[0m Cleanup (unmount) failed: %v\n", err)
	}

	workPath := "/home/eggs"
	fmt.Printf("\033[1;31m[coa]\033[0m Removing workspace: %s\n", workPath)
	rmCmd := exec.Command("sudo", "rm", "-rf", workPath)
	if err := rmCmd.Run(); err != nil {
		fmt.Printf("\033[1;31m[coa]\033[0m Physical removal failed: %v\n", err)
	} else {
		fmt.Println("\033[1;32m[coa]\033[0m Nest is empty. System clean.")
	}
}

// handleDetect mostra le info rilevate dal discovery
func handleDetect(d *Distro) {
	fmt.Println("\033[1;34m--- coa distro detect ---\033[0m")
	fmt.Printf("Host Distro:     %s\n", d.DistroID)
	fmt.Printf("Family:          %s\n", d.FamilyID)
	fmt.Printf("DistroLike:      %s\n", d.DistroLike)
	fmt.Printf("Codename:        %s\n", d.CodenameID)
	fmt.Printf("Release:         %s\n", d.ReleaseID)
	fmt.Printf("DistroUniqueID:  %s\n", d.DistroUniqueID)
}

// handleAdapt adatta la risoluzione del monitor per le VM
func handleAdapt() {
	fmt.Println("\033[1;33m[coa]\033[0m Adapting monitor resolution...")
	virtualOutputs := []string{"Virtual-0", "Virtual-1", "Virtual-2", "Virtual-3"}
	for _, output := range virtualOutputs {
		cmd := exec.Command("xrandr", "--output", output, "--auto")
		_ = cmd.Run()
	}
	fmt.Println("\033[1;32m[coa]\033[0m Resolution adapted.")
}

// handleExport copia in remoto la ISO utilizzando SSH/SCP (no sudo locale) con Multiplexing
func handleExport(clean bool) {
	remoteUserHost := "root@192.168.1.2"
	remotePath := "/var/lib/vz/template/iso/"
	srcDir := "/home/eggs"

	// --- 1. SELEZIONE LOCALE: Identifichiamo l'uovo più fresco ---
	allFiles, _ := filepath.Glob(filepath.Join(srcDir, "egg-of_*.iso"))
	if len(allFiles) == 0 {
		fmt.Println("\033[1;31m[coa]\033[0m Nest is empty. No ISOs found in", srcDir)
		return
	}

	latestFiles := make(map[string]string)
	re := regexp.MustCompile(`_\d{4}-\d{2}-\d{2}_\d{4}\.iso$`)

	for _, path := range allFiles {
		fileName := filepath.Base(path)
		prefix := re.ReplaceAllString(fileName, "")

		if info, err := os.Stat(path); err == nil {
			if current, exists := latestFiles[prefix]; exists {
				cInfo, _ := os.Stat(current)
				if info.ModTime().After(cInfo.ModTime()) {
					latestFiles[prefix] = path
				}
			} else {
				latestFiles[prefix] = path
			}
		}
	}

	// --- 2. CONFIGURAZIONE SSH MULTIPLEXING ---
	// Questi parametri dicono a SSH/SCP di riutilizzare la stessa connessione
	socketPath := "/tmp/coa-ssh-mux"
	muxArgs := []string{
		"-o", "ControlMaster=auto",
		"-o", "ControlPath=" + socketPath,
		"-o", "ControlPersist=2m", // Mantiene il tunnel aperto per 2 minuti
	}

	// Defer per smontare elegantemente il tunnel master quando la funzione termina
	defer func() {
		exec.Command("ssh", "-O", "exit", "-o", "ControlPath="+socketPath, remoteUserHost).Run()
		os.Remove(socketPath)
	}()

	// --- 3. PULIZIA REMOTA E TRASFERIMENTO ---
	for prefix, localPath := range latestFiles {
		targetFileName := filepath.Base(localPath)
		fmt.Printf("\n\033[1;35m[PROCESS]\033[0m Family: %s\n", prefix)

		// CANCELLAZIONE
		if clean {
			fmt.Printf("\033[1;34m[CLEAN]\033[0m Removing old versions on Proxmox...\n")

			rmCmdStr := fmt.Sprintf("rm -f %s%s*", remotePath, prefix)
			// Uniamo gli argomenti di multiplexing con quelli standard
			sshArgs := append(muxArgs, remoteUserHost, rmCmdStr)
			sshCmd := exec.Command("ssh", sshArgs...)

			sshCmd.Stdout = os.Stdout
			sshCmd.Stderr = os.Stderr
			sshCmd.Stdin = os.Stdin // FONDAMENTALE: permette a SSH di leggere la password dal terminale

			if err := sshCmd.Run(); err != nil {
				fmt.Printf("\033[1;33m[WARNING]\033[0m Remote cleanup failed or no old files found.\n")
			} else {
				fmt.Printf("\033[1;32m[CLEAN]\033[0m Old versions removed.\n")
			}
		}

		// COPIA
		fmt.Printf("\033[1;32m[COPY]\033[0m Sending %s to Proxmox...\n", targetFileName)

		dstStr := fmt.Sprintf("%s:%s", remoteUserHost, remotePath)
		scpArgs := append(muxArgs, localPath, dstStr)
		scpCmd := exec.Command("scp", scpArgs...)

		scpCmd.Stdout = os.Stdout
		scpCmd.Stderr = os.Stderr
		scpCmd.Stdin = os.Stdin // Stesso discorso, serve il flusso interattivo

		if err := scpCmd.Run(); err != nil {
			fmt.Printf("\033[1;31m[ERROR]\033[0m Copy failed: %v\n", err)
		} else {
			fmt.Printf("\033[1;32m[SUCCESS]\033[0m %s is now on Proxmox.\n", targetFileName)
		}
	}
}
