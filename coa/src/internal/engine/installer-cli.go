package engine

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SelectTargetDisk mostra i dischi disponibili e fa scegliere l'utente
func SelectTargetDisk() (*BlockDevice, error) {
	fmt.Println("\033[1;36m[oa-installer]\033[0m Ricerca dei dischi di sistema in corso...")

	disks, err := GetAvailableDisks()
	if err != nil {
		return nil, fmt.Errorf("impossibile leggere i dischi: %v", err)
	}

	if len(disks) == 0 {
		return nil, fmt.Errorf("nessun disco fisico trovato sul sistema")
	}

	fmt.Println("\nDischi disponibili per l'installazione:")
	fmt.Println(strings.Repeat("-", 60))
	for i, disk := range disks {
		fmt.Printf("[%d] /dev/%s \t- %s (%s)\n", i+1, disk.Name, disk.Model, FormatSize(disk.Size))
	}
	fmt.Println(strings.Repeat("-", 60))

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("\nSeleziona il disco su cui installare \033[1;31m(ATTENZIONE: I DATI VERRANNO CANCELLATI)\033[0m [1-%d]: ", len(disks))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Validazione dell'input
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(disks) {
			fmt.Printf("\033[1;33mScelta non valida. Inserisci un numero tra 1 e %d.\033[0m\n", len(disks))
			continue
		}

		selectedDisk := disks[choice-1]

		// Conferma di sicurezza finale
		fmt.Printf("\nHai selezionato: /dev/%s (%s).\n", selectedDisk.Name, FormatSize(selectedDisk.Size))
		fmt.Print("Sei ASSOLUTAMENTE sicuro di voler procedere? (scrivi 'si' per confermare): ")

		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))

		if confirm == "si" || confirm == "sì" {
			return &selectedDisk, nil // <-- Esce dal loop e restituisce il disco
		} else {
			fmt.Println("\nOperazione annullata. Scegli un altro disco o premi Ctrl+C per uscire.")
			// Il loop ricomincia e chiede di nuovo
		}
	}
}
