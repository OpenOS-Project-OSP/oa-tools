/*
 * src/actions/lay_iso.c
 * Versione "Zero Stress" per Artisan
 */
#include "oa.h"

int lay_iso(OA_Context *ctx) {
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");

    cJSON *volid = cJSON_GetObjectItemCaseSensitive(ctx->task, "volid");
    cJSON *outputISO = cJSON_GetObjectItemCaseSensitive(ctx->task, "output_iso");
    cJSON *bootloaders_path = cJSON_GetObjectItemCaseSensitive(ctx->root, "bootloaders_path");

    if (!cJSON_IsString(pathLiveFs) || !cJSON_IsString(outputISO)) return 1;

    const char *label = cJSON_IsString(volid) ? volid->valuestring : "OA_LIVE";
    const char *prefix = cJSON_IsString(bootloaders_path) ? bootloaders_path->valuestring : "";
    
    char mbr_path[PATH_SAFE];
    // Puntiamo all'MBR nel nostro prefisso universale
    snprintf(mbr_path, PATH_SAFE, "%s/ISOLINUX/isohdpfx.bin", prefix);

    char cmd[CMD_MAX];
    
    // USIAMO IL PERCORSO RELATIVO PIÙ PROBABILE: live/efi.img
    // Se lay_uefi lo mette altrove, xorriso ci avviserà, ma questo è lo standard eggs
    snprintf(cmd, sizeof(cmd),
             "xorriso -as mkisofs "
             "-volid \"%s\" "
             "-joliet -rock "
             "-b isolinux/isolinux.bin "
             "-c isolinux/boot.cat "
             "-no-emul-boot -boot-load-size 4 -boot-info-table "
             "-eltorito-alt-boot "
             "-e live/efi.img " 
             "-no-emul-boot "
             "-isohybrid-mbr %s "
             "-o %s "
             "%s/iso",
             label, mbr_path, outputISO->valuestring, pathLiveFs->valuestring);

    printf("\033[1;34m[oa ISO]\033[0m Finalizing ISO: %s\n", outputISO->valuestring);

    if (system(cmd) != 0) {
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m xorriso failed. Check if live/efi.img exists.\n");
        return 1;
    }

    return 0;
}