package engine

import (
	"os"
	"strings"
)

// GenerateExcludeList crea il file .list dinamico per mksquashfs.
// La 'G' maiuscola permette a remaster.go di chiamarla liberamente.
func GenerateExcludeList(mode string) string {
	outPath := "/tmp/coa/excludes.list"
	var excludes []string

	// ==========================================================
	// 1. Filesystem Virtuali e Temporanei
	// Usiamo l'asterisco (/*) per SVUOTARE la cartella ma
	// MANTENERE la cartella vuota (necessaria per i mount point al boot)
	// ==========================================================
	excludes = append(excludes,
		"dev/*",
		"proc/*",
		"sys/*",
		"run/*",
		"tmp/*",
		"var/tmp/*",
		"mnt/*",
		"media/*",
		"lost+found",
		// Aggiunte salvavita da ieri: non comprimere l'overlay di lavoro
		"home/eggs/.overlay/*",
		"home/eggs/isodir/*",
	)

	// ==========================================================
	// 2. Esclusioni Standard di Sistema (Da penguins-eggs)
	// ==========================================================
	excludes = append(excludes,
		"boot/efi/EFI",
		"boot/loader/entries/",
		"etc/fstab",
		"etc/mtab",
		"var/lib/docker/",
		"var/lib/containers/",
		"etc/udev/rules.d/70-persistent-cd.rules",
		"etc/udev/rules.d/70-persistent-net.rules",
	)

	// ==========================================================
	// 3. Hack per Debian: cryptdisks
	// Grazie a mksquashfs -wildcards, possiamo evitare di fare
	// scansioni del disco in Go. Basta questa stringa magica!
	// ==========================================================
	excludes = append(excludes, "etc/rc*.d/*cryptdisks*")

	// ==========================================================
	// 4. Sicurezza Root / Home (In base al mode)
	// ==========================================================
	if mode != "clone" && mode != "crypted" {
		// Come in TS: root/* e root/.*
		excludes = append(excludes, "root/*", "root/.*")
	}

	// ==========================================================
	// 5. Liste Utente (Custom)
	// ==========================================================
	userList := "/etc/coa/exclusion.list"
	if _, err := os.Stat(userList); os.IsNotExist(err) {
		userList = "conf/exclusion.list"
	}

	if data, err := os.ReadFile(userList); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				// FIX: Come facevi in TS, rimuoviamo lo slash iniziale.
				// mksquashfs lavora con percorsi relativi alla directory sorgente.
				line = strings.TrimPrefix(line, "/")
				excludes = append(excludes, line)
			}
		}
	}

	// Creiamo la directory temporanea e scriviamo il file list
	os.MkdirAll("/tmp/coa", 0755)

	// Uniamo tutto con a capo
	fileContent := strings.Join(excludes, "\n") + "\n"
	os.WriteFile(outPath, []byte(fileContent), 0644)

	return outPath
}
