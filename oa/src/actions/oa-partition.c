#include "oa.h"
// #include <stdio.h>
// #include <stdlib.h>

int oa_partition(OA_Context *ctx) {
    cJSON *device_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "device");
    cJSON *label_obj  = cJSON_GetObjectItemCaseSensitive(ctx->task, "label");
    cJSON *parts_obj  = cJSON_GetObjectItemCaseSensitive(ctx->task, "partitions");

    if (!cJSON_IsString(device_obj) || !cJSON_IsString(label_obj) || !cJSON_IsArray(parts_obj)) {
        LOG_ERR("oa_partition requires 'device', 'label', and a 'partitions' array.");
        return 1;
    }

    const char *device = device_obj->valuestring;
    const char *label = label_obj->valuestring;

    printf("\033[1;35m[oa]\033[0m Inizializzazione disco %s con tabella %s...\n", device, label);
    LOG_INFO("Partitioning device %s with label %s", device, label);


    // Costruiamo lo script per sfdisk in memoria
    char sfdisk_script[2048] = {0};
    snprintf(sfdisk_script, sizeof(sfdisk_script), "label: %s\n", label);

    cJSON *part;
    cJSON_ArrayForEach(part, parts_obj) {
        cJSON *size_obj = cJSON_GetObjectItemCaseSensitive(part, "size");
        cJSON *type_obj = cJSON_GetObjectItemCaseSensitive(part, "type");
        
        const char *size = cJSON_IsString(size_obj) ? size_obj->valuestring : "";
        const char *type = cJSON_IsString(type_obj) ? type_obj->valuestring : "";
        
        // Se la size è "100%", sfdisk lo capisce se non specifichiamo la dimensione (prende il resto)
        // Se invece è es. "512M", lo scriviamo.
        char part_line[128];
        if (strcmp(size, "100%") == 0) {
            snprintf(part_line, sizeof(part_line), "type=%s\n", type);
        } else {
            snprintf(part_line, sizeof(part_line), "size=%s, type=%s\n", size, type);
        }
        
        // Aggiungiamo la riga allo script
        strncat(sfdisk_script, part_line, sizeof(sfdisk_script) - strlen(sfdisk_script) - 1);
    }

    LOG_INFO("Script sfdisk generato:\n%s", sfdisk_script);

    // Prepariamo il comando di esecuzione
    char cmd[2560];
    // Usiamo echo per passare lo script a sfdisk. 
    // --wipe always cancella le firme di vecchi filesystem.
    snprintf(cmd, sizeof(cmd), "echo '%s' | sfdisk --wipe always --force %s", sfdisk_script, device);

    LOG_INFO("Esecuzione: %s", cmd);

    // SCOMMENTA QUESTA RIGA PER RENDERLO DISTRUTTIVO!
    // int res = system(cmd);
    // if (res != 0) {
    //     LOG_ERR("Partizionamento fallito su %s", device);
    //     return 1;
    // }
    
    // Piccolo delay per dare tempo al kernel di rileggere la tabella prima di formattare
    // system("udevadm settle");
   
    return 0;
}