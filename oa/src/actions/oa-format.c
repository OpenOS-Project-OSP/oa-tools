#include "oa.h"
// #include <stdio.h>
// #include <stdlib.h>

int oa_format(OA_Context *ctx) {
    cJSON *actions_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "actions");

    if (!cJSON_IsArray(actions_obj)) {
        LOG_ERR("oa_format requires an 'actions' array.");
        return 1;
    }

    printf("\033[1;35m[oa]\033[0m Formattazione partizioni in corso...\n");

    cJSON *action;
    cJSON_ArrayForEach(action, actions_obj) {
        cJSON *dev_obj = cJSON_GetObjectItemCaseSensitive(action, "device");
        cJSON *fs_obj  = cJSON_GetObjectItemCaseSensitive(action, "fs");
        cJSON *lbl_obj = cJSON_GetObjectItemCaseSensitive(action, "label");

        if (!cJSON_IsString(dev_obj) || !cJSON_IsString(fs_obj)) continue;

        const char *device = dev_obj->valuestring;
        const char *fs = fs_obj->valuestring;
        const char *label = cJSON_IsString(lbl_obj) ? lbl_obj->valuestring : "";

        char cmd[512];
        if (strcmp(fs, "vfat") == 0) {
            snprintf(cmd, sizeof(cmd), "mkfs.vfat -F32 -n '%s' %s", label, device);
        } else if (strcmp(fs, "swap") == 0) {
            snprintf(cmd, sizeof(cmd), "mkswap -L '%s' %s", label, device);
        } else {
            // L'opzione -F forza la formattazione senza chiedere conferme (ext4/btrfs)
            snprintf(cmd, sizeof(cmd), "mkfs.%s -F -L '%s' %s", fs, label, device);
        }

        LOG_INFO("Esecuzione: %s", cmd);
        
        // SCOMMENTA QUESTA RIGA PER RENDERLO DISTRUTTIVO!
        // int res = system(cmd);
        // if (res != 0) {
        //     LOG_ERR("Formattazione fallita su %s", device);
        //     return 1;
        // }
    }
    return 0;
}