package cmd

import (
	"github.com/spf13/cobra"
)

// sysinstallCmd è il comando padre: 'coa sysinstall'
// Funge da punto di ingresso unico per tutti i motori di installazione.
var sysinstallCmd = &cobra.Command{
	Use:   "sysinstall",
	Short: "Lancia l'installatore di sistema (GUI o TUI)",
	Long: `coa sysinstall è l'orchestratore per l'installazione del sistema su disco.
Permette di scegliere tra l'interfaccia grafica (Calamares) o quella testuale (Krill).

Esempi:
  sudo coa sysinstall calamares
  sudo coa sysinstall krill`,
	Run: func(cmd *cobra.Command, args []string) {
		// Se l'utente non specifica un sottocomando, mostriamo l'aiuto
		// Questo evita che il comando non faccia nulla se invocato da solo.
		cmd.Help()
	},
}

func init() {
	// Registriamo sysinstall nel comando principale di coa
	rootCmd.AddCommand(sysinstallCmd)
}
