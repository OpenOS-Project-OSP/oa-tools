/*
* oa: eggs in my dialect🥚🥚
* remastering core
*
* Author: Piero Proietti <piero.proietti@gmail.com>
* License: GPL-3.0-or-later
*/
#include "oa.h"

int action_livestruct(OA_Context *ctx) {
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    if (!cJSON_IsString(pathLiveFs)) return 1;

    char iso_dir[PATH_SAFE], live_dir[PATH_SAFE];
    snprintf(iso_dir, PATH_SAFE, "%s/iso", pathLiveFs->valuestring);
    snprintf(live_dir, PATH_SAFE, "%s/iso/live", pathLiveFs->valuestring);

    // 1. Setup Struttura Base
    char cmd[CMD_MAX];
    snprintf(cmd, sizeof(cmd), "mkdir -p %s", live_dir);
    system(cmd);

    // 2. Rilevamento Kernel
    struct utsname buffer;
    if (uname(&buffer) != 0) return 1;
    char *kversion = buffer.release;

    printf("\033[1;34m[oa LIVESTRUC]\033[0m Extracting kernel %s to live directory...\n", kversion);

    // 3. Copia Kernel
    snprintf(cmd, sizeof(cmd), "cp /boot/vmlinuz-%s %s/vmlinuz", kversion, live_dir);
    if (system(cmd) != 0) system("cp -L /vmlinuz %s/vmlinuz");

    return 0;
}
