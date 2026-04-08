/*
* oa: eggs in my dialect🥚🥚
* remastering core
*
* Author: Piero Proietti <piero.proietti@gmail.com>
* License: GPL-3.0-or-later
*/
#include "oa.h"

/**
 * @brief Prepara la struttura della directory live ed estrae il kernel host.
 * Gestisce i diversi schemi di denominazione tra Debian, Arch e derivate.
 */
int lay_livestruct(OA_Context *ctx) {
    // 1. Lookup del percorso di lavoro (Locale > Globale) [cite: 321, 324]
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    if (!cJSON_IsString(pathLiveFs)) return 1;

    char live_dir[PATH_SAFE];
    snprintf(live_dir, PATH_SAFE, "%s/iso/live", pathLiveFs->valuestring);

    // Crea la directory di destinazione [cite: 233]
    char cmd[CMD_MAX];
    snprintf(cmd, sizeof(cmd), "mkdir -p %s", live_dir);
    system(cmd);

    // 2. Rilevamento Kernel via syscall [cite: 234]
    struct utsname buffer;
    if (uname(&buffer) != 0) return 1;
    char *kversion = buffer.release;

    printf("\033[1;34m[oa LIVESTRUC]\033[0m Extracting kernel %s to live directory...\n", kversion);

    // -------------------------------------------------------------------------
    // 3. Strategia di copia sequenziale (Fallback) 
    // -------------------------------------------------------------------------
    
    // TENTATIVO A: Nome specifico con versione (Standard Debian/Ubuntu)
    snprintf(cmd, sizeof(cmd), "cp /boot/vmlinuz-%s %s/vmlinuz", kversion, live_dir);
    if (system(cmd) == 0) {
        LOG_INFO("Kernel extracted using versioned path: /boot/vmlinuz-%s", kversion);
        return 0;
    }

    // TENTATIVO B: Nome generico Arch Linux (vmlinuz-linux)
    LOG_INFO("Versioned kernel path failed, trying Arch Linux standard...");
    snprintf(cmd, sizeof(cmd), "cp /boot/vmlinuz-linux %s/vmlinuz", live_dir);
    if (system(cmd) == 0) {
        LOG_INFO("Kernel extracted using Arch standard path: /boot/vmlinuz-linux");
        return 0;
    }

    // TENTATIVO C: Nome generico Arch Linux LTS (vmlinuz-linux-lts)
    snprintf(cmd, sizeof(cmd), "cp /boot/vmlinuz-linux-lts %s/vmlinuz", live_dir);
    if (system(cmd) == 0) {
        LOG_INFO("Kernel extracted using Arch LTS path: /boot/vmlinuz-linux-lts");
        return 0;
    }

    // TENTATIVO D: Fallback estremo tramite symlink in root (Passepartout) 
    LOG_WARN("Standard paths failed, falling back to root symlink /vmlinuz");
    snprintf(cmd, sizeof(cmd), "cp -L /vmlinuz %s/vmlinuz", live_dir);
    
    if (system(cmd) != 0) {
        LOG_ERR("Failed to extract kernel from all known locations.");
        fprintf(stderr, "\033[1;31m[oa LIVESTRUC]\033[0m Error: Kernel not found!\n");
        return 1;
    }

    return 0;
}