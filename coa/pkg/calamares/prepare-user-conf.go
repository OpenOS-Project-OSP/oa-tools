package calamares

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func PrepareUserConf() error {
	// 1. Wishlist universale dei gruppi
	wishlist := []string{"wheel", "sudo", "audio", "video", "storage", "plugdev", "netdev", "network", "lpadmin", "scanner", "users"}

	// 2. Leggiamo i gruppi reali del sistema live
	data, err := os.ReadFile("/etc/group")
	if err != nil {
		return err
	}
	content := string(data)

	var validGroups []string
	for _, g := range wishlist {
		if strings.Contains(content, g+":") {
			validGroups = append(validGroups, g)
		}
	}

	// 3. Generiamo lo YAML "Libertario"
	var yamlGroups string
	for _, v := range validGroups {
		yamlGroups += fmt.Sprintf("    - %s\n", v)
	}

	config := fmt.Sprintf(`---
# OA-Tools: Configurazione Universale Dinamica
defaultGroups:
%s
checkPasswordQuality: false
passwordCheckMethod: none
minPasswordLength: 1
allowReusePassword: true
`, yamlGroups)

	// 4. Scrittura del file nella sessione live
	targetPath := "/etc/calamares/modules/users.conf"
	os.MkdirAll(filepath.Dir(targetPath), 0755)

	return os.WriteFile(targetPath, []byte(config), 0644)
}
