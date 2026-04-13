// Copyright 2026 Piero Proietti <piero.proietti@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"coa/src/internal/distro"
)

// BootloaderRoot definisce dove vengono estratti i bootloader.
const BootloaderRoot = "/tmp/coa/bootloaders"

// Action rappresenta un singolo blocco "command" nell'array "plan"
type Action struct {
	Command         string   `json:"command"`
	VolID           string   `json:"volid,omitempty"`
	OutputISO       string   `json:"output_iso,omitempty"`
	CryptedPassword string   `json:"crypted_password,omitempty"`
	RunCommand      string   `json:"run_command,omitempty"`
	ExcludeList     string   `json:"exclude_list,omitempty"`
	BootParams      string   `json:"boot_params,omitempty"` // Parametri dinamici per il bootloader
	Args            []string `json:"args,omitempty"`
}

// UserConfig definisce la struttura per la creazione nativa dell'utente live
type UserConfig struct {
	Login    string   `json:"login"`
	Password string   `json:"password"`
	Gecos    string   `json:"gecos"`
	Home     string   `json:"home"`
	Shell    string   `json:"shell"`
	Groups   []string `json:"groups"`
}

// FlightPlan è l'oggetto JSON principale inviato al motore oa
type FlightPlan struct {
	PathLiveFs      string       `json:"pathLiveFs"`
	Mode            string       `json:"mode"`
	Family          string       `json:"family"`
	InitrdCmd       string       `json:"initrd_cmd"`
	BootloadersPath string       `json:"bootloaders_path"`
	Users           []UserConfig `json:"users"`
	Plan            []Action     `json:"plan"`
}

// generateExcludeList crea il file .list dinamico per mksquashfs
func generateExcludeList(mode string) string {
	outPath := "/tmp/coa/excludes.list"
	var excludes []string

	excludes = append(excludes,
		"boot/efi/EFI",
		"boot/loader/entries/",
		"etc/fstab",
		"var/lib/docker/",
	)

	if mode != "clone" && mode != "crypted" {
		excludes = append(excludes, "root/*")
	}

	userList := "/etc/coa/exclusion.list"
	if _, err := os.Stat(userList); os.IsNotExist(err) {
		userList = "conf/exclusion.list"
	}

	if data, err := os.ReadFile(userList); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				excludes = append(excludes, line)
			}
		}
	}

	os.MkdirAll("/tmp/coa", 0755)
	os.WriteFile(outPath, []byte(strings.Join(excludes, "\n")+"\n"), 0644)

	return outPath
}

