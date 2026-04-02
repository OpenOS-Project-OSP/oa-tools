/*
 * oa: eggs in my dialect🥚🥚
 * remastering core
 *
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * License: GPL-3.0-or-later
 */
#include "oa.h"

/**
 * @brief action_suspend: Sosta tecnica per ispezione manuale
 */
int action_suspend(OA_Context *ctx) {
    // 1. Lookup del messaggio: Locale (task) > Fallback generico
    cJSON *msg_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "message");
    const char *message = cJSON_IsString(msg_obj) ? msg_obj->valuestring : "Ispezione manuale richiesta.";

    // 2. Lookup del percorso per informazione (Locale > Globale)
    cJSON *path_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!path_obj) path_obj = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    
    const char *path = cJSON_IsString(path_obj) ? path_obj->valuestring : "/unknown";

    printf("\n\033[1;33m[oa SUSPEND]\033[0m %s\n", message);
    printf("\033[1;34m->\033[0m You can inspect the liveroot at: %s/liveroot\n", path);
    printf("\033[1;32m->\033[0m Press ENTER to resume the flight plan...");
    
    // Pulizia buffer standard output per forzare la stampa a schermo
    fflush(stdout);
    
    // RISOLUZIONE BUG "DOPPIO INVIO":
    // Visto che 'oa' non usa stdin per input interattivi precedenti, 
    // il buffer è pulito. Questo ciclo aspetta l'Invio e termina subito.
    int c;
    while ((c = getchar()) != '\n' && c != EOF); 

    return 0;
}