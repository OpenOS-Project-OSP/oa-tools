/*
 * oa: eggs in my dialect🥚🥚
 * remastering core
 *
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * License: GPL-3.0-or-later
 */
#include "oa.h"

/**
 * @brief action_pause: Sosta tecnica con messaggio personalizzato dal JSON
 */
int action_pause(OA_Context *ctx) {
    // 1. Lookup del messaggio: Locale (task) > Fallback generico
    cJSON *msg_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "message");
    const char *message = cJSON_IsString(msg_obj) ? msg_obj->valuestring : "Ispezione manuale richiesta.";

    // 2. Lookup del percorso per informazione (Locale > Globale)
    cJSON *path_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!path_obj) path_obj = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    
    const char *path = cJSON_IsString(path_obj) ? path_obj->valuestring : "/unknown";

    printf("\n\033[1;33m[oa PAUSE]\033[0m %s\n", message);
    printf("\033[1;34m->\033[0m Puoi ispezionare la liveroot in: %s/liveroot\n", path);
    printf("\033[1;32m->\033[0m Premi INVIO per riprendere il piano di volo...");

    // Pulizia buffer e attesa input
    fflush(stdout);
    
    // Consuma eventuali residui nel buffer e aspetta il tasto Invio
    int c;
    while ((c = getchar()) != '\n' && c != EOF); 
    if (c != EOF) getchar(); 

    return 0;
}
