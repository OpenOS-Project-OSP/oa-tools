/*
* oa: remastering core
*
* Author: Piero Proietti <piero.proietti@gmail.com>
* License: GPL-3.0-or-later
*/
#include "oa.h"

int lay_iso(OA_Context *ctx) {
    // 1. Lookup Strategico: Locale (task) > Globale (root) [cite: 317, 321]
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");

    cJSON *volid = cJSON_GetObjectItemCaseSensitive(ctx->task, "volid");
    if (!volid) volid = cJSON_GetObjectItemCaseSensitive(ctx->root, "volid");

    cJSON *output_iso = cJSON_GetObjectItemCaseSensitive(ctx->task, "output_iso");
    if (!output_iso) output_iso = cJSON_GetObjectItemCaseSensitive(ctx->root, "output_iso");

    cJSON *bootloaders_path = cJSON_GetObjectItemCaseSensitive(ctx->task, "bootloaders_path");
    if (!bootloaders_path) bootloaders_path = cJSON_GetObjectItemCaseSensitive(ctx->root, "bootloaders_path");

    if (!cJSON_IsString(pathLiveFs)) {
        fprintf(stderr, "Error: pathLiveFs missing in ISO context\n");
        return 1;
    }

    char iso_root[PATH_SAFE], output_iso_path[PATH_SAFE];
    snprintf(iso_root, sizeof(iso_root), "%s/iso", pathLiveFs->valuestring);
    
    // Logica Smart Path: Relativo vs Assoluto [cite: 106, 582]
    const char *iso_target = cJSON_IsString(output_iso) ? output_iso->valuestring : "live-system.iso";
    if (iso_target[0] == '/') {
        strncpy(output_iso_path, iso_target, PATH_SAFE);
        printf("\033[1;34m[oa ISO]\033[0m Absolute output path detected. Writing directly to: %s\n", output_iso_path);
    } else {
        snprintf(output_iso_path, sizeof(output_iso_path), "%s/%s", pathLiveFs->valuestring, iso_target);
    }

    // 2. Risoluzione percorso isohdpfx.bin (MBR ibrido) [cite: 109, 586]
    char isohdpfx_src[PATH_SAFE];
    const char *prefix = (cJSON_IsString(bootloaders_path) && strlen(bootloaders_path->valuestring) > 0) 
                         ? bootloaders_path->valuestring 
                         : "";

    if (prefix[0] != '\0') {
        // Se c'è un prefisso (es. /tmp/coa/bootloaders), cerchiamo nella cartella ISOLINUX del tarball
        snprintf(isohdpfx_src, PATH_SAFE, "%s/ISOLINUX/isohdpfx.bin", prefix);
        printf("\033[1;34m[oa ISO]\033[0m Sourcing MBR from external prefix: %s\n", prefix);
    } else {
        // Fallback standard su Debian reale [cite: 588]
        strncpy(isohdpfx_src, "/usr/lib/ISOLINUX/isohdpfx.bin", PATH_SAFE);
    }

    char cmd[CMD_MAX];
    char uefi_args[1024] = "";
    char efi_check[PATH_SAFE];
    snprintf(efi_check, sizeof(efi_check), "%s/EFI", iso_root);

    // 3. Creazione dinamica di efi.img per UEFI hybrid boot [cite: 109, 590]
    if (access(efi_check, F_OK) == 0) {
        printf("\033[1;34m[oa ISO]\033[0m Generating FAT efi.img for UEFI hybrid boot...\n");
        char efi_img_path[PATH_SAFE];
        snprintf(efi_img_path, PATH_SAFE, "%s/boot/grub/efi.img", iso_root);

        snprintf(cmd, sizeof(cmd),
            "dd if=/dev/zero of=%s bs=1M count=4 status=none && "
            "mkfs.vfat %s >/dev/null && "
            "mkdir -p /tmp/oa_efi_mnt && "
            "mount -o loop %s /tmp/oa_efi_mnt && "
            "cp -r %s/EFI /tmp/oa_efi_mnt/ && "
            "umount /tmp/oa_efi_mnt && "
            "rmdir /tmp/oa_efi_mnt",
            efi_img_path, efi_img_path, efi_img_path, iso_root);

        if (system(cmd) == 0) {
            snprintf(uefi_args, sizeof(uefi_args), 
                     "-eltorito-alt-boot -e boot/grub/efi.img -no-emul-boot -isohybrid-gpt-basdat ");
            LOG_INFO("Successfully generated efi.img for UEFI support. [cite: 594]");
        } else {
            LOG_ERR("Failed to generate efi.img! ISO will be BIOS-only. [cite: 594]");
        }
    }

    // 4. Assemblaggio finale comando xorriso [cite: 224, 595]
    snprintf(cmd, sizeof(cmd),
             "xorriso -as mkisofs -J -joliet-long -r -l -iso-level 3 "
             "-isohybrid-mbr %s "
             "-partition_offset 16 -V '%s' "
             "-b isolinux/isolinux.bin -c isolinux/boot.cat "
             "-no-emul-boot -boot-load-size 4 -boot-info-table "
             "%s" 
             "-o %s %s/",
             isohdpfx_src,
             cJSON_IsString(volid) ? volid->valuestring : "OA_LIVE",
             uefi_args,
             output_iso_path, 
             iso_root);

    printf("\n\033[1;32m[oa ISO]\033[0m Finalizing Hybrid ISO: %s\n", output_iso_path);
            
    return system(cmd);
}
