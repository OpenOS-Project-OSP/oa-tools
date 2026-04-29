package cmd

import (
	"coa/pkg/utils"
	"os"

	"github.com/spf13/cobra"
)

// krillSubCmd definisce il sottocomando 'coa sysinstall krill'
var krillSubCmd = &cobra.Command{
	Use:   "krill",
	Short: "Lancia l'installatore testuale Krill (TUI)",
	Run: func(cmd *cobra.Command, args []string) {
		// Anche se è un mock, manteniamo la coerenza dei permessi
		CheckSudoRequirements("sysinstall krill", true)

		runKrillInstaller()
	},
}

// runKrillInstaller è il segnaposto per l'installatore testuale.
// Per ora non rompe i coglioni e ci permette di compilare tutto.
func runKrillInstaller() {
	utils.LogCoala("%s[Krill]%s L'installatore TUI non è ancora pronto. Usa Calamares per ora!", utils.ColorYellow, utils.ColorReset)
	os.Exit(0)
}

func init() {
	// Appendiamo il comando a sysinstallCmd
	sysinstallCmd.AddCommand(krillSubCmd)
}
