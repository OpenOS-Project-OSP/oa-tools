/*
 * oa: eggs in my dialect🥚🥚
 * remastering core
 *
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * License: GPL-3.0-or-later
 */
#include "oa.h"

// Helper per leggere il file JSON
char *read_file(const char *filename) {
    FILE *f = fopen(filename, "rb");
    if (!f) return NULL;
    fseek(f, 0, SEEK_END);
    long len = ftell(f);
    fseek(f, 0, SEEK_SET);
    char *data = malloc(len + 1);
    if (data) {
        fread(data, 1, len, f);
        data[len] = '\0';
    }
    fclose(f);
    return data;
}

// Il "Vigile Urbano": smista i verbi ai vari moduli tramite OA_Context
int execute_verb(cJSON *root, cJSON *task) {
    cJSON *command = cJSON_GetObjectItemCaseSensitive(task, "command");
    if (!cJSON_IsString(command) || (command->valuestring == NULL))
        return 1;

    // CREAZIONE DEL CONTESTO (Puntatori a Global + Local)
    OA_Context ctx = { .root = root, .task = task };

    printf("\033[1;34m[oa]\033[0m Executing action '%s'...\n", command->valuestring);

    // Mappatura comandi: tutte le azioni ora ricevono solo il puntatore a ctx
    if (strcmp(command->valuestring, "action_prepare") == 0)  return action_prepare(&ctx);
    if (strcmp(command->valuestring, "action_users") == 0)    return action_users(&ctx);
    if (strcmp(command->valuestring, "action_initrd") == 0)   return action_initrd(&ctx);
    if (strcmp(command->valuestring, "action_remaster") == 0) return action_remaster(&ctx);
    if (strcmp(command->valuestring, "action_squash") == 0)   return action_squash(&ctx);
    if (strcmp(command->valuestring, "action_iso") == 0)      return action_iso(&ctx);
    if (strcmp(command->valuestring, "action_pause") == 0)    return action_pause(&ctx);
    if (strcmp(command->valuestring, "action_cleanup") == 0)  return action_cleanup(&ctx);
    if (strcmp(command->valuestring, "action_run") == 0)      return action_run(&ctx);
    if (strcmp(command->valuestring, "action_scan") == 0)     return action_scan(&ctx);

    fprintf(stderr, "{\"error\": \"Unknown command '%s'\"}\n", command->valuestring);
    return 1;
}

int main(int argc, char *argv[]) {
    if (argc < 2) {
        printf("oa engine v0.2\nUsage: %s <plan.json>\n", argv[0]);
        return 1;
    }

    char *json_data = read_file(argv[1]);
    if (!json_data) {
        fprintf(stderr, "Error: Could not read file %s\n", argv[1]);
        return 1;
    }

    cJSON *json = cJSON_Parse(json_data);
    if (!json) {
        fprintf(stderr, "Error: Invalid JSON format\n");
        free(json_data);
        return 1;
    }

    cJSON *plan = cJSON_GetObjectItemCaseSensitive(json, "plan");
    int final_status = 0;

    // --- LOGICA DEL PIANO DI VOLO ---
    if (cJSON_IsArray(plan)) {
        cJSON *task;
        cJSON_ArrayForEach(task, plan) {
            // Esecuzione pulitissima: passiamo il root (json) e il task corrente
            if (execute_verb(json, task) != 0) {
                fprintf(stderr, "{\"status\": \"halted\", \"error\": \"Plan failed\"}\n");
                final_status = 1;
                break;
            }
        }
    } else {
        // Fallback per comando singolo (root e task coincidono)
        final_status = execute_verb(json, json);
    }

    cJSON_Delete(json);
    free(json_data);
    return final_status;
}