// GeneratePlan costruisce il piano di volo dinamico
func GeneratePlan(d *distro.Distro, mode string, workPath string) FlightPlan {
	plan := FlightPlan{
		PathLiveFs:      workPath,
		Mode:            mode,
		Family:          d.FamilyID,
		BootloadersPath: BootloaderRoot,
	}

	bootParams := "boot=live components"
	switch d.FamilyID {
	case "archlinux":
		bootParams = "archisobasedir=arch archisolabel=OA_LIVE"
	case "fedora", "rhel", "centos", "rocky", "almalinux", "opensuse":
		bootParams = "root=live:CDLABEL=OA_LIVE rd.live.image rd.live.dir=live rd.live.squashimg=filesystem.squashfs selinux=0"
	}

	switch d.FamilyID {
	case "debian":
		plan.InitrdCmd = "mkinitramfs -o {{out}} {{ver}}"
	case "archlinux":
		plan.InitrdCmd = "mkinitcpio -c /etc/coa_mkinitcpio.conf -g {{out}} -k {{ver}}"
	case "fedora", "rhel", "centos", "rocky", "almalinux":
		plan.InitrdCmd = "dracut --no-hostonly --nomdadmconf --nolvmconf --xz --add dmsquash-live --add rootfs-block --add bash --add-drivers \"overlay squashfs loop iso9660 cdrom sr_mod\" --force {{out}} {{ver}}"
	default:
		plan.InitrdCmd = "mkinitramfs -o {{out}} {{ver}}"
	}

	if mode == "standard" {
		adminGroup := "sudo"
		if d.FamilyID == "archlinux" || d.FamilyID == "fedora" || d.FamilyID == "rhel" || d.FamilyID == "centos" || d.FamilyID == "rocky" || d.FamilyID == "almalinux" {
			adminGroup = "wheel"
		}

		plan.Users = []UserConfig{
			{
				Login:    "live",
				Password: "$6$wM.wY0QtatvbQMHZ$QtIKXSpIsp2Sk57.Ny.JHk7hWDu.lxPtUYaTOiBnP4WBG5KS6JpUlpXj2kcSaaMje7fr01uiGmxZhE8kfZRqv.",
				Gecos:    "live,,,",
				Home:     "/home/live",
				Shell:    "/bin/bash",
				Groups:   []string{"cdrom", "audio", "video", "plugdev", "netdev", "autologin", adminGroup},
			},
		}

		plan.Plan = []Action{
			{Command: "oa_remaster_users"},
		}

		sudoersDir := "/etc/sudoers.d"
		sudoersFile := "/etc/sudoers.d/00-oa-live"
		sudoersContent := fmt.Sprintf("%%%s ALL=(ALL) NOPASSWD: ALL", adminGroup)
		sudoersCmd := fmt.Sprintf("mkdir -p %s && echo '%s' > %s && chmod 0440 %s", sudoersDir, sudoersContent, sudoersFile, sudoersFile)

		plan.Plan = append(plan.Plan, Action{
			Command:    "oa_sys_run",
			RunCommand: "sh",
			Args:       []string{"-c", sudoersCmd},
		})

	} else {
		plan.Users = []UserConfig{}
		plan.Plan = []Action{
			{Command: "oa_remaster_users"},
		}
	}

	if d.FamilyID == "fedora" || d.FamilyID == "rhel" || d.FamilyID == "centos" || d.FamilyID == "rocky" || d.FamilyID == "almalinux" {
		targetConfDir := fmt.Sprintf("%s/liveroot/etc/dracut.conf.d", workPath)
		targetConfPath := fmt.Sprintf("%s/coa.conf", targetConfDir)
		dracutConfig := `hostonly="no"\nadd_dracutmodules+=" dmsquash-live rootfs-block bash "\ncompress="xz"`
		writeCmd := fmt.Sprintf(`echo -e '%s' > %s`, dracutConfig, targetConfPath)

		plan.Plan = append(plan.Plan, Action{
			Command:    "oa_sys_run",
			RunCommand: "mkdir",
			Args:       []string{"-p", targetConfDir},
		})

		plan.Plan = append(plan.Plan, Action{
			Command:    "oa_sys_run",
			RunCommand: "sh",
			Args:       []string{"-c", writeCmd},
		})
	}

	excludeFilePath := generateExcludeList(mode)

	plan.Plan = append(plan.Plan,
		Action{Command: "oa_remaster_initrd"},
		Action{Command: "oa_remaster_livestruct"},
		Action{Command: "oa_remaster_isolinux", BootParams: bootParams},
		Action{Command: "oa_remaster_uefi", BootParams: bootParams},
		Action{
			Command:     "oa_remaster_squash",
			ExcludeList: excludeFilePath,
		},
	)

	if mode == "crypted" {
		plan.Plan = append(plan.Plan, Action{
			Command:         "oa_remaster_crypted",
			CryptedPassword: "evolution",
		})
	}

	hostname, _ := os.Hostname()
	timestamp := time.Now().Format("2006-01-02_1504")
	arch := runtime.GOARCH

	var nameParts []string
	nameParts = append(nameParts, d.DistroID)
	if d.CodenameID != "" {
		nameParts = append(nameParts, d.CodenameID)
	} else if d.ReleaseID != "" {
		nameParts = append(nameParts, d.ReleaseID)
	}
	if hostname != "" {
		nameParts = append(nameParts, hostname)
	}

	distroTag := strings.Join(nameParts, "-")
	isoName := fmt.Sprintf("egg-of_%s_%s_%s.iso", distroTag, arch, timestamp)

	plan.Plan = append(plan.Plan, Action{
		Command:   "oa_remaster_iso",
		VolID:     "OA_LIVE",
		OutputISO: isoName,
	})

	plan.Plan = append(plan.Plan, Action{Command: "oa_remaster_cleanup"})

	return plan
}
