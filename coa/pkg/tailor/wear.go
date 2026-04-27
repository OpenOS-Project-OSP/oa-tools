package tailor

import (
	"coa/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Wear(costumeName string, noAcc bool, noFirm bool) error {
	root, err := getWardrobeRoot()
	if err != nil {
		return err
	}

	// 1. Identifichiamo se è un costume o un accessorio
	costumeDir := filepath.Join(root, "costumes", costumeName)
	if _, err := os.Stat(costumeDir); os.IsNotExist(err) {
		costumeDir = filepath.Join(root, "accessories", costumeName)
	}

	// Verifica se la cartella esiste davvero prima di procedere
	if _, err := os.Stat(costumeDir); os.IsNotExist(err) {
		return fmt.Errorf("costume o accessorio '%s' non trovato nel wardrobe", costumeName)
	}

	// 🚀 CAMBIO DIRECTORY: Ci spostiamo nel costume per risolvere i path relativi degli script
	originalWd, _ := os.Getwd()
	if err := os.Chdir(costumeDir); err != nil {
		return fmt.Errorf("impossibile entrare nella cartella del costume: %v", err)
	}
	defer os.Chdir(originalWd) // Torna sempre alla cartella di partenza

	// 2. Trova e carica lo YAML compatibile
	yamlFile := findCompatibleYaml(costumeDir)
	suit, err := loadSuit(yamlFile)
	if err != nil {
		return fmt.Errorf("impossibile caricare il costume %s: %v", costumeName, err)
	}

	utils.LogCoala("Indossando il costume: %s", suit.Name)

	// 3. Gestione Repositories (Update/Upgrade)
	if suit.Sequence.Repositories.Update {
		utils.LogCoala("Aggiornamento repository...")
		utils.ExecQuiet("apt-get update")
	}

	// 4. Installazione Pacchetti
	if len(suit.Sequence.Packages) > 0 {
		utils.LogCoala("Verifica pacchetti nel database APT...")
		available := getAvailablePackages()
		var toInstall []string

		for _, pkg := range suit.Sequence.Packages {
			// Puliuamo il nome del pacchetto da spazi, tab o caratteri invisibili
			cleanPkg := strings.TrimSpace(pkg)

			// Se la riga era vuota o solo uno spazio (succede con i commenti nello YAML)
			if cleanPkg == "" {
				continue
			}

			if _, ok := available[cleanPkg]; ok {
				toInstall = append(toInstall, cleanPkg)
			} else {
				// Logghiamo tra parentesi quadre per vedere se ci sono spazi residui
				utils.LogCoala("⚠️  Pacchetto non trovato: [%s]", cleanPkg)
			}
		}

		if len(toInstall) > 0 {
			utils.LogCoala("Installazione di %d pacchetti in corso...", len(toInstall))
			installWithRetries(toInstall, 3)
		} else {
			utils.LogCoala("Nessun pacchetto nuovo da installare.")
		}
	}

	// 5. 🚀 LA PARACULATA: Rsync della sysroot
	sysrootPath := filepath.Join(costumeDir, "sysroot")
	if _, err := os.Stat(sysrootPath); err == nil {
		utils.LogCoala("Applicazione personalizzazioni (sysroot)...")
		// Usiamo il path assoluto della sysroot per rsync
		cmd := fmt.Sprintf("rsync -aHSX %s/ /", sysrootPath)
		if err := utils.Exec(cmd); err != nil {
			utils.LogError("Errore durante l'applicazione della sysroot: %v", err)
		}
	}

	// 6. Finalizzazione (comandi personalizzati)
	// Essendo dentro costumeDir, i path tipo ../../scripts/ ora funzionano!
	if len(suit.Sequence.Cmds) > 0 {
		utils.LogCoala("Esecuzione comandi di sequenza...")
		for _, command := range suit.Sequence.Cmds {
			utils.Exec(command)
		}
	}

	if suit.Finalize.Customize && len(suit.Finalize.Cmds) > 0 {
		utils.LogCoala("Finalizzazione costume...")
		for _, command := range suit.Finalize.Cmds {
			utils.Exec(command)
		}
	}

	utils.LogCoala("✅ Costume '%s' indossato con successo!", suit.Name)

	if suit.Reboot {
		utils.LogCoala("🔄 Questo costume richiede un riavvio.")
	}

	return nil
}
