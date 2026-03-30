/**
 * actions_users.c
 * Core engine 'oa' - User identity management (Yocto/Eggs philosophy)
 * Author: Piero Proietti
 */

#include "oa.h"

// Prototipo per utilità di sistema (da includere o definire nel tuo header)
// extern char *pathLiveFs; 

int action_users(cJSON *plan) {
    // 1. Estrazione parametri dal JSON
    cJSON *path_node = cJSON_GetObjectItemCaseSensitive(plan, "pathLiveFs");
    cJSON *mode_node = cJSON_GetObjectItemCaseSensitive(plan, "mode");
    
    char *pathLiveFs = path_node ? path_node->valuestring : "/home/eggs";
    char *mode = (mode_node && mode_node->valuestring) ? mode_node->valuestring : "";

    printf("\033[1;34m[oa]\033[0m Action: \033[1musers\033[0m | Mode: \033[1;32m'%s'\033[0m\n", mode);

    // --- CASE 1: CLONE ---
    if (strcmp(mode, "clone") == 0) {
        printf("[oa] Clone mode: Identity preserved. Skipping changes.\n");
        return 0;
    }

    char liveroot[1024];
    snprintf(liveroot, sizeof(liveroot), "%s/liveroot", pathLiveFs);

    // --- CASE 2: CRYPTED (Backup identities before purge) ---
    if (strcmp(mode, "crypted") == 0) {
        char backup_path[1024];
        snprintf(backup_path, sizeof(backup_path), "%s/var/lib/oa/identity_backup", liveroot);
        
        printf("[oa] Crypted mode: Backing up host identities to %s\n", backup_path);
        
        char mkdir_cmd[1024];
        snprintf(mkdir_cmd, sizeof(mkdir_cmd), "mkdir -p %s", backup_path);
        system(mkdir_cmd);

        const char *files[] = {"passwd", "shadow", "group", "gshadow"};
        for (int i = 0; i < 4; i++) {
            char cp_cmd[1024];
            snprintf(cp_cmd, sizeof(cp_cmd), "cp %s/etc/%s %s/", liveroot, files[i], backup_path);
            system(cp_cmd);
        }
    }

    // --- CASE 3: PURGE & CREATE LIVE USER (For "" and "crypted") ---
    printf("[oa] Resetting identities and creating live user...\n");

    /*
     * Costruiamo un unico comando shell complesso da eseguire nel chroot.
     * 1. Elimina tutti gli utenti con UID >= 1000 (esclusi quelli di sistema e nobody)
     * 2. Crea l'utente 'live' (UID 1000)
     * 3. Configura password e privilegi sudo
     */
    char chroot_cmd[4096];
    snprintf(chroot_cmd, sizeof(chroot_cmd),
        "chroot %s /bin/sh -c \""
        // Purge utenti reali (UID 1000-59999)
        "for u in $(awk -F: '$3 >= 1000 && $3 < 60000 {print $1}' /etc/passwd); do "
        "  userdel -r $u 2>/dev/null || userdel $u 2>/dev/null; "
        "done; "
        // Creazione utente Live (Filosofia Yocto/Eggs)
        "useradd -m -u 1000 -s /bin/bash -c 'Live User' live && "
        "echo 'live:live' | chpasswd && "
        // Gruppi standard (Agnosticismo: usiamo || true per non fallire se un gruppo non esiste)
        "for g in sudo wheel audio video cdrom netdev plugdev lpadmin scanner; do "
        "  usermod -aG $g live 2>/dev/null || true; "
        "done; "
        // Passwordless sudo
        "echo 'live ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/oa-live && "
        "chmod 0440 /etc/sudoers.d/oa-live"
        "\"", liveroot);

    int status = system(chroot_cmd);

    if (status == 0) {
        printf("\033[1;32m[oa]\033[0m User 'live' created successfully.\n");
    } else {
        fprintf(stderr, "\033[1;31m[oa]\033[0m Error during user creation!\n");
    }

    return status;
}
ls
