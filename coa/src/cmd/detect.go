package cmd

import (
	"coa/src/internal/distro"
	"coa/src/internal/engine"

	"github.com/spf13/cobra"
)

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Display detected host system information",
	Long: `The 'detect' command is a read-only diagnostic utility for the user. 

It performs a quick scan of the host environment to identify the running GNU/Linux distribution, its parent family (e.g., mapping Ubuntu to the Debian family), and the hardware architecture. 

It does not save this state or alter any configuration; it simply provides a clear overview of the environment 'coa' is currently running in.`,
	Example: `  # Display the host system profile
  coa detect`,
	Run: func(cmd *cobra.Command, args []string) {
		// Controllo sudo: è un comando informativo, non serve root
		CheckSudoRequirements(cmd.Name(), false)

		// 1. Rileva la distribuzione host
		myDistro := distro.NewDistro()

		// 2. Passa l'oggetto all'engine per la stampa a video
		engine.HandleDetect(myDistro)
	},
}

func init() {
	rootCmd.AddCommand(detectCmd)
}
